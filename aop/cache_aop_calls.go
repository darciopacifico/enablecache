package aop

import (
	"reflect"
)



//execute an one to one reflection + cache operation
func (cacheSpot CacheSpot) callOneToOne(originalIns []reflect.Value) (returnValue []reflect.Value) {
	defer func() { //assure for not panicking
		if r := recover(); r != nil {
			log.Error("Recovering! Error trying to call a swap function one-one!!  %v", r)
			log.Error("Falling back this request to direct hot function call, without cache!")

			returnValue = cacheSpot.callHotFunc(originalIns)
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
		hotOuts := cacheSpot.callHotFunc(originalIns)

		//store in cache
		cacheSpot.storeCacheOneOne(originalIns, hotOuts, cacheKey, hotOuts[0])

		return hotOuts
	}
}

//execute an many to many call
func (cacheSpot CacheSpot) callManyToMany(originalIns []reflect.Value) (returnVal []reflect.Value) {

	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovering! Error trying to call a swap function!! %v", r)
			log.Error("Falling back this request to direct hot function call, without cache!")

			returnVal = cacheSpot.callHotFunc(originalIns)

			return
		}
	}()

	return cacheSpot.callManyToAny(originalIns)
}

//execute an many to one call
func (cacheSpot CacheSpot) callManyToOne(originalIns []reflect.Value) (returnValue []reflect.Value) {

	defer func() { //assure for not panicking
		if r := recover(); r != nil {
			log.Error("Recovering! Error trying to call a swap function!! (callManyToOne) %v", r)
			log.Error("Falling back this request to direct hot function call, without cache!")

			hotReturnedValues := cacheSpot.callNotFoundedInputs(originalIns, fromWrappedToArray(originalIns[0]))
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
	hotReturnedValues := cacheSpot.callNotFoundedInputs(originalIns, notCachedIns)

	if len(hotReturnedValues) > 0 {
		cacheSpot.storeManyToAny(notCachedIns, hotReturnedValues)
	}

	joinedReturn := append(cachedOuts, hotReturnedValues...)

	return cacheSpot.putFirstArrResultEvidence(joinedReturn)
}


func (cacheSpot CacheSpot) callOneToMany(originalIns []reflect.Value) (returnValue []reflect.Value) {

	defer func() { //assure for not panicking out
		if r := recover(); r != nil {

			log.Error("Recovering! Error trying to call a swap function!! (callOneToMany) %v", r)
			log.Error("Falling back this request to direct hot function call, without cache!")

			fakeIns := cacheSpot.convertOneCallToManyCall(originalIns)
			deferOuts := cacheSpot.callHotFunc(fakeIns)
			oneReturn, _ := cacheSpot.convertManyReturnToOneReturn(deferOuts[0])

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
		manyOuts := cacheSpot.callHotFunc(fakeIns)
		oneOut, hasReturn := cacheSpot.convertManyReturnToOneReturn(manyOuts[0])

		if hasReturn { // returned array is greater that 0
			cacheSpot.storeCacheOneOne(fakeIns, manyOuts, strKey, oneOut)
		}

		valToReturn = oneOut
		returnBool = hasReturn
	}

	return cacheSpot.putFirstResultEvidence(valToReturn, returnBool)
}

func (cacheSpot CacheSpot) convertManyReturnToOneReturn(manyOuts reflect.Value) (reflect.Value, bool) {

	var hotOut reflect.Value
	hasReturn := manyOuts.Len() > 0
	if hasReturn {
		hotOut = manyOuts.Index(0)
	} else {
		hotOut = reflect.New(cacheSpot.spotOutType[0]).Elem()
	}

	return hotOut, hasReturn
}

func (cacheSpot CacheSpot) dynamicCall(inputs []reflect.Value) []reflect.Value {

	inTypes, _ := getInOutTypes(reflect.TypeOf(cacheSpot.HotFunc))
	_, outTypes := getInOutTypes(reflect.TypeOf(cacheSpot.CachedFunc))

	ci_IsMany := isMany(inTypes[0])
	ei_IsMany := isValMany(inputs[0])

	var finalResponse []reflect.Value

	if ei_IsMany == ci_IsMany { // same kind of functions
		finalResponse = cacheSpot.callHotFunc(inputs)

	} else if ei_IsMany && !ci_IsMany {

		arrInputs := fromWrappedToArray(inputs[0])

		responses := make([]reflect.Value, len(arrInputs))

		for index, input := range arrInputs {

			inputs[0] = input
			response := cacheSpot.callHotFunc(inputs) //TODO USE A POOL OF GO ROUTINES TO CALL HOT FUNCTION CONCURRENTLY
//			response := cacheSpot.callHotFunc([]reflect.Value{input}) //TODO USE A POOL OF GO ROUTINES TO CALL HOT FUNCTION CONCURRENTLY


			responses[index] = response[0] // take only first return value
		}

		wrappedResp := fromArrayToWrapped(responses, outTypes[0])

		newDefVal, _ := cacheSpot.getDefVals(true) //TODO: checkPositiveResponse() function

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

func (cacheSpot CacheSpot) convertOneCallToManyCall(oneCallIns []reflect.Value) []reflect.Value {

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

//call the hot function using reflection
func (cacheSpot CacheSpot) callHotFunc(inputs []reflect.Value) []reflect.Value {
	return reflect.ValueOf(cacheSpot.HotFunc).Call(inputs)
}

func (cacheSpot CacheSpot) callNotFoundedInputs(allInputParams []reflect.Value, notCachedIns []reflect.Value) []reflect.Value {
	if len(notCachedIns) > 0 {
		newAllParams := cacheSpot.resumeFoundedItens(allInputParams, notCachedIns)
		fullResponse := cacheSpot.dynamicCall(newAllParams)
		hotReturnedValues := fromWrappedToArray(fullResponse[0])
		return hotReturnedValues
	} else {
		return []reflect.Value{}
	}
}

func (cacheSpot CacheSpot) resumeFoundedItens(allInputParams []reflect.Value, nfIns []reflect.Value) []reflect.Value {

	typeDestiny := cacheSpot.spotInType[0]
	newAllParams := make([]reflect.Value, len(allInputParams))

	newAllParams[0] = fromArrayToWrapped(nfIns, typeDestiny)

	for i := 1; i < len(allInputParams); i++ {
		newAllParams[i] = allInputParams[i]
	}

	return newAllParams
}
