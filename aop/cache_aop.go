package aop

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"

	"github.com/darciopacifico/cachengo/cache"
	"github.com/op/go-logging"
)

var typeCacheable = reflect.TypeOf((*cache.Cacheable)(nil)).Elem()
var log = logging.MustGetLogger("cache")
var errorInterfaceModel = reflect.TypeOf((*error)(nil)).Elem()

//just to formalize the signature of swap function
type TypeSwapFunc func(ins []reflect.Value) []reflect.Value

type CacheSpot struct {
	CachedSpotFunction interface{}        //empty function ref, that will contain cacheable function. Pass a nil reference
	OriginalFunction   interface{}        //real hot function, that will have results cached
	CacheManager       cache.CacheManager //cache manager implementation
	TakeCache          bool               //mark if cache manager can take cached values or just store results

	Name          string  //(Optional) cache spot config for log and metrics.
	CacheIdPrefix *string //(Optional) cache prefix for cache registries
}

func MakeSwapPrefix(CachedSpotFunction interface{}, OriginalFunction interface{}, CacheManager cache.CacheManager, TakeCache bool, Prefix string) {
	MakeCachedSpotFunction(CacheSpot{
		CachedSpotFunction: CachedSpotFunction,
		OriginalFunction:   OriginalFunction,
		CacheManager:       CacheManager,
		TakeCache:          TakeCache,
		CacheIdPrefix:      &Prefix,
	})
}

func MakeSwap(CachedSpotFunction interface{}, OriginalFunction interface{}, CacheManager cache.CacheManager, TakeCache bool) {
	MakeCachedSpotFunction(CacheSpot{
		CachedSpotFunction: CachedSpotFunction,
		OriginalFunction:   OriginalFunction,
		CacheManager:       CacheManager,
		TakeCache:          TakeCache,
	})
}

//makes a swap function for reflection operations
func MakeCachedSpotFunction(cacheSpot CacheSpot) {

	defer func() { //assure for not panicking
		if r := recover(); r != nil {
			log.Error("Error trying to make a swap function! %v", r)
			os.Exit(1)
		}
	}()

	//assure for swap possibilities
	mustBePossibleToSwap(cacheSpot.CachedSpotFunction, cacheSpot.OriginalFunction)

	eI, eO := getInOutTypes(reflect.TypeOf(cacheSpot.CachedSpotFunction))
	cI, _ := getInOutTypes(reflect.TypeOf(cacheSpot.OriginalFunction))

	eIM, cIM := isMany(eI[0]), isMany(cI[0])

	defaultVals := defaultValues(cacheSpot.CachedSpotFunction, eO, true)

	swap := func(originalIns []reflect.Value) []reflect.Value {

		var retVals []reflect.Value

		//make a fit swap function
		if !eIM && !cIM { //one to one
			retVals = executeOneToOne(cacheSpot, originalIns, eO, defaultVals)

		} else if eIM && cIM { // many to many
			retVals = executeManyToMany(cacheSpot, originalIns, eO, eI, defaultVals)

		} else if eIM && !cIM { //many to one
			retVals = executeManyToOne(cacheSpot, originalIns, eO, eI, defaultVals)

		} else if !eIM && cIM { //one to many
			retVals = executeOneToMany(cacheSpot, originalIns, eO, eI, cI, defaultVals)

		} else {
			log.Critical("I is not logically supposed to be possible to reach this code! Something really wrong happened!")
			retVals = defaultVals
		}

		//fix return value type: there is no polimorphism in go
		return fixReturnTypes(eO, retVals)

	}

	//set emptyBodyFunction body with swapFunction containing cache mechanism
	setSwapAsFunctionBody(cacheSpot.CachedSpotFunction, swap)
}

//check if is array
func isMany(vType reflect.Type) bool {
	return vType.Kind() == reflect.Array || vType.Kind() == reflect.Slice
}

//check if some value is an array
func isValMany(value reflect.Value) bool {
	return isMany(value.Type())
}

