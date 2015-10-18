package aop

import (
	"reflect"
)

//Specialize validation results. Only valid results will be cached.
//called in one to any and many to any calls
//ValidateResults
type TypeValidateResults func(allIns []reflect.Value, allOuts []reflect.Value, cacheKey string, singleValueToCache interface{}) bool

//To be implemented by a function that will need to define keys for cache
//SpecifyCacheKeys
type TypeSpecifyInputKeys func(in []reflect.Value, outTypes []reflect.Type) []string

//contract to be implemented by a function that will need to define keys for cache
//determine cache key for each return value
//empty string cachekey means that default value must be used
//SpecifyOutKeys
type TypeSpecifyOutputKeys func(ins []reflect.Value, out []reflect.Value) ([]string, []reflect.Value)

//implemented by a function that needs to define default values to their returns
//determine default values for each return
//an string empty cache key (see method above) means that default value must be used
//DefaultValubleFunction
type TypeCreateDefVals func(success bool, outTypes []reflect.Type) []reflect.Value
