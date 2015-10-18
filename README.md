#EnableCache lib.
================

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
	//cache manager that will intermediate the store and read cache operations
	cacheManager := cache.SimpleCacheManager{
		CacheStorage: cache.NewRedisCacheStorage("localhost:6379", "", 8, "lab"),
	}

	//start cache spot reference.
	cacheSpot = aop.CacheSpot{
		CachedFunc: &CachedFindProduct,
		HotFunc: FindProduct,
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
```
- It's important to call `cacheSpot.MustStartCache()` at an `func init(){...}`. It's need to fail at startup if some cache config goes wrong!

- Check your Redis registries after. Some new keys was stored.

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