//return the the out types for a function
func getInOutTypes(someFunction reflect.Type) ([]reflect.Type, []reflect.Type) {
	if someFunction.Kind() == reflect.Ptr {
		someFunction = someFunction.Elem()
	}

	//recover the return types
	numOut := someFunction.NumOut()
	outTypes := make([]reflect.Type, numOut)
	for i := 0; i < numOut; i++ {
		outTypes[i] = someFunction.Out(i)
	}

	//recover the input types
	numIn := someFunction.NumIn()
	inTypes := make([]reflect.Type, numIn)
	for i := 0; i < numIn; i++ {
		inTypes[i] = someFunction.In(i)
	}
	return inTypes, outTypes
}

//split array results in another two arrays: found and not found
func splitFoundNotFound(ins []reflect.Value, outType reflect.Type, cacheRegs map[string]cache.CacheRegistry, prefix *string) ([]reflect.Value, []reflect.Value) {

	nfKeys := []string{}
	fKeys := []string{}
	nfIns := []reflect.Value{}
	fOuts := []reflect.Value{}

	valEntrada := ins[0]

	for i := 0; i < valEntrada.Len(); i++ {

		in := valEntrada.Index(i)

		key := getKeyForInput(outType, in, prefix)

		cacheReg, hasMap := cacheRegs[key]

		if !hasMap || !cacheReg.HasValue { //not found
			nfIns = append(nfIns, in)    //will be searched using hot concrete function
			nfKeys = append(nfKeys, key) //
		} else {
			fVal := reflect.Indirect(reflect.ValueOf(cacheReg.Payload))

			fOuts = append(fOuts, fVal)
			fKeys = append(fKeys, key)
		}
	}

	return fOuts, nfIns
}

// search for cached values
func getCachedMap(emptyBodyFunction interface{}, in reflect.Value, outType reflect.Type, cacheManager cache.CacheManager, prefix *string) map[string]cache.CacheRegistry {

	//keys array, based on inputs and return types
	keys, errCK := cacheKeysDyn(emptyBodyFunction, in, outType, prefix)
	if errCK != nil {
		log.Error("Error trying to solve cache keys! Is not possible to proceed with cache operations!", errCK)
		emptyMap := make(map[string]cache.CacheRegistry, 0)
		panic(errCK) // fckp
		return emptyMap
	}

	cacheRegMap, err := cacheManager.GetCaches(keys...)
	if err != nil {
		log.Error("Error trying to retrieve cache data x", errCK)
		emptyMap := make(map[string]cache.CacheRegistry, 0)
		panic(err) // fckp
		return emptyMap
	}

	return cacheRegMap
}

//execute an one to one reflection + cache operation
func executeOneToOne(cacheSpot CacheSpot, originalIns []reflect.Value, eO []reflect.Type, defaultVals []reflect.Value) (returnValue []reflect.Value) {

	emptyFunction := cacheSpot.CachedSpotFunction
	concreteFunction := cacheSpot.OriginalFunction
	cacheManager := cacheSpot.CacheManager
	prefix := cacheSpot.CacheIdPrefix

	defer func() { //assure for not panicking
		if r := recover(); r != nil {
			log.Error("Recovering! Error trying to call a swap function!! y %v", r)
			log.Error("Falling back this request to direct hot function call, without cache!")

			returnValue = reflect.ValueOf(concreteFunction).Call(originalIns)
			return
		}
	}()

	cacheRegMap := getCachedMap(emptyFunction, originalIns[0], eO[0], cacheManager, prefix)
	strKey := getKeyForInput(eO[0], originalIns[0], prefix)
	cachedVal, hasCacheVal := cacheRegMap[strKey]

	if hasCacheVal {
		return putFirstResultEvidence(reflect.ValueOf(cachedVal.Payload), defaultVals)

	} else {
		//hot call
		hotOuts := reflect.ValueOf(concreteFunction).Call(originalIns)

		//store in cache
		go func() {
			defer func() { //assure for not panicking
				if r := recover(); r != nil {
					log.Error("Recovering! Error trying to save cache registry y! %v", r)
					return
				}
			}()

			// check whether results are valid and must be cached
			if validateResults(emptyFunction, originalIns, hotOuts) {
				singleStoreInCache(hotOuts[0], strKey, cacheManager)
			}
		}()

		return hotOuts
	}
}

