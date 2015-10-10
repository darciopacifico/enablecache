package schema

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/garyburd/redigo/redis"
	"github.com/op/go-logging"
	"io/ioutil"
	"sort"
)

var log = logging.MustGetLogger("biro")

// JSon test files "containers"
var itemProdSkus, itemRespDeactivated, productResp,
	skusResp, itemResp, offerResp, itemQaMeia string

func init() {
	itemProdSkus = ReadFile("../test/t_itemProdSkus_test.json")
	itemRespDeactivated = ReadFile("../test/t_itemRespDeactivated_test.json")
	productResp = ReadFile("../test/t_productResp_test.json")
	skusResp = ReadFile("../test/t_skuResp_test.json")
	itemResp = ReadFile("../test/t_itemResp_test.json")
	offerResp = ReadFile("../test/t_offerResp_test.json")
	itemQaMeia = ReadFile("../test/t_itemQaMeia_test.json")
}

func ReadFile(file string) string {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return string(content)
}

func createFullProduct() ProductLegacy {
	p := ProductLegacy{}

	s := SkuLegacy{}
	s1 := SkuLegacy{}
	s2 := SkuLegacy{}

	o := OfferLegacy{}
	o2 := OfferLegacy{}

	p.Id = 11111111
	s.Id = 12222222
	s1.Id = 22222222
	s2.Id = 32222222
	o.Id = 13333333
	o2.Id = 23333333

	p.Status = true
	s.Status = true
	s1.Status = false
	s2.Status = false
	o.Status = true
	o2.Status = false

	s.OfferList.Offer = append(s.OfferList.Offer, o)
	s.OfferList.Offer = append(s.OfferList.Offer, o2)

	p.Skus = append(p.Skus, s)
	p.Skus = append(p.Skus, s1)
	p.Skus = append(p.Skus, s2)

	return p

}

func TestSchemaValidationAllValid(t *testing.T) {
	p := createFullProduct()

	err := p.IsValid()

	if err != nil {
		t.Fail()
		log.Error("Product template informed must be valid! %s", err.Error())
	}
}

func TestSchemaValidationProductInvalidStatus(t *testing.T) {
	p := createFullProduct()
	p.Status = false

	err := p.IsValid()

	if err == nil {
		t.Fail()
		log.Error("Product must not be valid at this test! %s", err.Error())
	}
}

func TestSchemaValidationSKUNoOffer(t *testing.T) {

	p := createFullProduct()

	p.Skus[0].OfferList.Offer = []OfferLegacy{} // set empty array

	err := p.IsValid()

	if err != nil {
		t.Fail()
		log.Error("Product must not be valid at this test! %s", err.Error())
	}
}

func TestSchemaValidationOfferInvalid(t *testing.T) {

	p := createFullProduct()

	p.Skus[0].OfferList.Offer[0].Status = false

	err := p.IsValid()

	if err != nil {
		t.Fail()
		log.Error("Product must not be valid at this test!", err.Error())
	}

}

func BenchmarkRegisterCost(b *testing.B) {
	res := &ProductLegacy{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gob.Register(res)
	}
}

func TestProductJson(t *testing.T) {

	res := &ProductLegacy{}
	json.Unmarshal([]byte(productResp), res)
	//	fmt.Println("Product Struct", *res)
	//	fmt.Println("Brand", res.Brand)
	//	fmt.Println("Skus", res.Skus)
	//	jsonBack, _ := json.MarshalIndent(*res, "", "    ")
	//	fmt.Println("Backto Json:", string(jsonBack))

}

func TestOfferJson(t *testing.T) {

	res := &OfferLegacy{}
	json.Unmarshal([]byte(offerResp), res)
	//	fmt.Println("Product Struct", *res)
	//	fmt.Println("Offer: offer_sku", res.Sku)
	//	fmt.Println("Offer: sku_product", res.Sku.Product)
	//	jsonBack, _ := json.MarshalIndent(*res, "", "    ")
	//	fmt.Println( /*"OfferLegacy Json:",*/ string(jsonBack))

}

