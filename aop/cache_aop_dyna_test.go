package aop

import (
	"encoding/gob"
	"reflect"
	"strconv"
	"testing"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"fmt"
	"github.com/darciopacifico/cachengo/cache"
	"github.com/op/go-logging"
	"os"
)

var (
	cacheAreaAuto     = "dyna_test"
	cacheStorageRedis = cache.NewRedisCacheStorage("localhost:6379", "", 8, cacheAreaAuto)
	cmAuto            = cache.SimpleCacheManager{
		cacheStorageRedis,
	}
)

func init() {

	format := logging.MustStringFormatter("%{color}%{time:15:04:05.000} PID:%{pid} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}")
	backend1 := logging.NewLogBackend(os.Stdout, "", 0)
	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	logging.SetBackend(backend1Formatter)

	logging.SetLevel(logging.DEBUG, "cache")

}

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

	CacheSpot{CachedFunc: &cFindOne, HotFunc: FindOneCustomer, CacheManager: cmAuto}.StartCache()
	CacheSpot{CachedFunc: &cFindMany, HotFunc: FindManyCustomers, CacheManager: cmAuto}.StartCache()
	CacheSpot{CachedFunc: &cFindOneB, HotFunc: FindManyCustomers, CacheManager: cmAuto}.StartCache()
	CacheSpot{CachedFunc: &cFindManyB, HotFunc: FindOneCustomer, CacheManager: cmAuto}.StartCache()
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

func TestAllSwaps(t *testing.T) {
	var f bool

	format := logging.MustStringFormatter("%{color}%{time:15:04:05.000} PID:%{pid} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}")
	backend1 := logging.NewLogBackend(os.Stdout, "", 0)
	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	logging.SetBackend(backend1Formatter)

	logging.SetLevel(logging.DEBUG, "xxx")

	log := *logging.MustGetLogger("xxx")

	log.Debug("teste 23")

	ps, errM := cFindMany([]int{1, 2, 3})
	fmt.Sprintf("procurando pessoas %v err = %v", ps, errM)

	ps, errM = cFindManyB([]int{2, 3, 4, 5})
	fmt.Sprintf("procurando pessoas %v err = %v", ps, errM)

	p, err, f := cFindOne(2)
	fmt.Sprintf("procurando pessoa %v err = %v, f=%v", p, err, f)

	p, err, f = cFindOne(6)
	fmt.Sprintf("procurando pessoa %v err = %v, f=%v", p, err, f)

	p, err, f = cFindOneB(4)
	fmt.Sprintf("procurando pessoa %v err = %v, f=%v", p, err, f)

	p, err, f = cFindOneB(9)
	fmt.Sprintf("procurando pessoa %v err = %v, f=%v", p, err, f)

}
