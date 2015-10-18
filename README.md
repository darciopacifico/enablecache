#Enablecache lib.
================

## Allow to enable cache in almost any golang function easily

### Minimum example
    
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

    func main() {
    
    	//cache manager that will intermediate the operations
    	cacheManager := cache.SimpleCacheManager{
    		CacheStorage: cache.NewRedisCacheStorage("localhost:6379", "", 8, "lab"),
    	}
    
    	//empty function that will receive cache spot, with same signature of FindProduct
    	var CachedFindProduct func(id int) string
    
    	//start cache spot reference.
    	cacheSpot := aop.CacheSpot{CachedFunc: &CachedFindProduct, HotFunc: FindProduct, CacheManager: cacheManager}.StartCache()
    
    
    	//call cached find product
    	fmt.Println(CachedFindProduct(9))
    
    
    	//cache storage is started in a separated go routine.
    	//Wait for finish
    	cacheSpot.WaitAllParallelOps()
    }

### Used in Production 
Currently in production in a big retailer e-commerce environment.

### Performance
Proved performance for almost 300 simultaneous requests per 1Gb RAM / 1 CPU Core. No leaks, very low processor overhead.

### Detailed function
Independent and cohesive layers, with well defined interfaces.
- Cache Spot: AOP like instrumentation, allowing almost any golang function to be transparently cached.
- Cache Manager: implements cache split algorithm.
- Cache Storage: interact with an external big memory layer (Redis).

### License
Enablecache is free software, licensed under the Apache License, Version 2.0 (the "License"). Commercial and non-commercial use are permitted in compliance with the License.
