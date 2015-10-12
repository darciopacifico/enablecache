package aop

import (
	"errors"
	"fmt"
	"reflect"
	"github.com/darciopacifico/cachengo/cache"
	"github.com/op/go-logging"
	"os"
)

var EMPTY_MAP = make(map[string]cache.CacheRegistry)
var typeCacheable = reflect.TypeOf((*cache.Cacheable)(nil)).Elem()
var log = logging.MustGetLogger("cache")
var errorInterfaceModel = reflect.TypeOf((*error)(nil)).Elem()


//template to make a cache spot function
type CacheSpot struct {
	CachedFunc    interface{}        //(required) empty function ref, that will contain cacheable function. Pass a nil reference
	HotFunc       interface{}        //(required) real hot function, that will have results cached
	CacheManager  cache.CacheManager //(required) cache manager for cache swaped function
	StoreOnly     bool               //(Optional) mark if cache manager can take cached values or just store results
	CacheIdPrefix *string            //(Optional) cache prefix for cache registries
	callContext                      // will be mounted at start up nothing to do
}

//reflect objects need to reflect function call
type callContext struct {
	spotOutType    []reflect.Type
	spotOutInnType []reflect.Type
	spotInType     []reflect.Type
	spotInInnType  []reflect.Type
	realInType     []reflect.Type
	realInInnType  []reflect.Type
	defaultVals    []reflect.Value
	cachedFuncName string // cached func name
	hotFunctName string // hot func name
	isManyOuts     bool   //compose cardinality of swap call
	isManyIns      bool   //compose cardinality of swap call

}

//makes a swap function for reflection operations
func (cacheSpot CacheSpot) StartCache() {
	//gracefull exit at panic
	defer func() { //assure for not panicking
		if r := recover(); r != nil {
			log.Error("Error trying to make a swap function! %v", r)
			os.Exit(1)
		}
	}()

	if cacheSpot.CacheManager == nil || !cacheSpot.CacheManager.Validade() {
		panic(errors.New(fmt.Sprintf("You must provide a ready and valid cache manager!")))
	}

	if cacheSpot.CachedFunc == nil {
		panic(errors.New(fmt.Sprintf("CachedFunction attribute not provided! Please set a pointer to a function type, compatible with OriginalFunction, to receive cacheable function!")))
	}

	if cacheSpot.HotFunc == nil {
		panic(errors.New(fmt.Sprintf("HotFunction attribute not provided! Please set function to be cached!")))
	}

	//analyse functions and check possibilities. Fill reflect stuffs, like in and out arr types. panic if is not possible!
	cacheSpot.parseFunctionsForSwap()
}


//Cache spot swap function. Used as a swap function in reflect calls
//Four swap calls combinations are possible: One-one, Many-Many, Many-One, One-Many.
func (c CacheSpot) swapFunction(inputParams []reflect.Value) []reflect.Value {
	var retVals []reflect.Value

	//make a fit swap function
	if !c.isManyOuts && !c.isManyIns { //one to one
		retVals = c.callOneToOne(inputParams)

	} else if c.isManyOuts && c.isManyIns { // many to many
		retVals = c.callManyToMany(inputParams)

	} else if c.isManyOuts && !c.isManyIns { //many to one
		retVals = c.callManyToOne(inputParams)

	} else if !c.isManyOuts && c.isManyIns { //one to many
		retVals = c.callOneToMany(inputParams)

	} else {
		log.Critical("I is not logically supposed to be possible to reach this code! Something really wrong happened!")
		retVals = c.defaultVals
	}

	//fix return value type. Emulates a polimorphic behaviour not present in golang
	// ex: an struct Customer{} returned as an interface{} wouldn't work,
	// after be recovered by the cache mechanism and used in an reflect return operation
	return c.fixReturnTypes(retVals)
}

