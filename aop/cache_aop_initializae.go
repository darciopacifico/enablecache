package aop

import (
	"errors"
	"fmt"
	"reflect"
	"os"
)

//parse function signatures, validators,
//makes a swap function for reflection operations
func (cacheSpot CacheSpot) MustStartCache() CacheSpot {
	//gracefull exit at panic
	defer func() {
		//assure for not panicking
		if r := recover(); r != nil {
			log.Error("Error trying to make a swap function! %v", r)
			os.Exit(1)
		}
	}()

	if cacheSpot.WaitingGroup == nil {
		panic(errors.New("Waiting group == null! You must specify a new waiting group for this cacheSpot!"))
	}

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

	return cacheSpot
}

//analize functions and fill reflect objects as need
func (cacheSpot *CacheSpot) parseFunctionsForSwap() {

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
	cacheSpot.CallContext = CallContext{
		spotOutType:    SpotOutType,
		spotInType:     SpotInType,
		realInType:     RealInType,
		spotOutInnType: SpotInnOutType,
		spotInInnType:  SpotInnInType,
		realInInnType:  RealInnInType,
	}

	//adjust default values
	cacheSpot.defValsSuccess = cacheSpot.createDefaultValues(true)
	cacheSpot.defValsFail = cacheSpot.createDefaultValues(false)

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

//take the value of emptyBodyFunction and sets swap function as function body implementation
func (cacheSpot CacheSpot) setSwapAsFunctionBody() {

	//recover the value for function in the pointer
	fn := reflect.ValueOf(cacheSpot.CachedFunc).Elem()

	//put a recently created swap function as a function body for the emptyBodyFunction
	fn.Set(reflect.MakeFunc(fn.Type(), cacheSpot.swapFunction))

	//	Creating cache spot for Swap aop.FindOneType to calling aop.FindOneCustomer
	log.Debug("Creating cache spot: %v->(cache return)->%v ", cacheSpot.cachedFuncName, cacheSpot.hotFunctName)

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
	cacheSpot.mustValidateCardinality()

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

	if cacheSpot.SpecifyOutputKeys != nil {
		log.Debug("Function %v implements SpecifyOutKeys! Its OK!", funcName)
	} else {

		firstType := cacheSpot.spotOutType[0]

		if isMany(firstType) {

			innerType := cacheSpot.spotOutInnType[0]

			//... is a Cacheable implementation??
			// Cacheable is capable to define its own cache key
			if innerType.Implements(typeCacheable) {
				log.Debug("Function '%s' dont implements SpecifyOutKeys, but return type '%v' implements Cacheable! its ok!", funcName, innerType.Name())
			} else {
				panic(errors.New(fmt.Sprintf("Function '%s' doesn't implements SpecifyOutKeys and return type '%v' doesn't implements Cacheable!", funcName, innerType.Name())))
			}
		}

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

//return value validation
func (c CacheSpot) mustHaveValidationMethod() {

	emptyBodyFunction := c.CachedFunc

	if reflect.ValueOf(emptyBodyFunction).Kind() != reflect.Ptr {
		log.Error("emptyBodyFunction needs to be a pointer!")
		panic(errors.New("emptyBodyFunction needs to be a pointer!"))
	}

	functionType := reflect.TypeOf(emptyBodyFunction)
	numOut := functionType.Elem().NumOut()

	//if function has a function with validation behaviour
	if c.ValidateResults == nil {
		log.Debug("Function '%s' doesn't implements ValidateResults!", c.cachedFuncName)
		if numOut > 1 && functionType.Elem().Out(1).Kind() == reflect.Bool {
			//for more than one outs, the second one must be a boolean
			log.Warning("Atention! The second return value (a boolean) will be used as validation criteria! Implement ValidateResults inferface to define a special criteria!")

		} else {
			log.Warning("Atention! Your function has no validation criteria! Implement ValidateResults inferface to define a special validation criteria!")
			log.Warning("All returns will be cached!")
		}
	}
}

func (c CacheSpot) mustValidateCardinality() {

	ine, oute := getInOutTypes(reflect.TypeOf(c.CachedFunc))
	ino, outo := getInOutTypes(reflect.TypeOf(c.HotFunc))

	//validate returns
	if len(ine) == 0 || len(oute) == 0 || len(ino) == 0 || len(outo) == 0 {
		panic(errors.New("The len of ins and outs of cached functions must be grater than 0!"))
	}

	//validate same cardinality for empty function
	if isMany(ine[0]) != isMany(oute[0]) {
		//XOR
		panic(errors.New("the empty destiny function doesn't have the same cardinality for first paramater and first return!"))
	}

	//validate same cardinality for original hot function
	if isMany(ino[0]) != isMany(outo[0]) {
		//XOR
		panic(errors.New("the original function can't be cached, because it doesn't have the same cardinality for first paramater and first return!"))
	}

}


//recover default values for an function call
//non cached values (string empty cache key), will be returned as a default values, by this function or by implementing DefaultValubleFunction interface
func (cacheSpot CacheSpot) createDefaultValues(success bool) []reflect.Value {

	outTypes := cacheSpot.spotOutType

	if cacheSpot.DefValsFunction != nil {
		//there is a specific default values function
		defaultVals := cacheSpot.DefValsFunction(success, outTypes)
		return defaultVals
	}

	defaultValues := make([]reflect.Value, len(outTypes))

	for i, outType := range outTypes {

		switch outType.Kind() {

		case reflect.Struct, reflect.String:
			defaultValues[i] = reflect.New(outType).Elem()

		case reflect.Bool:
			defaultValues[i] = reflect.ValueOf(success)

		case reflect.Interface:
			if outType.Implements(errorInterfaceModel) {
				var err error = nil
				defaultValues[i] = reflect.ValueOf(&err).Elem()

			} else if i > 0 {
				// an interface at 0 position can be ignored

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