func TestSkuJson(t *testing.T) {
	res := &SkuLegacy{}
	json.Unmarshal([]byte(skusResp), res)

	//	fmt.Println("Product Struct", *res)
	//	fmt.Println("Offer: offer_sku", res.Sku)
	//	fmt.Println("Offer: sku_product", res.Sku.Product)
	//	jsonBack, _ := json.MarshalIndent(*res, "", "    ")
	//	fmt.Println( /*"OfferLegacy Json:",*/ string(jsonBack))
}

func TestProductTeardown(t *testing.T) {

	res := &ProductLegacy{}
	json.Unmarshal([]byte(productResp), res)

	teardown, _ := res.CacheTearDown()
	jsonPrint, _ := json.MarshalIndent(teardown, "", "    ")
	fmt.Println( /*"OfferLegacy Json:",*/
		string(jsonPrint))

}

func BenchmarkMarshallProductStruct(b *testing.B) {
	res := &ProductLegacy{}
	json.Unmarshal([]byte(productResp), res)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(res)
	}
}

func BenchmarkEncodeGobProductStruct(b *testing.B) {
	res := &ProductLegacy{}
	json.Unmarshal([]byte(productResp), res)

	buffer := new(bytes.Buffer)
	e := gob.NewEncoder(buffer)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		e.Encode(res)
		buffer.Reset()
	}
	//	fmt.Println("Buffer len:", buffer.Len())

}

func BenchmarkUnmarshallProductStruct(b *testing.B) {
	res := &ProductLegacy{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Unmarshal([]byte(productResp), res)
	}
}

func BenchmarkDecodeGobProductStruct(b *testing.B) {
	res := &ProductLegacy{}
	json.Unmarshal([]byte(productResp), res)

	buffer := new(bytes.Buffer)
	e := gob.NewEncoder(buffer)
	e.Encode(res)
	//	fmt.Println(buffer)

	d := gob.NewDecoder(buffer)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		d.Decode(&res)
	}
	//	fmt.Println("Res:", res)

}

func BenchmarkSaveWeirdStuffRedis(b *testing.B) {

	c, _ := redis.Dial("tcp", conf.Config("ipPortRedis", "localhost:6379"))
	defer c.Close()
	res := &ProductLegacy{}
	json.Unmarshal([]byte(productResp), res)

	b.ResetTimer()

	buffer := new(bytes.Buffer)
	e := gob.NewEncoder(buffer)

	for i := 0; i < b.N; i++ {

		e.Encode(res)

		c.Do("SET", "wtf", buffer.Bytes())
		buffer.Reset()

		resp, _ := redis.Bytes(c.Do("GET", "wtf"))

		redisOut := ProductLegacy{}
		d := gob.NewDecoder(bytes.NewReader(resp))

		d.Decode(&redisOut)

	}
}

func BenchmarkSaveJsonWeirdRedis(b *testing.B) {

	c, _ := redis.Dial("tcp", conf.Config("ipPortRedis", "localhost:6379"))
	defer c.Close()
	res := &ProductLegacy{}
	json.Unmarshal([]byte(productResp), res)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		json.Marshal(res)
		c.Do("SET", "json", res)

		resp, _ := redis.String(c.Do("GET", "json"))
		redisOut := ProductLegacy{}
		json.Unmarshal([]byte(resp), redisOut)

	}
}

func TestProductToItemOfferV1(t *testing.T) {
	res := &ProductLegacy{}
	json.Unmarshal([]byte(productResp), res)

	checkout, status := res.ToItemOfferV1()
	fmt.Printf("\n product -> CheckoutV1: %+v\n", checkout)
	fmt.Printf("\n product -> CheckoutV1 -> status: %+v\n", status)
}

func TestOfferToItemOfferV1(t *testing.T) {
	res := &OfferLegacy{}
	json.Unmarshal([]byte(offerResp), res)

	checkout, status := res.ToItemOfferV1()
	fmt.Printf("\n Offer -> CheckoutV1: %+v\n", checkout)
	fmt.Printf("\n Offer -> CheckoutV1: %+v\n", status)
	fmt.Printf("\n Offer -> CheckoutV1 -> categories: %+v\n", checkout.Categories)
}

