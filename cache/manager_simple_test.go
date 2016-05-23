package cache

import (
	"testing"
	"time"

	"strconv"
)

const cacheKeyForTest = "Order:1"

var (
	ttlDefaultTestOrder = 1200.0
	ttlDefaultTestOrderItem = 600.0
	cacheArea = "dyna_test:" + strconv.Itoa(int(time.Now().Unix()))

	//all registries will be deleted, if test success

	cacheStorage = NewRedisCacheStorage("localhost:6379", "", 8, 200, 2000, cacheArea, SerializerGOB{})

	cacheManager = SimpleCacheManager{
		cacheStorage,
	}
)

type Order struct {
	Id       int
	Customer Customer
	OrderObs string
	ItensIds []string
	Itens    []OrderItem
	Ttl      float64
}

func (c Order) GetTtl() float64 {
	return c.Ttl
}

func (c Order) SetTtl(ttl float64) interface{} {
	c.Ttl = ttl
	return c
}

type Customer struct {
	Id   int
	Name string
	Ttl  int
}

func (c Customer) GetTtl() int {
	return c.Ttl
}
func (c Customer) SetTtl(ttl int) interface{} {
	c.Ttl = ttl
	return c
}

/*

type ExposeTTL interface {
	GetTtl() int
	SetTtl(int) interface{}
}
*/

func init() {

	order := createOrder(1)
	cacheRegistry := CacheRegistry{
		cacheKeyForTest,
		order,
		ttlDefaultTestOrderItem,
		time.Now(),
		true,
		"",
	}

	cacheManager.SetCache(cacheRegistry)

	time.Sleep(time.Millisecond * 50) // set cache é uma go routine.. um sleep garante q a mesma execute
}

type OrderItem struct {
	Id       int
	Name     string
	OrderAtt []OrderItemAtt
}

type OrderItemAtt struct {
	Id  int
	Key string
	Val string
}

func TestCacheManager(t *testing.T) {

	//item created previously
	cacheRegistry, _ := cacheManager.GetCache(cacheKeyForTest)
	if cacheRegistry.HasValue {
		log.Debug("valor recuperado ", cacheRegistry.Payload)
	} else {
		log.Error("valor inserido previamente nao encontrado ", cacheRegistry.CacheKey)
		t.Fail()
	}

	//test update
	newCustName := "cliente do pedido atualizado"
	orderUpdated := createOrder(1)

	orderUpdated.OrderObs = newCustName

	newCacheRegistry := CacheRegistry{
		cacheKeyForTest,
		orderUpdated,
		ttlDefaultTestOrder,
		time.Now(),
		true,
		"",
	}
	cacheManager.SetCache(newCacheRegistry)

	//check update payload
	somePayload, _ := cacheManager.GetCache(cacheKeyForTest)
	orderUpRecovered, _ := somePayload.Payload.(Order)

	if !somePayload.HasValue && orderUpRecovered.OrderObs == newCustName {
		log.Error("erro ao tentar atualizar registro do cache!!" + cacheKeyForTest)
		t.Fail()
	} else {
		log.Debug("registro de cache atualizado com sucesso!")
	}

	cacheManager.Invalidate(cacheKeyForTest)

	somePayload2, _ := cacheManager.GetCache(cacheKeyForTest)

	if somePayload2.HasValue || somePayload2.Payload != nil {
		log.Error("Error: Cache registry not invalidated")
		t.Fail()
	} else {
		log.Debug("Cache registry invalidated successfully!")
	}

	time.Sleep(time.Millisecond * 30)
}

func TestMinTTL(t *testing.T) {

	if (MinTTL(-22, 90) != -22) {
		log.Error("Error ttl calculation erro!")
		t.Fail()
	}

	if (MinTTL(TTL_INFINITY, 90) != 90) {
		log.Error("Error ttl calculation erro!")
		t.Fail()
	}

	if (MinTTL(70, TTL_INFINITY) != 70) {
		log.Error("Error ttl calculation erro!")
		t.Fail()
	}

}
