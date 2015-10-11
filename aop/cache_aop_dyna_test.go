package aop

import (
	"encoding/gob"
	"reflect"
	"strconv"
	"testing"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/darciopacifico/cachengo/cache"
)

var (
	cacheAreaAuto     = "dyna_test"
	cacheStorageRedis = cache.NewRedisCacheStorage("localhost:6379", "", 8, cacheAreaAuto)
	cacheManagerAuto  = cache.SimpleCacheManager{
		cacheStorageRedis,
	}
)

type People struct {
	Id   int
	Name string
	Uuid string
}

type FindOneType func(int) (People, error, bool)
type FindManyType func([]int) ([]People, error)

func (FindOneType) IsValidResults(in []reflect.Value, out []reflect.Value) bool {
	return true
}

func (FindOneType) KeysForCache(outs []reflect.Value) ([]string, []reflect.Value) {

	numOuts := len(outs)

	keysToCache := []string{}
	outsToCache := []reflect.Value{}

	for i := 0; i < numOuts; i++ {
		out := outs[i]
		realVal := out.Interface()
		customer, isCustomer := realVal.(People)

		if isCustomer {
			keysToCache = append(keysToCache, strconv.Itoa(customer.Id))
			outsToCache = append(outsToCache, out)

		} else {
			log.Warning("The object %v is not a customer!")
		}
	}

	return keysToCache, outsToCache

}

func (FindManyType) IsValidResults(in []reflect.Value, out []reflect.Value) bool {
	return true
}

func (FindManyType) KeysForCache(outs []reflect.Value) ([]string, []reflect.Value) {

	numOuts := len(outs)

	keysToCache := []string{}
	outsToCache := []reflect.Value{}

	for i := 0; i < numOuts; i++ {
		out := outs[i]
		realVal := out.Interface()
		customer, isCustomer := realVal.(People)

		if isCustomer {
			keysToCache = append(keysToCache, "Customer:"+strconv.Itoa(customer.Id))
			outsToCache = append(outsToCache, out)

		} else {
			log.Warning("The object %v is not a customer!")
		}
	}

	return keysToCache, outsToCache

}

func BenchmarkDynaSwap(b *testing.B) {

	var customers []People
	var err error

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//customers, err = FindManyCustomers([]int{6, 7, 8})
		customers, err = cFindManyB([]int{6, 7, 8})

	}

	log.Debug("%v, %v ", customers, err)

	time.Sleep(time.Millisecond * 1)

}

var cFindOne FindOneType
var cFindOneB FindOneType

var cFindMany FindManyType
var cFindManyB FindManyType

func init() {
	gob.Register(People{})

	MakeSwap(&cFindOne, FindOneCustomer, cacheManagerAuto, true)
	MakeSwap(&cFindMany, FindManyCustomers, cacheManagerAuto, true)

	MakeSwap(&cFindOneB, FindManyCustomers, cacheManagerAuto, true)
	MakeSwap(&cFindManyB, FindOneCustomer, cacheManagerAuto, true)
}

func FindOneCustomer(id int) (People, error, bool) {
	return People{Id: id, Name: "Some name ", Uuid: uuid.New()}, nil, true
}

func FindManyCustomers(ids []int) ([]People, error) {
	cs := make([]People, len(ids))
	for idx, id := range ids {
		cs[idx], _, _ = FindOneCustomer(id)
	}
	return cs, nil
}