func TestProductToItemV1(t *testing.T) {
	res := &ProductLegacy{}
	json.Unmarshal([]byte(itemProdSkus), res)
	res.ToItemV1()
	//	spew.Dump(item)
}

func TestProductToItemV1VariationOffers(t *testing.T) {
	res := &ProductLegacy{}
	json.Unmarshal([]byte(itemResp), res)
	item := res.ToItemV1()
	for _, variations := range item.Variations {
		for _, _ = range variations.Offers {
			//			spew.Dump(offer)
		}
	}
}

func TestSortFieldGroups(t *testing.T) {
	res := &ProductLegacy{}
	json.Unmarshal([]byte(itemResp), res)
	fieldGroups := res.FieldGroups
	sort.Sort(fieldGroups)

	//	spew.Config.MaxDepth = 4
	//	spew.Dump(fieldGroups)
}

func TestSortFieldGroupAttributes(t *testing.T) {
	res := &ProductLegacy{}
	json.Unmarshal([]byte(itemResp), res)
	fieldGroups := res.FieldGroups
	sort.Sort(fieldGroups)

	for _, fg := range fieldGroups {
		fmt.Println(" -------- FieldGroup:", fg.Name)
		attr := fg.Attributes
		sort.Sort(attr)
		//		spew.Config.MaxDepth = 4
		//		spew.Dump(attr)
	}
}

func TestSortAttributes(t *testing.T) {
	//	item := itemProdSkus
	item := itemQaMeia

	res := &ProductLegacy{}
	json.Unmarshal([]byte(item), res)

	for _, sku := range res.Skus {
		fmt.Println("Sku:", sku.Name, " ------- ")
		attr := sku.Attributes
		sort.Sort(attr)
		//spew.Config.MaxDepth = 4
		for _, a := range attr {
			fmt.Println("Priority:", a.Field.Priority, "F Id:", a.Field.Id,
				"Id:", a.Id, "Value:", a.Value, "F Name:", a.Field.Name)
		}
		fmt.Println()
	}

}

func TestProductToItemV1VariationAttributes(t *testing.T) {
	res := &ProductLegacy{}
	json.Unmarshal([]byte(itemQaMeia), res)
	res.ToItemV1()
	//	for _, variation := range item.Variations {
	//		spew.Dump(variation.Attributes)
	//		spew.Dump(variation.Filters)
	//	}
}

func TestOutOfCatalogFlag(t *testing.T) {
	s := SkuLegacy{}
	s.OfferList.Offer = []OfferLegacy{}

	// False because Sku has no offers
	r := isOutOfCatalog(s)
	logger.Debug("No Offers: flag OutOfCatalog: %v", r)
	if r {
		t.Fail()
	}

	// False for at least one valid offer
	o1 := OfferLegacy{Status: false, Quantity: 0}
	o2 := OfferLegacy{Status: true, Quantity: 1}
	s.OfferList.Offer = append(s.OfferList.Offer, o1, o2)
	r = isOutOfCatalog(s)
	logger.Debug("One valid offer: flag OutOfCatalog: %v", r)
	if r {
		t.Fail()
	}

	// No one has stock, but Walmart is a seller
	s.OfferList.Offer[1].Status = false
	s.OfferList.Offer[1].Quantity = 0
	s.OfferList.Offer[1].Seller.Id = "1"
	r = isOutOfCatalog(s)
	logger.Debug("Walmart seller: flag OutOfCatalog: %v", r)
	if r {
		t.Fail()
	}

	// There are offers, not one is active, but one seller has stock
	s.OfferList.Offer[1].Seller.Id = "2"
	s.OfferList.Offer[1].Quantity = 1
	r = isOutOfCatalog(s)
	logger.Debug("One seller has stock: flag OutOfCatalog: %v", r)
	if r {
		t.Fail()
	}

	// There are offers, not one is active, no one has stock and Walmart is not a seller == true
	s.OfferList.Offer[1].Seller.Id = "2"
	s.OfferList.Offer[1].Quantity = 0
	r = isOutOfCatalog(s)
	logger.Debug("All else fails: flag OutOfCatalog: %v", r)
	if !r {
		t.Fail()
	}

}