//analize functions and fill reflect objects as need
func (cacheSpot CacheSpot) parseFunctionsForSwap() {

	//take inputs and output types
	SpotInType, SpotOutType := getInOutTypes(reflect.TypeOf(cacheSpot.CachedFunc))
	RealInType, _ := getInOutTypes(reflect.TypeOf(cacheSpot.HotFunc))

	//take array element type,
	// ex: []string => string
	// []Products => Product
	SpotInnOutType := getArrayInnerTypes(SpotOutType)
	SpotInnInType := getArrayInnerTypes(SpotInType)
	RealInnInType := getArrayInnerTypes(RealInType)

	//Assembly an callContext object
	cacheSpot.callContext = callContext{
		spotOutType:    SpotOutType,
		spotInType:     SpotInType,
		realInType:     RealInType,
		spotOutInnType: SpotInnOutType,
		spotInInnType:  SpotInnInType,
		realInInnType:  RealInnInType,
	}

	//adjust default values
	cacheSpot.defaultVals = cacheSpot.defaultValues(true)

	//adjust cardinality
	cacheSpot.isManyOuts = isMany(cacheSpot.spotInType[0])
	cacheSpot.isManyIns = isMany(cacheSpot.realInType[0])

	//adjust func names, they differ because one is a type and other a real function
	cacheSpot.cachedFuncName = reflect.TypeOf(cacheSpot.CachedFunc).Elem().String()
	cacheSpot.hotFunctName = GetFunctionName(cacheSpot.HotFunc)

	//assure for swap possibilities. panic at startup if is not possible!
	cacheSpot.mustBePossibleToSwap()

	//set emptyBodyFunction body with swapFunction containing cache mechanism
	cacheSpot.setSwapAsFunctionBody()
}

//split array results in another two arrays: found and not founded values
func (cacheSpot CacheSpot) splitFoundNotFound(ins []reflect.Value, cacheRegs map[string]cache.CacheRegistry) ([]reflect.Value, []reflect.Value) {

	nfKeys := []string{}
	fKeys := []string{}
	nfIns := []reflect.Value{}
	fOuts := []reflect.Value{}

	valEntrada := ins[0]

	for i := 0; i < valEntrada.Len(); i++ {

		in := valEntrada.Index(i)

		key := cacheSpot.getKeyForInput(in)

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
func (cacheSpot CacheSpot) getCachedMap(in reflect.Value) map[string]cache.CacheRegistry {

	if cacheSpot.StoreOnly {
		return EMPTY_MAP
	}

	//keys array, based on inputs and return types
	keys, errCK := cacheSpot.cacheKeysDyn(in)
	if errCK != nil {
		log.Error("Error trying to solve cache keys! Is not possible to proceed with cache operations!", errCK)
		emptyMap := make(map[string]cache.CacheRegistry, 0)
		panic(errCK) // fckp
		return emptyMap
	}

	cacheRegMap, err := cacheSpot.CacheManager.GetCaches(keys...)
	if err != nil {
		log.Error("Error trying to retrieve cache data x", errCK)
		emptyMap := make(map[string]cache.CacheRegistry, 0)
		panic(err) // fckp
		return emptyMap
	}

	return cacheRegMap
}

//execute an one to one reflection + cache operation
func (cacheSpot CacheSpot) callOneToOne(originalIns []reflect.Value) (returnValue []reflect.Value) {
	defer func() { //assure for not panicking
		if r := recover(); r != nil {
			log.Error("Recovering! Error trying to call a swap function!! y %v", r)
			log.Error("Falling back this request to direct hot function call, without cache!")

			returnValue = reflect.ValueOf(cacheSpot.HotFunc).Call(originalIns)
			return
		}
	}()

	cacheRegMap := cacheSpot.getCachedMap(originalIns[0])
	cacheKey := cacheSpot.getKeyForInput(originalIns[0])
	cachedVal, hasCacheVal := cacheRegMap[cacheKey]

	if hasCacheVal {
		return cacheSpot.putFirstResultEvidence(reflect.ValueOf(cachedVal.Payload), true)

	} else {
		//hot call
		hotOuts := reflect.ValueOf(cacheSpot.HotFunc).Call(originalIns)

		//store in cache
		go func() {
			defer func() { //assure for not panicking
				if r := recover(); r != nil {
					log.Error("Recovering! Error trying to save cache registry y! %v", r)
					return
				}
			}()

			// check whether results are valid and must be cached
			if cacheSpot.validateResults(originalIns, hotOuts) {
				cacheSpot.singleStoreInCache(hotOuts[0], cacheKey)
			}
		}()

		return hotOuts
	}
}

//execute an many to many call
func (cacheSpot CacheSpot) callManyToMany(originalIns []reflect.Value) (returnVal []reflect.Value) {

	concreteFunction := cacheSpot.HotFunc

	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovering! Error trying to call a swap function!! %v", r)
			log.Error("Falling back this request to direct hot function call, without cache!")

			returnVal = reflect.ValueOf(concreteFunction).Call(originalIns)

			return
		}
	}()

	return cacheSpot.callManyToAny(originalIns)
}

