package aop
import "fmt"

//Wait for all store goroutines.
func (cacheSpot CacheSpot) WaitAllParallelOps() {
	log.Debug("Waiting for parallel operations...")
	fmt.Println("Waiting for parallel operations...")
	cacheSpot.wg.Wait()
}