//execute an many to many call
func executeManyToMany(cacheSpot CacheSpot, originalIns []reflect.Value, eO []reflect.Type, eI []reflect.Type, defaultVals []reflect.Value) (returnVal []reflect.Value) {

	emptyFunction := cacheSpot.CachedSpotFunction
	concreteFunction := cacheSpot.OriginalFunction
	cacheManager := cacheSpot.CacheManager
	prefix := cacheSpot.CacheIdPrefix

	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovering! Error trying to call a swap function!! %v", r)
			log.Error("Falling back this request to direct hot function call, without cache!")

			returnVal = reflect.ValueOf(concreteFunction).Call(originalIns)

			return
		}
	}()

	return executeManyToAny(emptyFunction, originalIns, eO, eI, cacheManager, defaultVals, concreteFunction, prefix)
}

//fix return type acoordingly to out type
func fixReturnTypes(outTypes []reflect.Type, values []reflect.Value) []reflect.Value {

	if len(values) == 0 || len(outTypes) == 0 {
		return values
	}

	if values[0].Type().AssignableTo(outTypes[0]) &&
		values[0].Type().ConvertibleTo(outTypes[0]) {

		newVal := reflect.New(outTypes[0])
		newVal.Elem().Set(values[0])
		values[0] = newVal.Elem()
	}

	return values
}

//execute an many to one call
func executeManyToOne(cacheSpot CacheSpot, originalIns []reflect.Value, eO []reflect.Type, eI []reflect.Type, defaultVals []reflect.Value) (returnValue []reflect.Value) {
	emptyFunction := cacheSpot.CachedSpotFunction
	concreteFunction := cacheSpot.OriginalFunction
	cacheManager := cacheSpot.CacheManager
	prefix := cacheSpot.CacheIdPrefix

	defer func() { //assure for not panicking
		if r := recover(); r != nil {

			log.Error("Recovering! Error trying to call a swap function!! y %v", r)
			log.Error("Falling back this request to direct hot function call, without cache!")

			hotReturnedValues := CallHotFunction(originalIns, FromWrappedToArray(originalIns[0]), concreteFunction, emptyFunction, eI[0])
			returnValue = putFirstArrResultEvidence(eO, hotReturnedValues, defaultVals)
			return
		}
	}()

	return executeManyToAny(emptyFunction, originalIns, eO, eI, cacheManager, defaultVals, concreteFunction, prefix)
}

//call from Many function interface to any kind of concrete funcition (many or one)
func executeManyToAny(emptyFunction interface{}, originalIns []reflect.Value, eO []reflect.Type, eI []reflect.Type,
	cacheManager cache.CacheManager, defaultVals []reflect.Value, concreteFunction interface{}, prefix *string) (returnValue []reflect.Value) {

	outType := getArrayInnerType(eO[0])
	cacheRegMap := getCachedMap(emptyFunction, originalIns[0], eO[0], cacheManager, prefix)
	cachedOuts, notCachedIns := splitFoundNotFound(originalIns, outType, cacheRegMap, prefix)
	hotReturnedValues := CallHotFunction(originalIns, notCachedIns, concreteFunction, emptyFunction, eI[0])

	if len(hotReturnedValues) > 0 {
		go newStoreInCache(hotReturnedValues, emptyFunction, cacheManager)
	}

	joinedReturn := append(cachedOuts, hotReturnedValues...)

	return putFirstArrResultEvidence(eO, joinedReturn, defaultVals)
}

//Take the substitute for first value and join with default values for other results
func putFirstArrResultEvidence(eO []reflect.Type, hotResult []reflect.Value, defaultVals []reflect.Value) []reflect.Value {
	firstResult := FromArrayToWrapped(hotResult, eO[0]) // set first value as joined return
	return putFirstResultEvidence(firstResult, defaultVals)
}

//Take the substitute for first value and join with default values for other results
func putFirstResultEvidence(firstResult reflect.Value, defaultVals []reflect.Value) []reflect.Value {
	arrOuts := make([]reflect.Value, len(defaultVals))
	arrOuts[0] = firstResult
	for index := 1; index < len(defaultVals); index++ {
		arrOuts[index] = defaultVals[index] // set other returns as default
	}
	return arrOuts
}

