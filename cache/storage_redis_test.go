package cache

import (
	"testing"
	"time"
	"github.com/garyburd/redigo/redis"
	"reflect"
)

var (
	_=redis.ErrNil
	cacheStorageRedis_test = NewRedisCacheStorage("localhost:6379", "", 8, 200, 2000, "cachetest", SerializerGOB{})
	qtdChaves = 100
)

func TestDelete(t *testing.T) {
	//cleanup repository
	cacheStorageRedis_test.DeleteValues("valtestdel1", "valtestdel2", "valtestdel3")

	cp, err := cacheStorageRedis_test.GetValuesMap("valtestdel1", "valtestdel2", "valtestdel3")

	if err != nil {
		log.Error("Erro ao tentar consultar cache storage", err)
		t.Fail()
	}

	if cp["valtestdel3"].HasValue {
		log.Error("de teste encontrado!! o teste falhou anteriormente??")
		t.Fail()
	}

	//insert some registries
	cacheStorageRedis_test.SetValues(
		CacheRegistry{"valtestdel1", "valor1", -1, time.Now(), true,""},
		CacheRegistry{"valtestdel2", "valor2", -1, time.Now(), true,""},
		CacheRegistry{"valtestdel3", "valor3", -1, time.Now(), true,""})

	//check insertion
	cpsNew, err := cacheStorageRedis_test.GetValuesMap("valtestdel1", "valtestdel2", "valtestdel3")

	if err != nil {
		log.Error("erro ao tentar consultar cache storage!", err)
		t.Fail()
	}

	if !cpsNew["valtestdel3"].HasValue {
		log.Error("valor nao encontrado para valtestdel3")
		t.Fail()
	} else {
		log.Debug("OK! dados inseridos!")
	}

	//delete values
	cacheStorageRedis_test.DeleteValues("valtestdel1", "valtestdel2", "valtestdel3")

	//check deletion
	cpCheckDel, err := cacheStorageRedis_test.GetValuesMap("valtestdel3")
	if err != nil {
		log.Error("erro ao tentar consultar cache storage!", err)
		t.Fail()
	}

	if cpCheckDel["valtestdel3"].HasValue {
		log.Error("valor encontrado! A delecao falhou! valtestdel3")
		t.Fail()
	} else {
		log.Debug("OK! dados deletados!")
	}
}

func TestSetTTL(t *testing.T) {
	cacheKey := "testSetTTL"
	ttl := 1.0

	cacheStorageRedis_test.SetValues(CacheRegistry{
		cacheKey,
		"some val",
		ttl,
		time.Now(),
		true,
		"",
	})

	log.Debug("Waiting for 2 seconds to test ttl update at get operation!")
	time.Sleep(time.Second * 2)

	cacheRegs, err := cacheStorageRedis_test.GetValuesMap(cacheKey)
	if err != nil {
		log.Error("Erro ao tentar recuperar cache registry!")
		t.Fail()
		return
	}

	cacheReg := cacheRegs[cacheKey]


	if cacheReg.GetTTLSeconds() >= ttl {
		log.Error("TTL setting was not updated as return val! %v, %v", cacheReg.GetTTLSeconds(), ttl)
		t.Fail()
	} else {
		log.Debug("TTL setting was updated in return val! %v, %v",  cacheReg.GetTTLSeconds(), ttl)
	}
}


func TestSetGet(t *testing.T) {
	cacheKey := "testSetGet_order1234"
	ttl := 3.0

	orderOrig := createOrder(1234)

	cacheStorageRedis_test.SetValues(CacheRegistry{
		cacheKey,
		orderOrig,
		ttl,
		time.Now(),
		true,
		"",
	})

	log.Debug("Waiting for 2 seconds to test ttl update at get operation!")
	time.Sleep(time.Second * 2)

	cacheRegs, err := cacheStorageRedis_test.GetValuesMap(cacheKey)
	if err != nil {
		log.Error("Erro ao tentar recuperar cache registry!")
		t.Fail()
		return
	}




	cacheReg := cacheRegs[cacheKey]

	if(cacheReg.HasValue){

		orderCast, _ := cacheReg.Payload.(Order)

		log.Debug("order casted, customer name %v", orderCast.Customer.Name)

		log.Debug("OK, key     returned %v ",cacheReg.CacheKey)
		log.Debug("OK, payload returned %v ",cacheReg.Payload)

		log.Debug("OK, type name  returned %v ",reflect.TypeOf(cacheReg.Payload).Name())

		log.Debug("OK, orig %v ",orderOrig)
		log.Debug("OK, orig type name %v ",reflect.TypeOf(orderOrig).Name())

	}else{
		log.Error("Paylod = null! Not expected!")
		t.Fail()
	}
}


func TestSetGetTTL(t *testing.T) {
	cacheKey := "testSetGet_order1234"
	ttl := 1.0

	orderOrig := createOrder(1234)

	cacheStorageRedis_test.SetValues(CacheRegistry{
		cacheKey,
		orderOrig,
		ttl,
		time.Now(),
		true,
		"",
	})

	log.Debug("Waiting for 2 seconds to test ttl update at get operation!")
	time.Sleep(time.Second * 2)

	cacheRegs, err := cacheStorageRedis_test.GetValuesMap(cacheKey)
	if err != nil {
		log.Error("Erro ao tentar recuperar cache registry!")
		t.Fail()
		return
	}

	cacheReg := cacheRegs[cacheKey]
	if(cacheReg.HasValue || cacheReg.Payload!=nil){
		log.Error("Cached value should be null! TTL should invalidate registry!")
		t.Fail()
	}else{
		log.Debug("Value not found on cached as expected!")
	}
}