//fix return type acoordingly to out type
func (c CacheSpot) fixReturnTypes(values []reflect.Value) []reflect.Value {

	if len(values) == 0 || len(c.spotOutType) == 0 {
		return values
	}

	if values[0].Type().AssignableTo(c.spotOutType[0]) &&
	values[0].Type().ConvertibleTo(c.spotOutType[0]) {

		newVal := reflect.New(c.spotOutType[0])
		newVal.Elem().Set(values[0])
		values[0] = newVal.Elem()
	}

	return values
}

//execute an many to one call
func (cacheSpot CacheSpot) callManyToOne(originalIns []reflect.Value) (returnValue []reflect.Value) {

	defer func() { //assure for not panicking
		if r := recover(); r != nil {
			log.Error("Recovering! Error trying to call a swap function!! y %v", r)
			log.Error("Falling back this request to direct hot function call, without cache!")

			hotReturnedValues := cacheSpot.callHotFunction(originalIns, fromWrappedToArray(originalIns[0]))
			returnValue = cacheSpot.putFirstArrResultEvidence(hotReturnedValues)
			return
		}
	}()

	return cacheSpot.callManyToAny(originalIns)
}

//call from Many function interface to any kind of concrete funcition (many or one)
func (cacheSpot CacheSpot) callManyToAny(originalIns []reflect.Value) (returnValue []reflect.Value) {

	cacheRegMap := cacheSpot.getCachedMap(originalIns[0])

	cachedOuts, notCachedIns := cacheSpot.splitFoundNotFound(originalIns, cacheRegMap)
	hotReturnedValues := cacheSpot.callHotFunction(originalIns, notCachedIns)

	if len(hotReturnedValues) > 0 {
		go cacheSpot.storeInCache(hotReturnedValues)
	}

	joinedReturn := append(cachedOuts, hotReturnedValues...)

	return cacheSpot.putFirstArrResultEvidence(joinedReturn)
}

//Take the substitute for first value and join with default values for other results
func (cacheSpot CacheSpot) putFirstArrResultEvidence(hotResult []reflect.Value) []reflect.Value {

	firstResult := fromArrayToWrapped(hotResult, cacheSpot.spotOutType[0]) // set first value as joined return

	return cacheSpot.putFirstResultEvidence(firstResult, true)
}

//Take the substitute for first value and join with default values for other results
func (cacheSpot CacheSpot) putFirstResultEvidence(firstResult reflect.Value, defBool bool) []reflect.Value {
	lenDefVals := len(cacheSpot.defaultVals)

	//put first result in evidence
	arrOuts := make([]reflect.Value, lenDefVals)
	arrOuts[0] = firstResult

	//set default values to others
	for index := 1; index < lenDefVals; index++ {
		arrOuts[index] = cacheSpot.defaultVals[index] // set other returns as default
	}

	return arrOuts
}

func (cacheSpot CacheSpot) callOneToMany(originalIns []reflect.Value) (returnValue []reflect.Value) {

	defer func() { //assure for not panicking out
		if r := recover(); r != nil {

			log.Error("Recovering! Error trying to call a swap function!! y %v", r)
			log.Error("Falling back this request to direct hot function call, without cache!")

			fakeIns := cacheSpot.convertOneCallToManyCall(originalIns)
			manyOuts := reflect.ValueOf(cacheSpot.HotFunc).Call(fakeIns)
			oneReturn, _ := cacheSpot.convertManyReturnToOneReturn(manyOuts[0])

			returnValue = cacheSpot.putFirstResultEvidence(oneReturn, true) //TODO: DEFINIR COMO DETERMINAR SUCESSO OU FALHA

			return
		}
	}()

	cacheRegMap := cacheSpot.getCachedMap(originalIns[0])

	strKey := cacheSpot.getKeyForInput(originalIns[0])
	cachedVal, hasCacheVal := cacheRegMap[strKey]

	var valToReturn reflect.Value

	var returnBool bool

	if hasCacheVal {
		valToReturn = reflect.Indirect(reflect.ValueOf(cachedVal.Payload))
		returnBool = true
	} else {

		fakeIns := cacheSpot.convertOneCallToManyCall(originalIns)
		manyOuts := reflect.ValueOf(cacheSpot.HotFunc).Call(fakeIns)
		oneOut, hasReturn := cacheSpot.convertManyReturnToOneReturn(manyOuts[0])

		if hasReturn { // returned array is greater that 0
			go func() {
				defer func() { //assure for not panicking
					if r := recover(); r != nil {
						log.Error("Recovering! Error trying to save cache registry y! %v", r)
					}
				}()
				// check whether results are valid and must be cached
				if cacheSpot.validateResults(originalIns, []reflect.Value{oneOut}) {
					cacheSpot.singleStoreInCache(oneOut, strKey)
				}
			}()
		}

		valToReturn = oneOut
		returnBool = hasReturn
	}

	return cacheSpot.putFirstResultEvidence(valToReturn, returnBool)
}

