package aop

import (
	"reflect"
	"sync"
	"github.com/darciopacifico/enablecache/cache"
	"github.com/op/go-logging"
)

var EMPTY_MAP = make(map[string]cache.CacheRegistry)
var typeCacheable = reflect.TypeOf((*cache.Cacheable)(nil)).Elem()
var log = logging.MustGetLogger("cache")
var errorInterfaceModel = reflect.TypeOf((*error)(nil)).Elem()


//template to make a cache spot function
type CacheSpot struct {
	CachedFunc        interface{}        //(required) empty function ref, that will contain cacheable function. Pass a nil reference
	HotFunc           interface{}        //(required) real hot function, that will have results cached
	CacheManager      cache.CacheManager //(required) cache manager for cache swaped function
	StoreOnly         bool               //(Optional) mark if cache manager can take cached values or just store results
	CacheIdPrefix     string             //(Optional) cache prefix for cache registries
	ValidateResults   TypeValidateResults
	SpecifyInputKeys  TypeSpecifyInputKeys
	SpecifyOutputKeys TypeSpecifyOutputKeys
	DefValsFunction   TypeCreateDefVals
	wg                *sync.WaitGroup
	CallContext       // will be mounted at start up nothing to do
}

//reflect objects need to reflect function call
type CallContext struct {
	spotOutType    []reflect.Type
	spotOutInnType []reflect.Type
	spotInType     []reflect.Type
	spotInInnType  []reflect.Type
	realInType     []reflect.Type
	realInInnType  []reflect.Type
	defValsSuccess []reflect.Value
	defValsFail    []reflect.Value
	cachedFuncName string // cached func name
	hotFunctName   string // hot func name
	isManyOuts     bool   //compose cardinality of swap call
	isManyIns      bool   //compose cardinality of swap call

}

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
