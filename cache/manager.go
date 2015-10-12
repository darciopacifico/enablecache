package cache

import (
	"math"
)

// CacheRegistry:
// Contains struct payload to be cached and additional information about cache registry
// Cachemanager operation must return cache registry always
type CacheRegistry struct {
	CacheKey string      //unique key in cache
	Payload  interface{} //payload to be cached
	Ttl      int         //time to live
	HasValue bool        //return the presence of cached value on cache storage. Useful in batch operations and to avoid nil errors
}

//Define an basic contract for CacheManager
//A good cache manager implementation must optimize caching operations,
//like to cache child useful objects recursively, detect caching inconsistency, invalidate cache registries
//invalidate cacheKey and its dependencies etc
//Cachemanager get operations must return cache registry always, no matter registry exists or not
type CacheManager interface {

	//set values to cache
	//SetCache(cacheKey string, cacheVal interface{}, ttl int) int
	SetCache(cacheRegistry ...CacheRegistry) error

	//recover value from cache
	//GetCache(cacheKey string) (interface{}, bool, int)
	GetCache(cacheKey string) (CacheRegistry, error)

	//recover value from cache
	//GetCache(cacheKey string) (interface{}, bool, int)
	GetCaches(cacheKey ...string) (map[string]CacheRegistry, error)

	//return time to live of cacheKey
	GetCacheTTL(cacheKey string) (int, error)

	//Invalidate cache registry. Means that this cache registry is not valid or consistent anymore.
	//Do not means that registry was deleted in original data source. This operation don't update parent registries,
	//removing references to this registry. Parent cache search must fail to find this registry,
	//meaning that all cache registry is inconsistent.
	//To represent a delete operation, updating parent registries, use exclusively the DeleteCache operation in UpdaterCacheManager interface.
	Invalidate(cacheKey ...string) error
}

// Define a extension contract to CacheManager, with DeleteCache operation.
// UpdateCacheManager contract can be used, actively, in events of Create, Update and Delete of registries.
// CacheManager can be used only, passively, in Read operations
type UpdaterCacheManager interface {
	CacheManager // import all CacheManager definitions
	//Means that cache registry was deleted on the source, and that parent registries must be updated.
	DeleteCache(cacheRegistry CacheRegistry)
}

//means that some struct will define their own ttl.
//cacheRegistry will copy this ttl definition for cache operations
type ExposeTTL interface { //teste
	GetTtl() int
	SetTtl(int) interface{}
}

//Implementation of this interface means that struct knows how to update his
// relateds Cached, when the struct is deleted or inserted. There is nothing to do when struct is updated
type UpdateCachedRelated interface {
	//who is my parent
	ParentsKey() []string

	//For delete operation
	RemoveFromParent(interface{}) interface{}

	//For insert operations
	InsertInParent(interface{}) interface{}
}

//calculate the minimal ttl
func MinTTL(ttl1 int, ttl2 int) int {
	if ttl1 == TTL_INFINITY {
		return ttl2
	}

	if ttl2 == TTL_INFINITY {
		return ttl1
	}

	ttlFinal := int(math.Min(float64(ttl1), float64(ttl2)))

	return ttlFinal
}