//store results in cache
func (cacheSpot CacheSpot) storeInCache(origOuts []reflect.Value) error {

	keys, values := cacheSpot.getKeysForOuts(origOuts)

	return cacheSpot.cacheValues(keys, values)
}


func (cacheSpot CacheSpot)convertManyReturnToOneReturn(manyOuts reflect.Value) (reflect.Value, bool) {

	var hotOut reflect.Value
	hasReturn := manyOuts.Len() > 0
	if hasReturn {
		hotOut = manyOuts.Index(0)
	} else {
		hotOut = reflect.New(cacheSpot.spotOutType[0]).Elem()
	}

	return hotOut, hasReturn
}


//store results in cache
func (cacheSpot CacheSpot) cacheValues(keys []string, values []reflect.Value) error {
	numOut := len(values)

	cacheRegs := make([]cache.CacheRegistry, numOut)

	//iterate over all function returns. All of then can be stored
	for index := 0; index < numOut; index++ {

		//index := 0 // hard coded index. refactor to use any quantity of return valures ASAP
		//setting cache
		cacheId := keys[index]

		//a empty cachekey means that this <value will not be stored
		if len(cacheId) > 0 {

			//get raw value
			valRet := values[index].Interface()

			ttl := discoverTTL(valRet, -1)

			log.Debug("TTL for reg %v %v!", cacheId, ttl)

			//invoke cache manager to persist returned value
			cacheRegs[index] = cache.CacheRegistry{CacheKey: cacheId, Payload: valRet, Ttl: ttl, HasValue: true}

		}
	}

	log.Debug("saving registries %s!", keys)
	err := cacheSpot.CacheManager.SetCache(cacheRegs...)
	if err != nil {
		log.Error("Erro trying to save cache keys %v, error %v!", keys, err)
		return err
	}

	return nil
}

//store results in cache
func (cacheSpot CacheSpot) singleStoreInCache(hotOut reflect.Value, cacheKey string) {
	//TODO REUSE THIS FUNCTION AT FORMER StoreInCache FUNCTION
	//a empty cachekey means that this <value will not be stored
	if len(cacheKey) > 0 {

		log.Debug("saving registry %s!", cacheKey)
		//get raw value
		valRet := hotOut.Interface()

		ttl := discoverTTL(valRet, -1)

		log.Debug("TTL for reg %v %v!", cacheKey, ttl)

		//invoke cache manager to persist returned value
		cacheRegistry := cache.CacheRegistry{CacheKey: cacheKey, Payload: valRet, Ttl: ttl, HasValue: true}
		cacheSpot.CacheManager.SetCache(cacheRegistry)
		log.Debug("registry %s saved successfully!", cacheKey)
	}
}

//Determine cache keys, based on function parameters (in array) and outTypes
func (cacheSpot CacheSpot) cacheKeysDyn(in reflect.Value) ([]string, error) {

	if isMany(in.Type()) {
		qtdIns := in.Len() // how many ids was requested
		keys := make([]string, qtdIns)

		for i := 0; i < qtdIns; i++ {
			keys[i] = cacheSpot.getKeyForInput(in.Index(i))
		}

		return keys, nil

	} else {
		key := cacheSpot.getKeyForInput(in)
		return []string{key}, nil

	}
}

//retur a equivalent cache key for a input parameter
func (cacheSpot CacheSpot) getKeyForInput(valueIn reflect.Value) string {

	outType := cacheSpot.spotOutInnType[0]

	//if is not possible to turn the first paramate to string, fail! Cache wll be missed!!
	strVal, err := valIntToString(valueIn)
	if err != nil {
		log.Error(" ERROR TRYING TO PARSE A CACHE KEY FOR %v. %v %v !", err, valueIn, outType)
		panic(err)
	}

	var keyPrefix string
	if cacheSpot.CacheIdPrefix != nil {
		keyPrefix = *cacheSpot.CacheIdPrefix
	} else {
		keyPrefix = outType.Name()
	}

	key := keyPrefix + ":" + strVal

	return key
}


