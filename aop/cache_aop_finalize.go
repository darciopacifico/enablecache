package aop

//Wait for all store goroutines.
func (cacheSpot CacheSpot) WaitAllParallelOps() {
	cacheSpot.wg.Wait()
}
