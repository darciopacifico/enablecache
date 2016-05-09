package cache

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"
)

var (
	cacheAreaAuto    = "tests"
	cacheStorageAuto = NewRedisCacheStorage("localhost:6379", "", 8, 200, 2000, cacheAreaAuto, SerializerGOB{})
	cacheManagerAuto = AutoCacheManager{
		cacheStorageAuto,
	}
)

func init() {

	gob.Register(Order{})
	gob.Register(OrderItem{})
	gob.Register(OrderItemAtt{})
	gob.Register(Customer{})
	gob.Register(make(map[string]interface{}))

}

func TestReflexiveRecursiveTeardown(t *testing.T) {
	cacheKey := "Order:13"
	var order = createOrder(13)
	//log.Debug("Order original!====================")
	//spew.Dump(order)
	//cr := CacheRegistry{cacheKey, order, ttlDefaultTestOrder, true}
	cr := CacheRegistry{cacheKey, order, 12001, true, ""}

	setTTLToPayload(&cr)
	err := cacheManagerAuto.SetCache(cr)
	if err != nil {
		t.Fail()
		log.Error("Error trying to save registries", err)
	}

	time.Sleep(time.Millisecond * 100)

	cr2, err2 := cacheManagerAuto.GetCache(cacheKey)
	if err2 != nil {
		log.Error("error trying to retrieve cache reg %v", err2)
		t.Fail()
		return
	}
	order2 := cr2.Payload.(Order)
	//spew.Dump(order2)

	customerTtlBaixo := order2.Customer.SetTtl(888)

	cacheReg := CacheRegistry{"Customer:66", customerTtlBaixo, 6666, true, ""}

	cacheStorageAuto.SetExpireTTL(cacheReg)

	time.Sleep(time.Millisecond * 30)

	cr3, _ := cacheManagerAuto.GetCache(cacheKey)

	order3 := cr3.Payload.(Order)
	//spew.Dump(order3)
	log.Debug("cacheRegistry.ttl %v payload.ttl %v", order3.Ttl)

	/*
		if order3.Ttl > 6666 {
			t.Fail()
			log.Error("expected ttl for order must be less then 6666 seconds %v", order3.Ttl)
		}
	*/
}

func TestReflexiveTypeAssertion(t *testing.T) {

	var order = createOrder(14)

	cr := CacheRegistry{"Order:14", order, ttlDefaultTestOrder, true, ""}

	expectedOrderItemCK := "OrderItem:" + strconv.Itoa(order.Itens[1].Id)

	payload := cr.Payload
	payloadVal := reflect.ValueOf(payload)
	payloadType := payloadVal.Type()

	newValue := reflect.New(payloadType).Elem()
	newValue.Set(payloadVal)

	mapVisits := make(map[string]CacheRegistry)

	_, err := cacheManagerAuto.mapAttributesToCacheKeys(cr.CacheKey, &cr, &mapVisits)

	if err != nil {
		log.Error("Eror trying to teardown dependency tree!", err)
		t.Fail()
	}

	if _, hasVal := mapVisits[expectedOrderItemCK]; !hasVal {
		log.Error("Some expected itens to save was not found!")
		t.Fail()
	}

	for key, _ := range mapVisits {
		fmt.Println(key)
	}

}

func TestSerDeserGOB(t *testing.T) {
	mapStr := make(map[string]interface{})
	mapStr2 := make(map[string]interface{})
	mapStr["someAttr"] = mapStr2
	mapStr2["someCacheKey"] = make(map[string]interface{})

	cacheReg := CacheRegistry{"teste", mapStr, -1, true, ""}

	var destBytes []byte
	bufferE := bytes.NewBuffer(destBytes)
	e := gob.NewEncoder(bufferE)
	e.Encode(cacheReg)
	destBytes = bufferE.Bytes()
	bufferE.Reset()

	reader := bytes.NewReader(destBytes)
	d := gob.NewDecoder(reader)
	var newCacheReg CacheRegistry
	d.Decode(&newCacheReg)

	if newCacheReg.Payload == nil {
		log.Error("Error trying to serialize and deserialize a map! Payload becomes nil after deserialization!")
		t.Fail()
	}

	//fmt.Println("novo bjeto deserializado ", newCacheReg)
}

//test reflexive capabilities
func TestReflexiveCacheKey(t *testing.T) {
	order := createOrder(13)
	value := reflect.ValueOf(&order).Elem()

	strCacheKey, err := GetKey(value)

	if err != nil {
		log.Error("Error trying to retrieve cachekey! %s", err.Error())
		t.Fail()
		return
	}

	if strCacheKey != "Order:13" {
		log.Error("valor da cache key nao esperado!", strCacheKey)
		t.Fail()
	} else {
		log.Debug("cache key of order is %s", strCacheKey)
	}
}

func createOrder(id int) Order {

	customer := Customer{
		66,
		"NomeCliente",
		600,
	}

	order := Order{
		id,
		customer,
		"Order reg for test",
		[]string{},
		make([]OrderItem, 3),
		ttlDefaultTestOrder,
	}

	order.Itens[0] = createOrderItem(order)
	order.Itens[1] = createOrderItem(order)
	order.Itens[2] = createOrderItem(order)

	return order
}

var idOI = 1

func createOrderItem(order Order) OrderItem {

	idOI++

	orderItem := OrderItem{
		idOI,
		"some item description",
		[]OrderItemAtt{
			createOIA(order),
			createOIA(order),
			createOIA(order),
		},
	}
	return orderItem
}

var idOIA = 1

func createOIA(order Order) OrderItemAtt {
	idOIA++
	oia := OrderItemAtt{
		idOIA,
		"chave",
		"valor",
	}

	return oia
}

func (o Order) GetCacheKey() string {
	return reflect.TypeOf(o).Name() + ":" + strconv.Itoa(o.Id)
}

func (c Customer) GetCacheKey() string {
	return reflect.TypeOf(c).Name() + ":" + strconv.Itoa(c.Id)
}

func (i OrderItem) GetCacheKey() string {
	return reflect.TypeOf(i).Name() + ":" + strconv.Itoa(i.Id)
}

func (a OrderItemAtt) GetCacheKey() string {
	return reflect.TypeOf(a).Name() + ":" + strconv.Itoa(a.Id)
	//return ""
}
