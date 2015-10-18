package aop

import (
	"errors"
	"github.com/darciopacifico/enablecache/cache"
	"github.com/op/go-logging"
	"reflect"
)

var EMPTY_MAP = make(map[string]cache.CacheRegistry)
var typeCacheable = reflect.TypeOf((*cache.Cacheable)(nil)).Elem()
var log = logging.MustGetLogger("cache")
var errorInterfaceModel = reflect.TypeOf((*error)(nil)).Elem()

//recover default values for an function call
//non cached values (string empty cache key), will be returned as a default values, by this function or by implementing DefaultValubleFunction interface
func (cacheSpot CacheSpot) defaultValues(success bool) []reflect.Value {

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
