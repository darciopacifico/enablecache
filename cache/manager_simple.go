package cache

import (
	"github.com/op/go-logging"
	"time"
)

var log = logging.MustGetLogger("cache")

// define statistics util for iro
var st *Stats = NewStats("iro-cache")

//
type SimpleCacheManager struct {
	CacheStorage CacheStorage
}

func (c SimpleCacheManager) Validade() bool {
	return true
}

//invalidate cache registry
func (c SimpleCacheManager) Invalidate(cacheKeys ...string) error {
	errDel := c.CacheStorage.DeleteValues(cacheKeys...)
	if errDel != nil {
		log.Error("Error trying to delete values from cache %v", errDel)
	}
	return errDel
}

//set cache implementation
func (c SimpleCacheManager) SetCache(cacheRegistry ...CacheRegistry) error {
	//call cachestorage to store data
	return c.CacheStorage.SetValues(cacheRegistry...)
}

//implement getCache operation that can recover child data in other cache registries.
func (c SimpleCacheManager) GetCache(cacheKey string) (CacheRegistry, error) {

	//get the raw value from cache storage
	//this registry maybe missed some child reference, that will be check some lines below
	cacheRegistries, err := c.GetCaches(cacheKey)
	if err != nil {
		log.Error("Error trying to recover value from cache storage! %s", cacheKey)
		st.Miss()
		return CacheRegistry{
			cacheKey,
			nil,
			-2,
			time.Unix(0, 0),
			false,
			""}, err
	}
	if len(cacheRegistries) == 0 {
		log.Debug("Cache registry not found! %s", cacheKey)
		st.Miss()
		return CacheRegistry{
			cacheKey,
			nil,
			-2,
			time.Unix(0, 0),
			false,
			""}, nil
	}

	cacheRegistry := cacheRegistries[cacheKey]

	//cache miss for raw cache value!
	if !cacheRegistry.HasValue {
		st.Miss()
		return cacheRegistry, nil // empty, hasValue=false, cacheRegistry
	}

	//return final cache registry
	st.Hit()
	return cacheRegistry, nil

}

//implement getCache operation that can recover child data in other cache registries.
func (c SimpleCacheManager) GetCaches(cacheKeys ...string) (map[string]CacheRegistry, error) {
	mapCR, err := c.CacheStorage.GetValuesMap(cacheKeys...)

	for key, cacheRegistry := range mapCR {

		mapCR[key] = setTTLToPayload(cacheRegistry)

	}

	return mapCR, err
}


//transfer the ttl information from cacheRegistry to paylaod interface, if it is ExposeTTL
func setTTLToPayload(cacheRegistry CacheRegistry) CacheRegistry {

	if (cacheRegistry.Payload == nil) {
		return cacheRegistry
	}

	payload := cacheRegistry.Payload

	exposeTTL, isExposeTTL := payload.(ExposeTTL)




	if isExposeTTL {

		ttl := cacheRegistry.GetTTLSeconds()
		log.Debug("Setting ttl to %v, ttl value %v", cacheRegistry.CacheKey, ttl)

		payload = exposeTTL.SetTtl(ttl) // assure the same type, from set ttl
	} else {
		cacheRegistry.Payload = payload
		log.Debug("Payload doesn't ExposeTTL %v", cacheRegistry.CacheKey)
	}

	return cacheRegistry
}


