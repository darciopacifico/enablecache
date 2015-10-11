package cache

import (
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("cache")

// define statistics util for iro
var Sta *Stats = NewStats("iro-cache")

//
type SimpleCacheManager struct {
	Ps CacheStorage
}

//invalidate cache registry
func (c SimpleCacheManager) Invalidate(cacheKeys ...string) error {

	errDel := c.Ps.DeleteValues(cacheKeys...)

	if errDel != nil {
		log.Error("Error trying to delete values from cache %v", errDel)
	}

	return errDel
}

//set cache implementation
func (c SimpleCacheManager) SetCache(cacheRegistry CacheRegistry) error {

	//call cachestorage to store data
	err := c.Ps.SetValues(cacheRegistry)

	return err
}

//return time to live
func (c SimpleCacheManager) GetCacheTTL(cacheKey string) (int, error) {
	return c.Ps.GetTTL(cacheKey)
}

//implement getCache operation that can recover child data in other cache registries.
func (c SimpleCacheManager) GetCache(cacheKey string) (CacheRegistry, error) {

	//get the raw value from cache storage
	//this registry maybe missed some child reference, that will be check some lines below

	cacheRegistries, err := c.Ps.GetValues(cacheKey)
	if err != nil {
		log.Error("Error trying to recover value from cache storage! %s", cacheKey)
		Sta.Miss()
		return CacheRegistry{cacheKey, nil, -2, false}, err
	}
	if len(cacheRegistries) == 0 {
		log.Debug("Cache registry not found! %s", cacheKey)
		Sta.Miss()
		return CacheRegistry{cacheKey, nil, -2, false}, nil
	}

	cacheRegistry := cacheRegistries[cacheKey]

	//cache miss for raw cache value!
	if !cacheRegistry.HasValue {
		Sta.Miss()
		return cacheRegistry, nil // empty, hasValue=false, cacheRegistry
	}

	//return final cache registry
	Sta.Hit()
	return cacheRegistry, nil

}
