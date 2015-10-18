*Enablecache lib.

*Allow to enable cache in almost any golang function easily
    
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

