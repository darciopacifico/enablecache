package cache

import (
	"strconv"
	"testing"
)

var simpleCacheManager = SimpleCacheManager{
	Ps: NewRedisCacheStorage("localhost:6379", "", 8, cacheArea),
}

var updaterCacheManager = UpdaterCacheManagerImpl{
	SimpleCacheManager: simpleCacheManager,
}

//prove the hability of UpdaterCacheManager to recycle cache values
func TestReactiveCacheManager(t *testing.T) {

	//simulate some order finding or creation
	order := createOrder(1) // an order, 3 itens, 3 attributes each item

	cp := CacheRegistry{
		"Order:1",
		order,
		-1,
		true,
	}

	simpleCacheManager.SetCache(cp)

	//update some item information, simulation some regular update operation from CRUD interface
	orderItem := order.Itens[1]
	orderItem.Name = "nome atualizado Nao invalidar!!!"

	cpItem := CacheRegistry{
		"OrderItem:" + strconv.Itoa(orderItem.Id),
		orderItem,
		-1,
		true,
	}

	updaterCacheManager.SetCache(cpItem)

	//order 1 must not be invalidated
	//order item must have same new description
	cpWithNewDesc, _ := simpleCacheManager.GetCache("Order:1")

	if !cpWithNewDesc.HasValue {
		t.Fail()
		log.Error("order cache was invalidated/evicted incorrectly")
	}

}
