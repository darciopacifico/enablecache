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
	"sync"
)

var (
	cacheAreaAuto = "dyna_test:" + strconv.Itoa(int(time.Now().Unix()))
	cacheStorageRedis = cache.NewRedisCacheStorage("localhost:6379", "", 8, 200, 2000, cacheAreaAuto, cache.SerializerGOB{}, true)
	cmAuto = cache.SimpleCacheManager{
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
		customers, err = cFindManyB_many_one([]int{6, 7, 8})

	}

	log.Debug("%v, %v ", customers, err)

	time.Sleep(time.Millisecond * 1)

}

var cFindOne_one_one func(int) (People, error, bool)
var cFindOneB_one_many func(int) (People, error, bool)
var cFindMany_many_many func([]int) ([]People, error)
var cFindManyB_many_one func([]int) ([]People, error)

func init() {
	gob.Register(People{})

	//one to one
	CacheSpot{
		CachedFunc:   &cFindOne_one_one,
		HotFunc:      FindOneCustomer,
		CacheManager: cmAuto,
		Ttl: 	time.Second*100,
		WaitingGroup: &sync.WaitGroup{},
		//ValidateResults: func (allIns []reflect.Value, allOuts []reflect.Value, cacheKey string, singleValueToCache interface{}) bool{return true},
		//SpecifyOutputKeys:KeysForCache,
	}.MustStartCache()

	//one to many
	CacheSpot{
		CachedFunc:   &cFindOneB_one_many,
		HotFunc:      FindManyCustomers,
		Ttl: 	time.Second*100,
		CacheManager: cmAuto,
		WaitingGroup: &sync.WaitGroup{},
		//ValidateResults:IsValidResults,
		//SpecifyOutputKeys:KeysForCache,
	}.MustStartCache()

	//many to many
	CacheSpot{CachedFunc: &cFindMany_many_many,
		HotFunc:      FindManyCustomers,
		Ttl: 	time.Second*100,
		CacheManager: cmAuto,
		WaitingGroup: &sync.WaitGroup{},
		//ValidateResults:ManyIsValidResults,
		//SpecifyOutputKeys:ManyKeysForCache,
	}.MustStartCache()


	//many to one
	CacheSpot{CachedFunc: &cFindManyB_many_one,
		HotFunc:      FindOneCustomer,
		Ttl: 	time.Second*100,
		CacheManager: cmAuto,
		WaitingGroup: &sync.WaitGroup{},
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

	ps, errM := cFindMany_many_many([]int{1, 2, 3})

	fmt.Printf("procurando pessoas %v err = %v  \n", ps, errM)

	ps, errM = cFindManyB_many_one([]int{2, 3, 4, 5})
	fmt.Printf("procurando pessoas %v err = %v \n", ps, errM)

	ps, errM = cFindManyB_many_one([]int{4, 5, 10, 11})
	fmt.Printf("procurando pessoas %v err = %v \n", ps, errM)

	p, err, f := cFindOne_one_one(2)
	fmt.Printf("procurando pessoa %v err = %v, f=%v \n", p, err, f)

	p, err, f = cFindOne_one_one(6)
	fmt.Printf("procurando pessoa %v err = %v, f=%v \n", p, err, f)

	p, err, f = cFindOneB_one_many(4)
	fmt.Printf("procurando pessoa %v err = %v, f=%v \n", p, err, f)

	p, err, f = cFindOneB_one_many(12)
	fmt.Printf("procurando pessoa %v err = %v, f=%v \n", p, err, f)

	p, err, f = cFindOneB_one_many(9)
	fmt.Printf("procurando pessoa %v err = %v, f=%v \n", p, err, f)

}
