package aop

import (
	"errors"
	"os"
	"reflect"
	"strconv"

	"github.com/darciopacifico/cachengo/cache"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("cache")

//
var errorInterfaceModel = reflect.TypeOf((*error)(nil)).Elem()

//just to formalize the signature of swap function
type typeSwapFunc func(ins []reflect.Value) []reflect.Value

// cache spot configuration
type CacheSpot struct {
	CachedSpotFunction interface{}        //empty function ref, that will contain cacheable function. Pass a nil reference
	OriginalFunction   interface{}        //real hot function, that will have results cached
	CacheManager       cache.CacheManager //cache manager implementation
	TakeCache          bool               //mark if cache manager can take cached values or just store results

	Name          string //(Optional) cache spot config for log and metrics.
	CacheIdPrefix string //(Optional) cache prefix for cache registries
}

func MakeSwap(CachedSpotFunction interface{}, OriginalFunction   interface{}, CacheManager       cache.CacheManager, TakeCache          bool              ){

	MakeCachedSpotFunction(CacheSpot{
		CachedSpotFunction:  CachedSpotFunction,
		OriginalFunction  :  OriginalFunction  ,
		CacheManager :  CacheManager ,
		TakeCache: TakeCache,
	})

}

//Put in the emptyBodyFunction the original function call and caching mechanism
//WILL EXIT APPLICATION IF SOME PREREQ WAS NOT ACCOMPLISHED
//MUST BE CALLED AT APPLICATION STARTUP ONCE AND ONLY ONCE PER FUNCTION.
func MakeCachedSpotFunction(cacheSpot CacheSpot) {

	/*
		if err := notSameSignature(emptyBodyFunction, originalFunction); err != nil {
			log.Error("As funcoes %v e %v nao possuem a mesma assinatura!", err)
			os.Exit(1)
		}
	*/

	//basic validation of emptyBodyFunction pre requirements
	if isValidationOK, err := hasValidationMethod(cacheSpot.CachedSpotFunction); !isValidationOK {
		//if this function doesnt implement ValidateResults and is not possible to infer return validation
		//will exit imediatelly. Application must not run with any inconsistency
		log.Error("Initilization Error:", err)
		os.Exit(1)
	}

	//build a swap function that can call the original function and cache results
	swapFunction := getSwapFunctionForCache(cacheSpot)

	//set emptyBodyFunction body with swapFunction containing cache mechanism
	setSwapAsFunctionBody(cacheSpot.CachedSpotFunction, swapFunction)
}

//create a swap function for cache operation
func getSwapFunctionForCache(cacheSpot CacheSpot) typeSwapFunc {

	//vars that must stay out of closure. execute once and only once
	//retrieve the out types for the function
	outTypes := getOutTypes(cacheSpot.OriginalFunction)
	//default values for each return of a function

	defaultVals, errDefVal := defaultValues(cacheSpot.CachedSpotFunction, outTypes)

	if errDefVal != nil {
		log.Error("Error trying to define default values for function %v. %v ", cacheSpot.OriginalFunction, errDefVal)
		os.Exit(1)
	}

	//swap function implementation for cache operation
	swap := func(ins []reflect.Value) []reflect.Value {
		//keys array, based on inputs and return types
		keys, errCK := cacheKeys(cacheSpot.CachedSpotFunction, ins, outTypes, cacheSpot.CacheIdPrefix)
		if errCK != nil {
			log.Error("Error trying to solve cache keys! Is not possible to proceed with cache operations!", errCK)
		}

		//try to find a cached value
		if cacheSpot.TakeCache && errCK == nil {
			values, hasCachedValue, err := getCache(keys, defaultVals, cacheSpot.CacheManager)
			if err != nil {
				log.Error("Error trying to retrieve cache data", errCK)
			}
			if hasCachedValue {

				values := fixReturnTypes(outTypes, values)

				return values
			}
		}
		//CACHED DATA OPERATIONS AREA CODE ABOVE
		//==========================================================
		//NOT CACHED DATA OPERATIONS CODE BELOW

		//legacy execution
		outs := reflect.ValueOf(cacheSpot.OriginalFunction).Call(ins) //ok there is no cached value. lets call the original function

		// if some error happens trying to define CacheKeys, cache operations will be canceled at all
		if errCK == nil {
			//start a new go routine to store cache data
			go validateStoreInCache(cacheSpot, ins, outs, keys)
		}
		return outs
	}

	return swap
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

//validate results and store in cache.
//assure for not panicking in goroutine for validating and store
func validateStoreInCache(cacheSpot CacheSpot, ins2 []reflect.Value, outs2 []reflect.Value, keys2 []string) {

	defer func() { //assure for not panicking
		if r := recover(); r != nil {
			log.Error("Recovering! Error trying to save cache registry! %v", r)
		}
	}()

	// check whether results are valid and must be cached
	if validateResults(cacheSpot.CachedSpotFunction, ins2, outs2) {
		storeInCache(outs2, keys2, cacheSpot.CacheManager)
	}
}

//return the the out types for a function
func getOutTypes(originalFunction interface{}) []reflect.Type {

	functionType := reflect.TypeOf(originalFunction)
	numOut := functionType.NumOut()

	//recover the return types
	outTypes := make([]reflect.Type, numOut)
	for i := 0; i < numOut; i++ {
		outTypes[i] = functionType.Out(i)
	}

	return outTypes
}

//search inputs on cachemanager
func getCache(keys []string, defaultVals []reflect.Value, cacheManager cache.CacheManager) ([]reflect.Value, bool, error) {
	hasCachedValue := true //by default, has cache
	numOut := len(keys)    //
	values := make([]reflect.Value, numOut)

	//itera os retornos da funcao
	for index := 0; index < numOut; index++ {

		key := keys[index]
		defaultVal := defaultVals[index]

		//emtpy key means that will be used the defaultValue
		if len(key) > 0 {
			cacheRegistry, err := cacheManager.GetCache(key)
			if err != nil {
				return values, false, err
			}

			if cacheRegistry.HasValue {
				cacheVal := reflect.Indirect(reflect.ValueOf(cacheRegistry.Payload))
				values[index] = cacheVal
				hasCachedValue = true
			} else {
				hasCachedValue = false
				break
			}
		} else {
			//if key is a empty string, means that must return a default value, like a bool=true, or err=nil
			values[index] = defaultVal
		}
	}
	return values, hasCachedValue, nil
}

//store results in cache
func storeInCache(outs []reflect.Value, keys []string, cacheManager cache.CacheManager) {
	numOut := len(outs)

	//iterate over all function returns. All of then can be stored
	for index := 0; index < numOut; index++ {

		//index := 0 // hard coded index. refactor to use any quantity of return valures ASAP
		//setting cache
		cacheId := keys[index]

		//a empty cachekey means that this <value will not be stored
		if len(cacheId) > 0 {

			log.Debug("saving registry %s!", cacheId)
			//get raw value
			valRet := outs[index].Interface()

			ttl := discoverTTL(valRet, -1)

			log.Debug("TTL for reg %v %v!", cacheId, ttl)

			//invoke cache manager to persist returned value
			cacheRegistry := cache.CacheRegistry{CacheKey: cacheId, Payload: valRet, Ttl: ttl, HasValue: true}
			cacheManager.SetCache(cacheRegistry)
			log.Debug("registry %s saved successfully!", cacheId)
		}
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
func cacheKeys(emptyBodyFunction interface{}, in []reflect.Value, outTypes []reflect.Type, cacheIdPrefix string) ([]string, error) {

	determineKeysFunc, isDetermineKey := emptyBodyFunction.(SpecifyCacheKeys)

	if isDetermineKey {
		//function itself will determine cache keys
		return determineKeysFunc.CacheKeys(in, outTypes), nil //There is no foresee treatable errors.

	} else {

		var prefix string
		if len(cacheIdPrefix) > 0 {
			prefix = cacheIdPrefix
		} else {
			prefix = outTypes[0].Name()
		}

		//there is no SpecifyCacheKeys implemented. will try to use a default keys
		qtdOuts := len(outTypes)
		keys := make([]string, qtdOuts)

		//if is not possible to turn the first paramater to string, fail! Cache will be missed!!
		strVal, err := IntToString(in[0])

		//by default, the type of first return object will be used as cache prefix
		firstKey := prefix + ":" + strVal

		keys[0] = firstKey // first key will be returned...

		for i := 1; i < qtdOuts; i++ { //... other keys will be an empty string
			keys[i] = ""
		}

		return keys, err

	}
}

// Try to convert a int value to string. if is not a integer raise error
func IntToString(value reflect.Value) (string, error) {

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

//return value validation
func validateResults(emptyBodyFunction interface{}, in []reflect.Value, out []reflect.Value) bool {

	//try to convert a function in a ValidateResults interface
	functionValidator, hasValidatorImpl := emptyBodyFunction.(ValidateResults)

	//if function has a function with validation behaviour
	if hasValidatorImpl {
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
func hasValidationMethod(emptyBodyFunction interface{}) (bool, error) {
	if reflect.ValueOf(emptyBodyFunction).Kind() != reflect.Ptr {
		log.Error("emptyBodyFunction needs to be a pointer!")
		return false, errors.New("emptyBodyFunction needs to be a pointer!")
	}

	functionType := reflect.TypeOf(emptyBodyFunction)
	numOut := functionType.Elem().NumOut()
	funcName := functionType.Elem().Name()

	//try to convert a function in a ValidateResults interface
	_, hasValidatorImpl := emptyBodyFunction.(ValidateResults)

	//if function has a function with validation behaviour
	if hasValidatorImpl {
		return true, nil

	} else {
		log.Debug("Function %s dont implements ValidateResults", funcName)
		if numOut > 1 && functionType.Elem().Out(1).Kind() == reflect.Bool { //for more than one outs, the second one must be a boolean
			return true, nil

		} else {
			log.Debug("Erro", emptyBodyFunction)
			return false, errors.New("Its not possible to infer a return value validation. Your function '" + funcName + "' needs to implement ValidateResults inferface!")

		}
	}
}

//recover default values for an function call
//non cached values (string empty cache key), will be returned as a default values, by this function or by implementing DefaultValubleFunction interface
func defaultValues(emptyBodyFunction interface{}, outTypes []reflect.Type) ([]reflect.Value, error) {

	defValuable, isDefValuable := emptyBodyFunction.(DefaultValubleFunction)

	if isDefValuable {
		//there is a specific default values function
		defaultVals := defValuable.DefaultValues(outTypes)
		return defaultVals, nil
	}

	defaultValues := make([]reflect.Value, len(outTypes))

	for i, outType := range outTypes {

		switch outType.Kind() {

		case reflect.Struct:
			defaultValues[i] = reflect.New(outType).Elem()

		case reflect.Bool:
			defaultValues[i] = reflect.ValueOf(true)

		case reflect.Interface:
			if outType.Implements(errorInterfaceModel) {
				var err error = nil
				defaultValues[i] = reflect.ValueOf(&err).Elem()

			} else if i > 0 { // an interface at 0 position can be ignored

				log.Error("(not error) It's not possible to identify an default value for function %v return index %v!", defValuable, i)
				return []reflect.Value{}, errors.New("It's not possible to identify an default value for function!")
			}

		default:
			log.Error("(default) Its not possible to identify an default value for function %v return index %v!", defValuable, i)
			return []reflect.Value{}, errors.New("It's not possible to identify an default value for function!")
		}
	}

	return defaultValues, nil
}

//take the value of emptyBodyFunction and sets swap function as function body implementation
func setSwapAsFunctionBody(emptyBodyFunction interface{}, swap func([]reflect.Value) []reflect.Value) {

	//recover the value for function in the pointer
	fn := reflect.ValueOf(emptyBodyFunction).Elem()

	//put a recently created swap function as a function body for the emptyBodyFunction
	fn.Set(reflect.MakeFunc(fn.Type(), swap))
}
