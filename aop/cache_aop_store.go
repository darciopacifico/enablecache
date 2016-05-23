package aop

import (
	"github.com/darciopacifico/enablecache/cache"
	"reflect"
	"time"
)

//store results in cache
func (cacheSpot CacheSpot) cacheValues(notCachedIns []reflect.Value, origOuts []reflect.Value) error {
	keys, values := cacheSpot.getKeysForOuts(notCachedIns, origOuts)

	numOut := len(values)

	cacheRegs := make([]cache.CacheRegistry, 0, numOut)

	//iterate over all function returns. All of then can be stored
	for index := 0; index < numOut; index++ {

		//index := 0 // hard coded index. refactor to use any quantity of return valures ASAP
		//setting cache
		cacheId := keys[index]

		//a empty cachekey means that this <value will not be stored
		if len(cacheId) > 0 {

			//get raw value
			valRet := values[index].Interface()

			if cacheSpot.validateResults(notCachedIns, origOuts, cacheId, valRet) {

				log.Debug("TTL from cachespot 2!", cacheId, cacheSpot.Ttl)
				//invoke cache manager to persist returned value
				cacheRegs = append(cacheRegs, cache.CacheRegistry{
					CacheKey: cacheId,
					Payload: valRet,
					StoreTTL: cacheSpot.Ttl,
					CacheTime: time.Now(),
					HasValue: true,
					TypeName: ""})
			} else {
				log.Warning("Reg %v is not valid to cache!", cacheId)
			}
		}
	}

	log.Debug("saving registries %s!", keys)
	err := cacheSpot.CacheManager.SetCache(cacheRegs...)
	if err != nil {
		log.Error("Erro trying to save cache keys %v, error %v!", keys, err)
		return err
	}

	return nil
}

//store results in cache
func (cacheSpot CacheSpot) singleStoreInCache(hotOut reflect.Value, cacheKey string) {
	//TODO REUSE THIS FUNCTION AT FORMER StoreInCache FUNCTION
	//a empty cachekey means that this <value will not be stored
	if len(cacheKey) > 0 {

		log.Debug("saving registry %s!", cacheKey)
		//get raw value
		valRet := hotOut.Interface()

		log.Debug("TTL from cachespot !", cacheKey, cacheSpot.Ttl)

		//invoke cache manager to persist returned value
		cacheRegistry := cache.CacheRegistry{
			CacheKey: cacheKey,
			Payload: valRet,
			StoreTTL: cacheSpot.Ttl,
			CacheTime: time.Now(),
			HasValue: true}
		cacheSpot.CacheManager.SetCache(cacheRegistry)
		log.Debug("registry %s saved successfully!", cacheKey)
	}
}

//return value validation
func (cacheSpot CacheSpot) getKeysForOuts(ins []reflect.Value, outs []reflect.Value) ([]string, []reflect.Value) {
	//try to convert a function in a ValidateResults interface

	if cacheSpot.SpecifyOutputKeys != nil {
		return cacheSpot.SpecifyOutputKeys(ins, outs)

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

func (cacheSpot CacheSpot) storeCacheOneOne(originalIns []reflect.Value, hotOuts []reflect.Value, cacheKey string, valueToCache reflect.Value) {
	cacheSpot.WaitingGroup.Add(1)

	go func() {
		defer func() {
			//assure for not panicking

			if r := recover(); r != nil {
				log.Error("Recovering! Error trying to save cache registry y! %v", r)
			}

			cacheSpot.WaitingGroup.Done()
		}()

		// check whether results are valid and must be cached
		if cacheSpot.validateResults(originalIns, hotOuts, cacheKey, valueToCache.Interface()) {
			cacheSpot.singleStoreInCache(valueToCache, cacheKey)
		}
	}()
}

func (cacheSpot CacheSpot) storeManyToAny(notCachedIns []reflect.Value, hotReturnedValues []reflect.Value) {
	cacheSpot.WaitingGroup.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("Recovering! Error trying to save cache registry y! %v", r)
			}
			cacheSpot.WaitingGroup.Done()
		}()
		cacheSpot.cacheValues(notCachedIns, hotReturnedValues)
	}()
}



//analyze and define if some result is valid. Usually used before a cache operation
func (c CacheSpot) validateResults(allIns []reflect.Value, allOuts []reflect.Value, cacheKey string, value interface{}) bool {

	//if function has a function with validation behaviour
	if c.ValidateResults != nil {
		//custom validation
		return c.ValidateResults(allIns, allOuts, cacheKey, value)

	} else {

		//has some return value
		if len(allOuts) > 1 &&
		allOuts[1].IsValid() &&
		allOuts[1].Kind() == reflect.Bool {

			boolVal, _ := allOuts[1].Interface().(bool)

			return boolVal
		}

		log.Debug("Function '%v' doesn't implement ValidateResults inferface! All return will be cached!", c.cachedFuncName)
		return true
	}

}

