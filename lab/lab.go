package main

import (
	"fmt"
	"github.com/darciopacifico/enablecache/aop"
	"github.com/darciopacifico/enablecache/cache"
	"strconv"
)

//concrete no cached function
func FindProduct(id int) string {
	fmt.Println("calling a very expensive function...")
	return "product:" + strconv.Itoa(id)
}

//empty function that will receive cache spot, with same signature of FindProduct
var CachedFindProduct func(id int) string

//cache spot configuration
var cacheSpot aop.CacheSpot

//initialize cache
func init() {
	//cache manager that will intermediate the operations
	cacheManager := cache.SimpleCacheManager{
		CacheStorage: cache.NewRedisCacheStorage("localhost:6379", "", 8, "lab"),
	}

	//start cache spot reference.
	cacheSpot = aop.CacheSpot{
		CachedFunc:   &CachedFindProduct,
		HotFunc:      FindProduct,
		CacheManager: cacheManager,
	}.MustStartCache()
}

func main() {
	//call new cached find product as usually call original FindProduct
	fmt.Println(CachedFindProduct(9))

	//cache storage is started in a separated go routine.
	//Wait for finish
	cacheSpot.WaitAllParallelOps()
}
