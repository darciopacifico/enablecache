package aop

import (
	"github.com/darciopacifico/enablecache/cache"
	"reflect"
)

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
		retVals = c.defValsFail
	}

	//fix return value type. Emulates a polimorphic behaviour not present in golang
	// ex: an struct Customer{} returned as an interface{} wouldn't work,
	// after be recovered by the cache mechanism and used in an reflect return operation
	return c.fixReturnTypes(retVals)
}

//fix return type acoordingly to out type
func (c CacheSpot) fixReturnTypes(values []reflect.Value) []reflect.Value {

	values[0] = fixReturnTypes(c.spotOutType[0], values[0])
	/*
		for i,v:=range values{
			values[i] = fixReturnTypes(c.spotOutType[i], v)
		}
	*/

	return values
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

//Take the substitute for first value and join with default values for other results
func (cacheSpot CacheSpot) putFirstArrResultEvidence(hotResult []reflect.Value) []reflect.Value {

	firstResult := fromArrayToWrapped(hotResult, cacheSpot.spotOutType[0]) // set first value as joined return

	return cacheSpot.putFirstResultEvidence(firstResult, true)
}

//Take the substitute for first value and join with default values for other results
func (cacheSpot CacheSpot) putFirstResultEvidence(firstResult reflect.Value, success bool) []reflect.Value {
	defVals, lenDefVals := cacheSpot.getDefVals(success)

	//put first result in evidence
	arrOuts := make([]reflect.Value, lenDefVals)
	arrOuts[0] = firstResult

	//set default values to others
	for index := 1; index < lenDefVals; index++ {
		arrOuts[index] = defVals[index] // set other returns as default
	}

	return arrOuts
}

func (cacheSpot CacheSpot) getDefVals(success bool) ([]reflect.Value, int) {
	var defVals []reflect.Value

	if success {
		defVals = cacheSpot.defValsSuccess
	} else {
		defVals = cacheSpot.defValsFail
	}

	return defVals, len(defVals)
}
