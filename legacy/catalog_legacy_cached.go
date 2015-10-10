package legacy

import (
	"gitlab.wmxp.com.br/bis/biro/aop"
	"gitlab.wmxp.com.br/bis/biro/cache"
	"gitlab.wmxp.com.br/bis/biro/rest"
	"gitlab.wmxp.com.br/bis/biro/schema"
	"gitlab.wmxp.com.br/bis/biro/utilsbiro"
	"os"
	"reflect"
)

//declare function types
type TypeFindProductById func(id int) (schema.ProductLegacy, bool, error)
type TypeFindSKUById func(id int) (schema.SkuLegacy, bool, error)
type TypeFindOfferById func(id int) (schema.OfferLegacy, bool, error)
type TypeFindProductNoOfferById func(id int) (schema.ProductLegacy, bool, error)

//declare function variables
var (
	//declare cached functions

	CachedFindProductById        TypeFindProductById
	CachedFindSKUById            TypeFindSKUById
	CachedFindOfferById          TypeFindOfferById
	CachedFindProductNoOfferById TypeFindProductNoOfferById

	NoCacheFindProductById        TypeFindProductById
	NoCacheFindSKUById            TypeFindSKUById
	NoCacheFindOfferById          TypeFindOfferById
	NoCacheFindProductNoOfferById TypeFindProductNoOfferById

	//isntantiate the cache storage redis
	biroRedisCacheStorage = cache.NewRedisCacheStorage(
		conf.Config("ipPortRedis", "localhost:6379"),
		conf.Config("passwordRedis", ""),
		conf.ConfigInt("maxIdle", "8"),
		"legacyrepo",
	)

	cacheImplementation = conf.Config("CacheImplementation", "simple")

	BiroCacheManager cache.CacheManager
)

func init() {

	if cacheImplementation == "simple" {
		log.Warning("Loading simple cache implementation!")
		BiroCacheManager = cache.SimpleCacheManager{
			Ps: biroRedisCacheStorage,
		}

	} else if cacheImplementation == "auto" {
		log.Warning("Loading automatic cache implementation!")
		BiroCacheManager = cache.AutoCacheManager{
			Ps: biroRedisCacheStorage,
		}
	} else {
		log.Error("Cache implementation '%v' not recognized!", cacheImplementation)
		os.Exit(1)

	}

	SvcAngus := CatalogLegacy{
		rest.GenericRESTClient{
			HttpCaller: rest.ExecuteRequestHot,
			//HttpCaller: pool.ExecuteRequestPool,
		},
	}

	//swap all cacheable functions
	//put in the cacheableFunction all the caching mechanism
	aop.MakeSwap(&CachedFindProductById, SvcAngus.FindProductById, BiroCacheManager, true)
	aop.MakeSwap(&CachedFindSKUById, SvcAngus.FindSKUById, BiroCacheManager, true)
	aop.MakeSwap(&CachedFindOfferById, SvcAngus.FindOfferById, BiroCacheManager, true)
	aop.MakeSwap(&CachedFindProductNoOfferById, SvcAngus.FindProductNoOfferById, BISCacheManager, true)

	//put in the noCacheFunction all the caching mechanism, just for search in the original function and cache results
	aop.MakeSwap(&NoCacheFindProductById, SvcAngus.FindProductById, BiroCacheManager, false)
	aop.MakeSwap(&NoCacheFindSKUById, SvcAngus.FindSKUById, BiroCacheManager, false)
	aop.MakeSwap(&NoCacheFindOfferById, SvcAngus.FindOfferById, BiroCacheManager, false)
	aop.MakeSwap(&NoCacheFindProductNoOfferById, SvcAngus.FindProductNoOfferById, BISCacheManager, false)

}

/**FUNCTIONS FOR PRODUCT FIND**/
func (TypeFindProductById) IsValidResults(in []reflect.Value, out []reflect.Value) bool {
	return utilsbiro.IsValidResults_ForFindBiroLegacy(in, out)
}

/**FUNCTIONS FOR SKU FIND*/

func (TypeFindSKUById) IsValidResults(in []reflect.Value, out []reflect.Value) bool {
	return utilsbiro.IsValidResults_ForFindBiroLegacy(in, out)
}

/*FUNCTIONS FOR OFFER FIND*/
func (TypeFindOfferById) IsValidResults(in []reflect.Value, out []reflect.Value) bool {
	return utilsbiro.IsValidResults_ForFindBiroLegacy(in, out)
}

func FindProductNoOfferByIdCache(id int, cache bool) (schema.ProductLegacy, bool, error) {
	if cache {
		return CachedFindProductNoOfferById(id)
	} else {
		return NoCacheFindProductNoOfferById(id)
	}
}