func (cacheSpot CacheSpot)dynamicCall(inputs []reflect.Value) []reflect.Value {

	emptyFunction := cacheSpot.CachedFunc
	concreteFunction := cacheSpot.HotFunc
	inTypes, _ := getInOutTypes(reflect.TypeOf(concreteFunction))
	_, outTypes := getInOutTypes(reflect.TypeOf(emptyFunction))

	ci_IsMany := isMany(inTypes[0])
	ei_IsMany := isValMany(inputs[0])

	var finalResponse []reflect.Value

	if ei_IsMany == ci_IsMany { // same kind of functions
		finalResponse = reflect.ValueOf(concreteFunction).Call(inputs) // simple like that

	} else if ei_IsMany && !ci_IsMany {

		arrInputs := fromWrappedToArray(inputs[0])

		responses := make([]reflect.Value, len(arrInputs))

		for index, input := range arrInputs {
			response := reflect.ValueOf(concreteFunction).Call([]reflect.Value{input})
			responses[index] = response[0] // take only first return value
		}

		wrappedResp := fromArrayToWrapped(responses, outTypes[0])

		newDefVal := cacheSpot.defaultValues(true) //TODO: checkPositiveResponse() function

		arrOuts := make([]reflect.Value, len(outTypes))
		arrOuts[0] = wrappedResp

		for index := 1; index < len(outTypes); index++ {
			val := newDefVal[index]
			arrOuts[index] = val // set other returns as default
		}
		finalResponse = arrOuts
	}

	return finalResponse

}


func (cacheSpot CacheSpot)resumeFoundedItens(allInputParams []reflect.Value, nfIns []reflect.Value) []reflect.Value {

	typeDestiny := cacheSpot.spotInType[0]
	newAllParams := make([]reflect.Value, len(allInputParams))

	newAllParams[0] = fromArrayToWrapped(nfIns, typeDestiny)

	for i := 1; i < len(allInputParams); i++ {
		newAllParams[i] = allInputParams[i]
	}

	return newAllParams
}


//return value validation
func (cacheSpot CacheSpot)getKeysForOuts(outs []reflect.Value) ([]string, []reflect.Value) {
	//try to convert a function in a ValidateResults interface
	functionValidator, hasValidatorImpl := cacheSpot.CachedFunc.(SpecifyOutKeys)

	if hasValidatorImpl {
		return functionValidator.KeysForCache(outs)

	} else {
		numOuts := len(outs)

		keysToCache := []string{}
		outsToCache := []reflect.Value{}

		for i := 0; i < numOuts; i++ {
			out := outs[i]
			realVal := out.Interface()
			cacheable, isCacheable := realVal.(cache.Cacheable)

			if isCacheable {
				keysToCache = append(keysToCache, cacheable.GetCacheKey())
				outsToCache = append(outsToCache, out)

			} else {
				log.Warning("The object %v doesn't implements Cacheable, and function %v doesn't implements SpecifyOutKeys! Resulta value will not be cached!")
			}
		}

		return keysToCache, outsToCache
	}
}


func (cacheSpot CacheSpot)convertOneCallToManyCall(oneCallIns []reflect.Value) []reflect.Value {

	eI := cacheSpot.spotInType
	cI := cacheSpot.realInType
	qtdInputs := len(eI)
	manyCallIns := make([]reflect.Value, qtdInputs)

	fakeIn := reflect.MakeSlice(cI[0], 1, 1)
	fakeIn.Index(0).Set(oneCallIns[0])
	manyCallIns[0] = fakeIn

	for i := 1; i < qtdInputs; i++ { // not the first
		manyCallIns[i] = oneCallIns[i]
	}

	return manyCallIns
}

//check whether it's possible to swap between this functions at application loading
//must panic if is not
func (cacheSpot CacheSpot) mustBePossibleToSwap() {

	//basic validation of emptyBodyFunction pre requirements
	cacheSpot.mustBeCompatibleSignatures()

	//must be possible to check return value validation before call cache storage
	cacheSpot.mustHaveValidationMethod()

	//basic validation of emptyBodyFunction pre requirements
	cacheSpot.mustHaveKeyDefiner()

	//cardinality of inputs and outputs must be the same on the same function signature (many->many or one->one)
	//between two function signatures it is possible to swap
	cacheSpot.validateCardinality()

	//check for default values
	cacheSpot.defaultValues(true)
}

