package cache

//Basic contract for cache storage. Define basic key/value operations
//could be implemented with any kind of key/value persistence mechanism
//a good cachestorage implementation can store and recover data efficiently, like batch recover,
//using an go routine to store data, maybe how to inherity and share cache areas etc
type CacheStorage interface {
	//include cache registries
	SetValues(values ...CacheRegistry) error

	//recover cache values (map of values, map of hasValue bool, map of ttls, error)
	//GetValues(keys ...string) (map[string]interface{}, map[string]bool, map[string]int, error)
	GetValuesMap(keys ...string) (map[string]CacheRegistry, error)

	//delete cache values
	DeleteValues(cacheKey ...string) error
}