func executeOneToMany(cacheSpot CacheSpot, originalIns []reflect.Value, eO []reflect.Type, eI []reflect.Type, cI []reflect.Type, defaultVals []reflect.Value) (returnValue []reflect.Value) {

	emptyFunction := cacheSpot.CachedSpotFunction
	concreteFunction := cacheSpot.OriginalFunction
	cacheManager := cacheSpot.CacheManager
	prefix := cacheSpot.CacheIdPrefix

	defer func() { //assure for not panicking out
		if r := recover(); r != nil {

			log.Error("Recovering! Error trying to call a swap function!! y %v", r)
			log.Error("Falling back this request to direct hot function call, without cache!")

			fakeIns := convertOneCallToManyCall(originalIns, eI, cI)
			manyOuts := reflect.ValueOf(concreteFunction).Call(fakeIns)
			oneReturn, _ := convertManyReturnToOneReturn(manyOuts[0], eO)

			returnValue = putFirstResultEvidence(oneReturn, defaultVals) //TODO: DEFINIR COMO DETERMINAR SUCESSO OU FALHA

			return
		}
	}()

	cacheRegMap := getCachedMap(emptyFunction, originalIns[0], eO[0], cacheManager, prefix)

	strKey := getKeyForInput(eO[0], originalIns[0], prefix)
	cachedVal, hasCacheVal := cacheRegMap[strKey]

	var valToReturn reflect.Value

	var returnBool bool

	if hasCacheVal {
		valToReturn = reflect.Indirect(reflect.ValueOf(cachedVal.Payload))
		returnBool = true
	} else {

		fakeIns := convertOneCallToManyCall(originalIns, eI, cI)
		manyOuts := reflect.ValueOf(concreteFunction).Call(fakeIns)
		oneOut, hasReturn := convertManyReturnToOneReturn(manyOuts[0], eO)

		if hasReturn { // returned array is greater that 0
			go func() {
				defer func() { //assure for not panicking
					if r := recover(); r != nil {
						log.Error("Recovering! Error trying to save cache registry y! %v", r)
					}
				}()
				// check whether results are valid and must be cached
				if validateResults(emptyFunction, originalIns, []reflect.Value{oneOut}) {
					singleStoreInCache(oneOut, strKey, cacheManager)
				}
			}()
		}

		valToReturn = oneOut
		returnBool = hasReturn
	}

	newDefVal := defaultValues(emptyFunction, eO, returnBool)

	arrOuts := putFirstResultEvidence(valToReturn, newDefVal)

	return arrOuts
}

//store results in cache
func newStoreInCache(origOuts []reflect.Value, emptyFunction interface{}, cacheManager cache.CacheManager) error {
	keys, outs := getKeysForOuts(origOuts, emptyFunction)

	return cacheValues(outs, keys, cacheManager)

}

//store results in cache
func cacheValues(outs []reflect.Value, keys []string, cacheManager cache.CacheManager) error {
	numOut := len(outs)

	cacheRegs := make([]cache.CacheRegistry, numOut)

	//iterate over all function returns. All of then can be stored
	for index := 0; index < numOut; index++ {

		//index := 0 // hard coded index. refactor to use any quantity of return valures ASAP
		//setting cache
		cacheId := keys[index]

		//a empty cachekey means that this <value will not be stored
		if len(cacheId) > 0 {

			//get raw value
			valRet := outs[index].Interface()

			ttl := discoverTTL(valRet, -1)

			log.Debug("TTL for reg %v %v!", cacheId, ttl)

			//invoke cache manager to persist returned value
			cacheRegs[index] = cache.CacheRegistry{CacheKey: cacheId, Payload: valRet, Ttl: ttl, HasValue: true}

		}
	}

	log.Debug("saving registries %s!", keys)
	err := cacheManager.SetCache(cacheRegs...)
	if err != nil {
		log.Error("Erro trying to save cache keys %v, error %v!", keys, err)
		return err
	}

	return nil
}

