package cache

import (
	"testing"
	"time"
)

var (
	cacheStorageRedis_test = NewRedisCacheStorage("localhost:6379", "", 8, "cachetest")
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
		CacheRegistry{"valtestdel1", "valor1", -1, true},
		CacheRegistry{"valtestdel2", "valor2", -1, true},
		CacheRegistry{"valtestdel3", "valor3", -1, true})

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
	/*
	 */
}

func TestSetTTL(t *testing.T) {
	cacheKey := "testSetTTL"
	ttl := 100

	cacheStorageRedis_test.SetValues(CacheRegistry{
		cacheKey,
		"some val",
		ttl,
		true,
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

	if cacheReg.Ttl >= ttl {
		log.Error("TTL setting was not updated as return val! %v, %v", cacheReg.Ttl, ttl)
		t.Fail()
	} else {
		log.Debug("TTL setting was updated in return val! %v, %v", cacheReg.Ttl, ttl)
	}
}
