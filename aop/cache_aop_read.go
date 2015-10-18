package aop

import (
	"github.com/darciopacifico/enablecache/cache"
	"reflect"
)

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
		//panic(errCK) // fckp
		return emptyMap
	}

	cacheRegMap, err := cacheSpot.CacheManager.GetCaches(keys...)
	if err != nil {
		log.Error("Error trying to retrieve cache data x", errCK)
		emptyMap := make(map[string]cache.CacheRegistry, 0)
		//panic(err) // fckp
		return emptyMap
	}

	return cacheRegMap
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
	strVal, err := valToString(valueIn)
	if err != nil {
		log.Error(" ERROR TRYING TO PARSE A CACHE KEY FOR %v. %v %v !", err, valueIn, outType)
		panic(err)
	}

	var keyPrefix string
	if len(cacheSpot.CacheIdPrefix) > 0 {
		keyPrefix = cacheSpot.CacheIdPrefix
	} else {
		keyPrefix = outType.Name()
	}

	key := keyPrefix + ":" + strVal

	return key
}