//store results in cache
func singleStoreInCache(hotOut reflect.Value, cacheId string, cacheManager cache.CacheManager) {
	//TODO REUSE THIS FUNCTION AT FORMER StoreInCache FUNCTION
	//a empty cachekey means that this <value will not be stored
	if len(cacheId) > 0 {

		log.Debug("saving registry %s!", cacheId)
		//get raw value
		valRet := hotOut.Interface()

		ttl := discoverTTL(valRet, -1)

		log.Debug("TTL for reg %v %v!", cacheId, ttl)

		//invoke cache manager to persist returned value
		cacheRegistry := cache.CacheRegistry{CacheKey: cacheId, Payload: valRet, Ttl: ttl, HasValue: true}
		cacheManager.SetCache(cacheRegistry)
		log.Debug("registry %s saved successfully!", cacheId)
	}
}

//Retrieve ttl value, if interfaca implements cache.ExposeTTL
func discoverTTL(valRet interface{}, defaultTTL int) int {

	//switch valRel.(type)
	exposeTTL, isExposeTTL := valRet.(cache.ExposeTTL)

	ttl := defaultTTL

	if isExposeTTL {
		ttl = exposeTTL.GetTtl()
	}
	return ttl
}

//Determine cache keys, based on function parameters (in array) and outTypes
func cacheKeysDyn(emptyBodyFunction interface{}, in reflect.Value, outType reflect.Type, prefix *string) ([]string, error) {
	if isMany(in.Type()) {
		qtdIns := in.Len() // how many ids was requested
		keys := make([]string, qtdIns)
		outType := getArrayInnerType(outType) // get the return prefix

		for i := 0; i < qtdIns; i++ {
			keys[i] = getKeyForInput(outType, in.Index(i), prefix)
		}

		return keys, nil

	} else {
		key := getKeyForInput(outType, in, prefix)
		return []string{key}, nil

	}
}

//recursivelly iterate over a type until find a non array type
func getArrayInnerType(arrType reflect.Type) reflect.Type {
	if isMany(arrType) {
		return getArrayInnerType(arrType.Elem())
	} else {
		return arrType
	}
}

//retur a equivalent cache key for a input parameter
func getKeyForInput(outType reflect.Type, valueIn reflect.Value, prefix *string) string {
	//if is not possible to turn the first paramate to string, fail! Cache wll be missed!!
	strVal, err := valIntToString(valueIn)
	if err != nil {
		log.Error(" ERROR TRYING TO PARSE A CACHE KEY FOR %v. %v %v !", err, valueIn, outType)
		panic(err)
	}

	var keyPrefix string
	if prefix != nil {
		keyPrefix = *prefix
	} else {
		keyPrefix = outType.Name()
	}

	key := keyPrefix + ":" + strVal

	return key
}

// Try to convert a int value to string. if is not a integer raise error
func valIntToString(value reflect.Value) (string, error) {

	var strVal string

	switch value.Type().Kind() {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		strVal = strconv.Itoa(int(value.Int()))
	default:
		log.Error("Error trying to convert value to string!", value)
		return "", errors.New("Error trying to convert value to string!" + value.String())
	}

	return strVal, nil
}

//analyze and define if some result is valid. Usually used before a cache operation
func validateResults(emptyBodyFunction interface{}, in []reflect.Value, out []reflect.Value) bool {

	//try to convert a function in a ValidateResults interface
	functionValidator, hasValidatorImpl := emptyBodyFunction.(ValidateResults)

	//if function has a function with validation behaviour
	if hasValidatorImpl {
		//custom validation
		return functionValidator.IsValidResults(in, out)

	} else {

		//has some return value
		if len(out) > 1 &&
			out[1].IsValid() &&
			out[1].Kind() == reflect.Bool {

			boolVal, _ := out[1].Interface().(bool)

			return boolVal
		}

		log.Error("Erro ", emptyBodyFunction)
		funcName := reflect.TypeOf(emptyBodyFunction).Name()
		log.Error("", errors.New("Its not possible to infer a return value validation. Your function "+funcName+" need to implement ValidateResults inferface!"))
		return false
	}

}

