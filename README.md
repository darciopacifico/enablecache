#EnableCache lib. [![GoDoc](https://godoc.org/github.com/darciopacifico/enablecache?status.svg)](https://godoc.org/github.com/darciopacifico/enablecache) [![Build Status](https://travis-ci.org/darciopacifico/enablecache.svg?branch=master)](https://travis-ci.org/darciopacifico/enablecache)
Allow to enable cache in almost any golang function easily.

### Minimum example
```go
package main

import (
	"fmt"
	"strconv"
	
	"github.com/darciopacifico/enablecache/aop"
	"github.com/darciopacifico/enablecache/cache"
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
		CacheStorage: cache.NewRedisCacheStorage("localhost:6379", "", 8, "lab"),
	}

	//start cache spot reference.
	cacheSpot = aop.CacheSpot{
		HotFunc: FindProduct,		// concrete FindProduct function
		CachedFunc: &CachedFindProduct, // Empty cached function as ref. Will receive a swap function
		CacheManager: cacheManager,	// Cache Manager implementation
	}.MustStartCache()			// Validate function signatures, assoaciate swap to CachedFunc
}

func main() {
	//call new cached find product as usually call original FindProduct
	fmt.Println(CachedFindProduct(9))

	//Cache storage operations is started in separateds go routines.
	//A waiting group ensure for all operations to finish.
	cacheSpot.WaitAllParallelOps()
}
```
- It's important to call `cacheSpot.MustStartCache()` at an `func init(){...}`. It's need to fail at startup if some cache config goes wrong!

- Check your Redis registries after. Some new keys was stored.

- Call `CachedFindProduct` many times and note that the fake "expensive operation" will not be called anymore, until cache expires.

- Allways call `cacheSpot.WaitAllParallelOps()` at the end of yor program, or when need to sincronize pending store operations.

### Used in Production 
Currently in production in a big retailer e-commerce environment ;-)

### Performance
- Proved performance for almost 300 simultaneous requests per 1Gb RAM and 1 CPU Core. No leaks, minimum CPU overhead.

### Detailed function
- Independent and cohesive layers, with well defined interfaces.
	- Cache Spot: AOP like instrumentation, allowing almost any golang function to be transparently cached.
	- Cache Manager: implements cache split algorithm.
	- Cache Storage: interact with an external big memory layer (Redis).

### License
- Enablecache is free software, licensed under the Apache License, Version 2.0 (the "License"). Commercial and non-commercial use are permitted in compliance with the License.
