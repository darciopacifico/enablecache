//documentation for package sample

//package sample from sample.go from sample.go
package main

import (
	"fmt"
	"strconv"

	"github.com/darciopacifico/enablecache/aop"
	"github.com/darciopacifico/enablecache/cache"
	"sync"
)

//concrete no cached function
func FindProduct(id int) string {
	fmt.Println("calling a very expensive function...")
	return "product:" + strconv.Itoa(id)
}

//empty function, currently pointing to nil, that will receive cache spot, with same signature of FindProduct
var CachedFindProduct func(id int) string

//cache spot configuration
var cacheSpot aop.CacheSpot

//initialize cache
func init(){
	//cache manager that will intermediate all operations for cache store/read.
	cacheManager := cache.SimpleCacheManager{
		CacheStorage: cache.NewRedisCacheStorage("localhost:6379", "", 8, 200, 3000, "lab", cache.SerializerGOB{}),
	}

	//start cache spot reference.
	cacheSpot = aop.CacheSpot{
		HotFunc: FindProduct,       // concrete FindProduct function
		CachedFunc: &CachedFindProduct, // Empty cached function as ref. Will receive a swap function
		CacheManager: cacheManager, // Cache Manager implementation
		WaitingGroup: &sync.WaitGroup{},
	}.MustStartCache()          // Validate function signatures, assoaciate swap to CachedFunc
}

func main() {
	//call new cached find product as usually call original FindProduct
	fmt.Println(CachedFindProduct(9))

	//Cache storage operations is started in separateds go routines.
	//A waiting group ensure for all operations to finish.
	cacheSpot.WaitAllParallelOps()
}