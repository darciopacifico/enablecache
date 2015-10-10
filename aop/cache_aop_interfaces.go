package aop

import (
	"reflect"
)

//implemented by a function that needs to validate their values before caching operation
//denotes that this function has a validator function
type ValidateResults interface {
	//check whether results must be cached
	IsValidResults(in []reflect.Value, out []reflect.Value) bool
}

//contract to be implemented by a function that will need to define keys for cache
type SpecifyCacheKeys interface {

	//determine cache key for each return value
	//empty string cachekey means that default value must be used
	CacheKeys(in []reflect.Value, outTypes []reflect.Type) []string
}

//implemented by a function that needs to define default values to their returns
//determine default values for each return
//an string empty cache key (see method above) means that default value must be used
type DefaultValubleFunction interface {
	DefaultValues(outTypes []reflect.Type) []reflect.Value
}
