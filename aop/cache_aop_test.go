package aop

import (
	"encoding/gob"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"
)

//FindUser
type FindUserType func(id int) User

func (f FindUserType) IsValidResults(in []reflect.Value, out []reflect.Value) bool {
	return true
}

type FindCustomerType func(id int, name string, isActive bool) (Customer, bool, time.Time)

type FindCustomerSimpleType func(id int) Customer

func (f FindCustomerSimpleType) IsValidResults(in []reflect.Value, out []reflect.Value) bool {
	return true
}

func TestAutoValidation(t *testing.T) {
	idTest := 13

	cacheManager.Invalidate("Customer:13")
	time.Sleep(time.Millisecond * 10)

	var cachedFinder FindCustomerType

	MakeSwap(&cachedFinder, FindCustomer, cacheManager, true)

	//first finding, must be cached
	_, _, creationTime := cachedFinder(idTest, "teste", true)

	//wait for a cache storage flush
	time.Sleep(time.Millisecond * 10)

	//must returned same cached val
	cPostCache2, _, _ := cachedFinder(idTest, "teste", true)
	if cPostCache2.Creation != creationTime {
		log.Error("o dado consultado na primeira consulta e o retornado nas demais nao é o mesmo!")
		t.Fail()
	}

	//simulate some invalidation of cache
	cacheStorage.DeleteValues("Customer:" + strconv.Itoa(idTest))

	//wait for a cache storage flush
	time.Sleep(time.Millisecond * 10)

	//must returned same cached val
	cPostCache3, _, creationTimePostInvalidation := cachedFinder(idTest, "teste", true)
	if cPostCache3.Creation == creationTime {
		log.Error("O dado consultado foi invalidado do cache anteriormente! Nao poderia ter sido o mesmo timestamp da primeira consulta!")
		t.Fail()
	}

	//wait for a cache storage flush
	time.Sleep(time.Millisecond * 10)

	//must returned same cached val
	cPostCache4, _, _ := cachedFinder(idTest, "teste", true)
	if cPostCache4.Creation != creationTimePostInvalidation {
		log.Error("O registro de cache falhou!")
		t.Fail()
	}

	cacheManager.Invalidate("Customer:13")

	time.Sleep(time.Millisecond * 10)

}

func TestCustomValidation(t *testing.T) {
	idTest := 99

	var cachedFinderSimple FindCustomerSimpleType

	MakeSwap(&cachedFinderSimple, FindCustomerSimple, cacheManager, true)

	//first finding, must be cached
	cpCache := cachedFinderSimple(idTest)

	//wait for a cache storage flush
	time.Sleep(time.Millisecond * 30)

	//must returned same cached val
	cPostCache2 := cachedFinderSimple(idTest)
	if cPostCache2.Creation != cpCache.Creation {
		log.Error("o dado consultado na primeira consulta e o retornado nas demais nao é o mesmo!")
		t.Fail()
	}

	//must returned same cached val
	cPostCache3 := cachedFinderSimple(idTest)
	if cPostCache3.Creation != cpCache.Creation {
		log.Error("o dado consultado na primeira consulta e o retornado nas demais nao é o mesmo!")
		t.Fail()
	}

	cacheStorage.DeleteValues("Customer:" + strconv.Itoa(idTest))
}

//test the ttl function of cachemanager
func TestTTL(t *testing.T) {
	idUser := 42

	//prepared a cached function, using the original one
	var cachedFindUser FindUserType
	MakeSwap(&cachedFindUser, FindUser, cacheManager, true)

	//first search will be uncached
	user1 := cachedFindUser(idUser)

	//wait for some time, inside the ttl window
	time.Sleep(time.Millisecond * 20) //just wait for storage flush

	//user2 must be the same registry of user1
	user2 := cachedFindUser(idUser)
	if user1.Creation != user2.Creation {
		log.Error("Cache operation failed!")
		t.Fail()
	} else {
		log.Debug("Operations inside TTL window succeed! Cache responds correctly!")
	}

	//wait the end of ttl window
	log.Debug("WAITING by the end of TTL Window (1.2s)...")
	time.Sleep(time.Millisecond * 1200)

	//search for same user again. At this time, the registry must be another one. Can't be the same of user2 or user1
	user3 := cachedFindUser(idUser)
	if user3.Creation == user2.Creation {
		log.Error("Value still cached after the expected ttl time!")
		t.Fail()
	} else {
		log.Debug("Registry gone from cache! Operation succeed!")
	}

	//again, a new searche operatio, inside the new ttl window
	time.Sleep(time.Millisecond * 20) //just wait for storage flush

	//inside ttl window, user4 must be the same as user3
	user4 := cachedFindUser(idUser)
	if user4.Creation != user3.Creation {
		log.Error("Cache operation failed!")
		t.Fail()
	} else {
		log.Debug("New cache operation, inside the new ttl window succeed!")
	}

	//a final sleep to flush data
	time.Sleep(time.Millisecond * 10)
}

func isFunc(myInt interface{}) {
	defValuable, isDefValuable := myInt.(DefaultValubleFunction)

	fmt.Println(defValuable, isDefValuable)
}

func init() {
	gob.Register(Customer{})
	gob.Register(InsurantePolicy{})
	gob.Register(IPItem{})
	gob.Register(User{})
	//gob.Register(cache.DefineTTLGeneric{})

}
