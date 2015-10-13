package main

import (
	"fmt"
	"github.com/darciopacifico/cachengo/aop"
	"github.com/darciopacifico/cachengo/cache"
	"strconv"
)

var (
	cacheStorage = cache.NewRedisCacheStorage("localhost:6379", "", 8, "test")
	cacheManager = cache.SimpleCacheManager{
		Ps: cacheStorage,
	}
)

func main() {
	aop.CacheSpot{CachedFunc: &CachedFindProduct, HotFunc: FindProduct, CacheManager: cacheManager}.StartCache()
	fmt.Println(CachedFindProduct(2))

}

var CachedFindProduct func(id int) string

func FindProduct(id int) string {
	fmt.Println("calling a very expensive function...")
	return "product:" + strconv.Itoa(id)
}