//return value validation
func mustHaveKeyDefiner(emptyBodyFunction interface{}) {

	if reflect.ValueOf(emptyBodyFunction).Kind() != reflect.Ptr {
		log.Error("emptyBodyFunction needs to be a pointer!")
		panic(errors.New("emptyBodyFunction needs to be a pointer!"))
	}

	functionType := reflect.TypeOf(emptyBodyFunction)
	funcName := functionType.Elem().Name()

	_, hasOutKeyDefiner := emptyBodyFunction.(SpecifyOutKeys)

	if hasOutKeyDefiner {
		log.Debug("Function %v implements SpecifyOutKeys! Its OK!", funcName)
	} else {

		_, outTypes := getInOutTypes(reflect.TypeOf(emptyBodyFunction))
		firstType := outTypes[0] // get the first return type and check...

		if isMany(firstType) {

			innerType := getArrayInnerType(firstType)

			//... is a Cacheable implementation??
			// Cacheable is capable to define its own cache key
			if innerType.Implements(typeCacheable) {
				log.Debug("Function %s dont implements SpecifyOutKeys, but return type %v implements Cacheable! its ok!", funcName, innerType.Name())
			} else {
				panic(errors.New(fmt.Sprintf("Function %s doesn't implements SpecifyOutKeys and return type %v doesn't implements Cacheable!", funcName, innerType.Name())))
			}
		}

	}
}

//return value validation
func mustHaveValidationMethod(emptyBodyFunction interface{}) {
	if reflect.ValueOf(emptyBodyFunction).Kind() != reflect.Ptr {
		log.Error("emptyBodyFunction needs to be a pointer!")
		panic(errors.New("emptyBodyFunction needs to be a pointer!"))
	}

	functionType := reflect.TypeOf(emptyBodyFunction)
	numOut := functionType.Elem().NumOut()
	funcName := functionType.Elem().Name()

	//try to convert a function in a ValidateResults interface
	_, hasValidatorImpl := emptyBodyFunction.(ValidateResults)

	//if function has a function with validation behaviour
	if !hasValidatorImpl {
		log.Debug("Function %s dont implements ValidateResults", funcName)
		if numOut > 1 && functionType.Elem().Out(1).Kind() == reflect.Bool { //for more than one outs, the second one must be a boolean
			log.Debug("Function is a self validator!")

		} else {
			panic(errors.New("Its not possible to infer a return value validation. Your function '" + funcName + "' needs to implement ValidateResults inferface!"))
		}
	}
}

//recover default values for an function call
//non cached values (string empty cache key), will be returned as a default values, by this function or by implementing DefaultValubleFunction interface
func defaultValues(emptyFunction interface{}, outTypes []reflect.Type, defBool bool) []reflect.Value {

	defValuable, isDefValuable := emptyFunction.(DefaultValubleFunction)

	if isDefValuable {
		//there is a specific default values function
		defaultVals := defValuable.DefaultValues(outTypes)
		return defaultVals
	}

	defaultValues := make([]reflect.Value, len(outTypes))

	for i, outType := range outTypes {

		switch outType.Kind() {

		case reflect.Struct:
			defaultValues[i] = reflect.New(outType).Elem()

		case reflect.Bool:
			defaultValues[i] = reflect.ValueOf(defBool)

		case reflect.Interface:
			if outType.Implements(errorInterfaceModel) {
				var err error = nil
				defaultValues[i] = reflect.ValueOf(&err).Elem()

			} else if i > 0 { // an interface at 0 position can be ignored

				log.Error("(not error) It's not possible to identify an default value for function return index %v!", i)
				panic(errors.New("It's not possible to identify an default value for function!"))
			}

		case reflect.Slice, reflect.Array:
			defaultValues[i] = reflect.MakeSlice(outType, 0, 0)

		default:
			log.Error("(default) Its not possible to identify an default value for function  return index %v!", i)
			panic(errors.New("It's not possible to identify an default value for function!"))
		}
	}

	return defaultValues
}

//take the value of emptyBodyFunction and sets swap function as function body implementation
func setSwapAsFunctionBody(emptyBodyFunction interface{}, swap func([]reflect.Value) []reflect.Value) {

	//recover the value for function in the pointer
	fn := reflect.ValueOf(emptyBodyFunction).Elem()

	//put a recently created swap function as a function body for the emptyBodyFunction
	fn.Set(reflect.MakeFunc(fn.Type(), swap))
}
