package aop

import (
	"errors"
	"fmt"
	"reflect"

	"gitlab.wmxp.com.br/bis/biro/cache"
)

func FromWrappedToArray(wrappedResponseArray reflect.Value) []reflect.Value {

	hotResponseLen := wrappedResponseArray.Len()
	hotReturnedValues := make([]reflect.Value, hotResponseLen)
	for i := 0; i < hotResponseLen; i++ {
		hotReturnedValues[i] = wrappedResponseArray.Index(i)
	}

	return hotReturnedValues
}

func FromArrayToWrapped(arrayVales []reflect.Value, typeDestinySlice reflect.Type) reflect.Value {
	wrappedArray := reflect.MakeSlice(typeDestinySlice, len(arrayVales), len(arrayVales))

	for index, value := range arrayVales {
		wrappedArray.Index(index).Set(value)
	}

	return wrappedArray
}

func CallHotFunction(allInputParams []reflect.Value, notCachedIns []reflect.Value, concreteFunction interface{}, emptyFunction interface{}, typeDestiny reflect.Type) []reflect.Value {

	//inType, outType := getInOutTypes(concreteFunction)

	if len(notCachedIns) > 0 {
		newAllParams := ResumeFoundedItens(allInputParams, notCachedIns, typeDestiny)
		fullResponse := dynamicCall(concreteFunction, emptyFunction, newAllParams)
		hotReturnedValues := FromWrappedToArray(fullResponse[0])

		return hotReturnedValues
	} else {

		return []reflect.Value{}
	}

}

func dynamicCall(concreteFunction interface{}, emptyFunction interface{}, inputs []reflect.Value) []reflect.Value {

	inTypes, _ := getInOutTypes(reflect.TypeOf(concreteFunction))
	_, outTypes := getInOutTypes(reflect.TypeOf(emptyFunction))

	ci_IsMany := isMany(inTypes[0])
	ei_IsMany := isValMany(inputs[0])

	var finalResponse []reflect.Value

	if ei_IsMany == ci_IsMany { // same kind of functions
		finalResponse = reflect.ValueOf(concreteFunction).Call(inputs) // simple like that

	} else if ei_IsMany && !ci_IsMany {

		arrInputs := FromWrappedToArray(inputs[0])

		responses := make([]reflect.Value, len(arrInputs))

		for index, input := range arrInputs {
			response := reflect.ValueOf(concreteFunction).Call([]reflect.Value{input})
			responses[index] = response[0] // take only first return value
		}

		wrappedResp := FromArrayToWrapped(responses, outTypes[0])

		newDefVal := defaultValues(emptyFunction, outTypes, true) //TODO: checkPositiveResponse() function

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

func ResumeFoundedItens(allInputParams []reflect.Value, nfIns []reflect.Value, typeSliceDestiny reflect.Type) []reflect.Value {

	newAllParams := make([]reflect.Value, len(allInputParams))

	newAllParams[0] = FromArrayToWrapped(nfIns, typeSliceDestiny)

	for i := 1; i < len(allInputParams); i++ {
		newAllParams[i] = allInputParams[i]
	}

	return newAllParams
}

//return value validation
func getKeysForOuts(outs []reflect.Value, emptyBodyFunction interface{}) ([]string, []reflect.Value) {
	//try to convert a function in a ValidateResults interface
	functionValidator, hasValidatorImpl := emptyBodyFunction.(SpecifyOutKeys)

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

func convertOneCallToManyCall(oneCallIns []reflect.Value, eI []reflect.Type, cI []reflect.Type) []reflect.Value {

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

func convertManyReturnToOneReturn(manyOuts reflect.Value, eO []reflect.Type) (reflect.Value, bool) {

	var hotOut reflect.Value
	hasReturn := manyOuts.Len() > 0
	if hasReturn {
		hotOut = manyOuts.Index(0)
	} else {
		hotOut = reflect.New(eO[0]).Elem()
	}

	return hotOut, hasReturn
}

//check whether it's possible to swap between this functions at application loading
//must panic if is not
func mustBePossibleToSwap(emptyFunction interface{}, concreteFunction interface{}) {

	//basic validation of emptyBodyFunction pre requirements
	mustBeCompatibleSignatures(emptyFunction, concreteFunction)
	//must be possible to check return value validation before call cache storage
	mustHaveValidationMethod(emptyFunction)

	//basic validation of emptyBodyFunction pre requirements
	mustHaveKeyDefiner(emptyFunction)

	eI, eO := getInOutTypes(reflect.TypeOf(emptyFunction))
	cI, cO := getInOutTypes(reflect.TypeOf(concreteFunction))

	//cardinality of inputs and outputs must be the same on the same function signature (many->many or one->one)
	//between two function signatures it is possible to swap
	assureFunctionValid(eI, eO, cI, cO)

	//check for default values
	defaultValues(emptyFunction, eO, true)
}

//check whether functions are compatible or not
func mustBeCompatibleSignatures(emptyFunction interface{}, concreteFunction interface{}) {
	mustBePointer(emptyFunction)

	ins_a, outs_a := getInOutTypes(reflect.TypeOf(emptyFunction))
	ins_b, outs_b := getInOutTypes(reflect.TypeOf(concreteFunction))


	if(len(ins_a) > 1 || len(ins_b) > 1){
		log.Warning("In multiple input functions, only the first paramater will be considered as cache key!")
	}

	firstIAType := getArrayInnerType(ins_a[0])
	firstIBType := getArrayInnerType(ins_b[0])

	firstOAType := getArrayInnerType(outs_a[0])
	firstOBType := getArrayInnerType(outs_b[0])

	mustBeCompatible(firstIAType, firstIBType)
	mustBeCompatible(firstOAType, firstOBType)

}

func mustBeCompatible(a, b reflect.Type) {
	if !a.ConvertibleTo(b) || !a.AssignableTo(b) {
		panic(errors.New(fmt.Sprintf(
			"Types '%v' and '%v' is not compatile to each other! "+
				"It is no possible to make a swap function that "+
				"return or receive different kinds of objects!", a.Name(), b.Name())))
	}
}

func mustBePointer(someValues ...interface{}) {
	for _, value := range someValues {
		if reflect.ValueOf(value).Kind() != reflect.Ptr {
			log.Error("Is not a pointer!!")
			panic(errors.New("Paramater needs to be a pointer!"))
		}
	}
}

func assureFunctionValid(ine, oute, ino, outo []reflect.Type) {
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
