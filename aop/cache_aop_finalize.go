package aop

//Wait for all store goroutines.
func (cacheSpot CacheSpot) WaitAllParallelOps() {

	log.Debug("Waiting for parallel operations...")
	if(cacheSpot.WaitingGroup !=nil){
		cacheSpot.WaitingGroup.Wait()
	}else{
		log.Error("Warning! cacheSpot.waitingGroup is null!")
	}

}
