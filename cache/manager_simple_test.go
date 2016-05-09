package cache

import (
	"testing"
	"time"

	"math/rand"
	"strconv"
)

const cacheKeyForTest = "Order:1"

var (
	ttlDefaultTestOrder     = 1200
	ttlDefaultTestOrderItem = 600
	cacheArea               = strconv.Itoa(rand.Int())
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
	Ttl      int
}

func (c Order) GetTtl() int {
	return c.Ttl
}

func (c Order) SetTtl(ttl int) interface{} {
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

	time.Sleep(time.Millisecond * 30)
}