func (cacheSpot CacheSpot) callHotFunction(allInputParams []reflect.Value, notCachedIns []reflect.Value) []reflect.Value {
	if len(notCachedIns) > 0 {
		newAllParams := cacheSpot.resumeFoundedItens(allInputParams, notCachedIns)
		fullResponse := cacheSpot.dynamicCall(newAllParams)
		hotReturnedValues := fromWrappedToArray(fullResponse[0])
		return hotReturnedValues
	} else {
		return []reflect.Value{}
	}
}


//check whether functions are compatible or not
func (c CacheSpot) mustBeCompatibleSignatures() {

	mustBePointer(c.CachedFunc)

	ins_a, outs_a := getInOutTypes(reflect.TypeOf(c.CachedFunc))
	ins_b, outs_b := getInOutTypes(reflect.TypeOf(c.HotFunc))

	if len(ins_a) > 1 || len(ins_b) > 1 {
		log.Warning("In multiple input functions, only the first paramater will be considered as cache key!")
	}

	firstIAType := getArrayInnerType(ins_a[0])
	firstIBType := getArrayInnerType(ins_b[0])

	firstOAType := getArrayInnerType(outs_a[0])
	firstOBType := getArrayInnerType(outs_b[0])

	mustBeCompatible(firstIAType, firstIBType)
	mustBeCompatible(firstOAType, firstOBType)

}

func (c CacheSpot) validateCardinality() {

	ine, oute := getInOutTypes(reflect.TypeOf(c.CachedFunc))
	ino, outo := getInOutTypes(reflect.TypeOf(c.HotFunc))

	//validate returns
	if len(ine) == 0 || len(oute) == 0 || len(ino) == 0 || len(outo) == 0 {
		panic(errors.New("The len of ins and outs of cached functions must be grater than 0!"))
	}

	//validate same cardinality for empty function
	if isMany(ine[0]) != isMany(oute[0]) { //XOR
		panic(errors.New("the empty destiny function doesn't have the same cardinality for first paramater and first return!"))
	}

	//validate same cardinality for original hot function
	if isMany(ino[0]) != isMany(outo[0]) { //XOR
		panic(errors.New("the original function can't be cached, because it doesn't have the same cardinality for first paramater and first return!"))
	}

}

//analyze and define if some result is valid. Usually used before a cache operation
func (c CacheSpot) validateResults(in []reflect.Value, out []reflect.Value) bool {

	//try to convert a function in a ValidateResults interface
	functionValidator, hasValidatorImpl := c.CachedFunc.(ValidateResults)

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

		log.Error("Erro ", c.CachedFunc)
		funcName := reflect.TypeOf(c.CachedFunc).Name()
		log.Error("", errors.New("Its not possible to infer a return value validation. Your function " + funcName + " need to implement ValidateResults inferface!"))
		return false
	}

}

//return value validation
func (cacheSpot CacheSpot) mustHaveKeyDefiner() {

	emptyBodyFunction := cacheSpot.CachedFunc

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

		firstType := cacheSpot.spotOutType[0]

		if isMany(firstType) {

			innerType := cacheSpot.spotOutInnType[0]

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
func (c CacheSpot) mustHaveValidationMethod() {

	emptyBodyFunction := c.CachedFunc

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
func (cacheSpot CacheSpot) defaultValues(defBool bool) []reflect.Value {

	outTypes := cacheSpot.spotOutType
	emptyFunction := cacheSpot.CachedFunc

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
func (cacheSpot CacheSpot) setSwapAsFunctionBody() {

	//recover the value for function in the pointer
	fn := reflect.ValueOf(cacheSpot.CachedFunc).Elem()

	//put a recently created swap function as a function body for the emptyBodyFunction
	fn.Set(reflect.MakeFunc(fn.Type(), cacheSpot.swapFunction))


	//	Creating cache spot for Swap aop.FindOneType to calling aop.FindOneCustomer


	log.Debug("Creating cache spot: %v->(cache return)->%v ", cacheSpot.cachedFuncName, cacheSpot.hotFunctName)

}
