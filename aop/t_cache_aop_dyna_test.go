package aop

import (
	"encoding/gob"
	"strconv"
	"testing"
	"time"

	"fmt"
	"github.com/darciopacifico/enablecache/cache"
	"github.com/op/go-logging"
	"math/rand"
	"os"
)

var (
	cacheAreaAuto     = "dyna_test"
	cacheStorageRedis = cache.NewRedisCacheStorage("localhost:6379", "", 8, 200, 2000, cacheAreaAuto, cache.SerializerGOB{}, true)
	cmAuto            = cache.SimpleCacheManager{
		CacheStorage: cacheStorageRedis,
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
	Id    int
	Name  string
	Email string
	Uuid  string
}

func (p People) GetCacheKey() string {
	return "People:" + strconv.Itoa(p.Id)
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

var cFindOne func(int) (People, error, bool)
var cFindOneB func(int) (People, error, bool)
var cFindMany func([]int) ([]People, error)
var cFindManyB func([]int) ([]People, error)

func init() {
	gob.Register(People{})

	CacheSpot{
		CachedFunc:   &cFindOne,
		HotFunc:      FindOneCustomer,
		CacheManager: cmAuto,
		//ValidateResults: func (allIns []reflect.Value, allOuts []reflect.Value, cacheKey string, singleValueToCache interface{}) bool{return true},
		//SpecifyOutputKeys:KeysForCache,
	}.MustStartCache()

	CacheSpot{
		CachedFunc:   &cFindOneB,
		HotFunc:      FindManyCustomers,
		CacheManager: cmAuto,
		//ValidateResults:IsValidResults,
		//SpecifyOutputKeys:KeysForCache,
	}.MustStartCache()

	CacheSpot{CachedFunc: &cFindMany,
		HotFunc:      FindManyCustomers,
		CacheManager: cmAuto,
		//ValidateResults:ManyIsValidResults,
		//SpecifyOutputKeys:ManyKeysForCache,
	}.MustStartCache()

	CacheSpot{CachedFunc: &cFindManyB,
		HotFunc:      FindOneCustomer,
		CacheManager: cmAuto,
		//ValidateResults:ManyIsValidResults,
		//SpecifyOutputKeys:ManyKeysForCache,
	}.MustStartCache()
}

func FindOneCustomer(id int) (People, error, bool) {
	return People{Id: id, Name: "Some name ", Uuid: "randon" + strconv.Itoa(rand.Int())}, nil, true
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

	fmt.Printf("procurando pessoas %v err = %v  \n", ps, errM)

	ps, errM = cFindManyB([]int{2, 3, 4, 5})
	fmt.Printf("procurando pessoas %v err = %v \n", ps, errM)

	p, err, f := cFindOne(2)
	fmt.Printf("procurando pessoa %v err = %v, f=%v \n", p, err, f)

	p, err, f = cFindOne(6)
	fmt.Printf("procurando pessoa %v err = %v, f=%v \n", p, err, f)

	p, err, f = cFindOneB(4)
	fmt.Printf("procurando pessoa %v err = %v, f=%v \n", p, err, f)

	p, err, f = cFindOneB(9)
	fmt.Printf("procurando pessoa %v err = %v, f=%v \n", p, err, f)

}
