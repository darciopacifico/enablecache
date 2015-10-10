package legacy

import (
	//	"reflect"
	"encoding/gob"
	"testing"
	"time"

	"gitlab.wmxp.com.br/bis/biro/rest"
	"gitlab.wmxp.com.br/bis/biro/schema"
	//	"gitlab.wmxp.com.br/bis/biro/aop"
)

var (
	//hot component catalog legacy. Depends on legacy server availability
	catalogLegacy = CatalogLegacy{
		rest.GenericRESTClient{
			HttpCaller: rest.ExecuteRequestHot,
		},
	}

	//mock catalog legacy
	catalogLegacyMocked = CatalogLegacy{
		rest.GenericRESTClient{
			HttpCaller: ExecuteRequestMock,
		},
	}
)

func init() {
	gob.Register(schema.AttributesLegacy{})
	gob.Register(schema.CategoryLegacy{})
	gob.Register(schema.OfferLegacy{})
	gob.Register(schema.ProductLegacy{})
	gob.Register(schema.SkuFile{})
	gob.Register(schema.SkuLegacy{})
}

func TestSetter(t *testing.T) {

}

func TestLegacyFindProductMock(t *testing.T) {
	prod, hasProd, err := catalogLegacyMocked.FindProductById(666666666)

	if err != nil {
		log.Error("Erro ao tentar executar busca!", err)
		t.Fail()
	}

	if hasProd {
		log.Debug("has some prod id %d", prod.Id)
		log.Debug("has some prod name %s", prod.Name)
	}
}

func _TestLegacyFindProductHot(t *testing.T) {
	prod, hasProd, err := catalogLegacy.FindProductById(2774376)

	if err != nil {
		log.Error("Erro ao tentar executar busca!", err)
		t.Fail()
	}

	if hasProd {
		log.Debug("has some prod id %d", prod.Id)
		log.Debug("has some prod name %s", prod.Name)
	}
}

func _TestLegacyFindProductNoOffer(t *testing.T) {
	prod, hasProd, err := catalogLegacy.FindProductNoOfferById(2774376)

	if err != nil {
		log.Error("Erro ao tentar executar busca de prod no offer!", err)
		t.Fail()
	}

	if hasProd {
		log.Debug("has some prod id %d", prod.Id)
		log.Debug("has some prod name %s", prod.Name)
	}
}

func TestLegacyFindSKUMock(t *testing.T) {
	sku, hasProd, err := catalogLegacyMocked.FindSKUById(666666666)

	if err != nil {
		log.Error("Erro ao tentar executar busca!", err)
		t.Fail()
	}

	if hasProd {
		log.Debug("has some sku id %d", sku.Id)
		log.Debug("has some sku name %s", sku.Name)
	}
}

/*
func TestLegacyFindOfferPrice(t *testing.T) {
	offer, hasProd, err := catalogLegacy.FindOfferById(946410)

	if err != nil {
		log.Error("Erro ao tentar executar busca!", err)
		t.Fail()
	}

	if hasProd {
		log.Debug("has some offer id %d", offer.Id)
		log.Debug("has some offer seller id %s", offer.SellerId)
		log.Debug("has some offer price	%d", offer.Price)
		log.Debug("has some offer listPrice  %d", offer.ListPrice)
	}

	if offer.Price == 0 || offer.ListPrice == 0 {
		log.Error("Erro! offer price is zero! Please check offer %v at origin database!", offer.Id)
		//t.Fail()
	} else {
		log.Debug("Offer %v price is %v and listprice %v", offer.Id, offer.Price, offer.ListPrice)
	}

}
*/
func _TestLegacyFindOfferHot(t *testing.T) {
	offer, hasProd, err := catalogLegacy.FindOfferById(192776)

	if err != nil {
		log.Error("Erro ao tentar executar busca!", err)
		t.Fail()
	}

	if hasProd {
		log.Debug("has some offer id %d", offer.Id)
		log.Debug("has some offer seller id %s", offer.SellerId)
		log.Debug("has some offer price	%d", offer.Price)
		log.Debug("has some offer listPrice  %d", offer.ListPrice)

	}

}
func TestLegacyFindOfferMock(t *testing.T) {
	offer, hasProd, err := catalogLegacyMocked.FindOfferById(666666666)

	if err != nil {
		log.Error("Erro ao tentar executar busca!", err)
		t.Fail()
	}

	if hasProd {
		log.Debug("has some offer id %d", offer.Id)
		log.Debug("has some offer seller id %s", offer.SellerId)
		log.Debug("has some offer price	%f", offer.Price)
		log.Debug("has some offer listPrice  %f", offer.ListPrice)

	}
}

func TestLegacyFindCheckoutOfferMock(t *testing.T) {
	someVal, hasProd, err := catalogLegacyMocked.OfferTransformer_412(666666666, map[string]string{}, "",
		func(id int) (interface{}, bool, error) {
			return catalogLegacyMocked.FindOfferById(id)
		})

	switch e := err.(type) {
	case schema.BIROError:
		log.Debug("OK! Error Schema.BIROError as expected!", e)
		t.Fail()
	case error:
		log.Error("Erro ao tentar executar busca! Erro nao esperado!", e)
		t.Fail()
	default:
		log.Debug("Has no errors!")
	}

	if hasProd {
		switch val := someVal.(type) {

		case schema.ItemOfferV1:
			log.Debug("has some checkout id %d", val.Id)
			log.Debug("has some checkout seller id %s", val.Name)
			log.Debug("current  %d", val.Offer.Price.Current)
			log.Debug("original %d", val.Offer.Price.Original)

		case schema.ArrayCategories:
			log.Error("There are some inconsistency error at Offer! Category return was not expected! ", val.Categories)
			t.Fail()
		}

	}

	time.Sleep(time.Millisecond * 100)

}

func _TestLegacyFindHot(t *testing.T) {
	prod, hasProd, err := catalogLegacy.FindProductById(2023268)

	if err != nil {
		log.Error("Erro ao tentar executar busca!", err)
		t.Fail()
	}

	if hasProd {
		log.Debug("has some prod id %d", prod.Id)
		log.Debug("has some prod name %s", prod.Name)
	}
}

/*
func TestLegacyFindCheckoutOfferHot(t *testing.T) {
	offer := 464375
	//	offer := 192776
	someVal, hasProd, err := catalogLegacy.FindCheckoutByOfferTransformer(offer, "v1",
		func(id int, cache bool) (interface{}, bool, error) {
			return catalogLegacy.FindOfferById(id)
		})
	checkout, isCheckout := someVal.(schema.ItemOfferV1)

	if !isCheckout {
		t.Fail()
	}

	if err != nil {
		log.Error("Erro ao tentar executar busca!", err)
		t.Fail()
	}

	if hasProd {
		log.Debug("has some checkout id %d", checkout.Id)
		log.Debug("has some checkout seller id %s", checkout.Name)
		log.Debug("current  %d", checkout.Offer.Price.Current)
		log.Debug("original %d", checkout.Offer.Price.Original)
		log.Debug("specialization %+v", checkout.Specializations)
	}

	time.Sleep(time.Millisecond * 100)
}
*/

func _TestLegacyFindCheckoutOfferHotCached(t *testing.T) {
	// same call as usual
	/*
		someVal, hasProd, err := catalogLegacy.FindCheckoutByOffer(192776, "v1")
		someVal, hasProd, err = catalogLegacy.FindCheckoutByOffer(192776, "v1")
	*/
	someVal1, hasProd, err := catalogLegacy.OfferTransformer_412(348231, map[string]string{}, "v1",
		func(id int) (interface{}, bool, error) {
			return catalogLegacy.FindOfferById(id)
		})

	checkout, isCheckout := someVal1.(schema.ItemOfferV1)

	if !isCheckout {
		log.Error("Expected object is not an ItemOfferV1", someVal1)
		t.Fail()
	}

	if err != nil {
		log.Error("Erro ao tentar executar busca!", err)
		t.Fail()
	}

	if hasProd {
		log.Debug("has some checkout id %d", checkout.Id)
		log.Debug("has some checkout seller id %s", checkout.Name)
		log.Debug("current  %d", checkout.Offer.Price.Current)
		log.Debug("original %d", checkout.Offer.Price.Original)
		log.Debug("specialization %+v", checkout.Specializations)
	}

	time.Sleep(time.Millisecond * 100)
}

func TestLegacyFindCheckoutOfferHotCachedNotFound(t *testing.T) {
	// same call as usual
	/*
		someVal, hasProd, err := catalogLegacy.FindCheckoutByOffer(192776, "v1")
		someVal, hasProd, err = catalogLegacy.FindCheckoutByOffer(192776, "v1")
	*/

	someVal1, hasProd, err := catalogLegacy.OfferTransformer_412(666666666, map[string]string{}, "v1",
		func(id int) (interface{}, bool, error) {
			return catalogLegacyMocked.FindOfferById(id)
		})

	returnedVal, isCheckout := someVal1.(schema.ItemOfferV1)

	if !isCheckout {
		log.Error("the returned object is not expected type", returnedVal)
		t.Fail()
	}

	if err != nil {
		log.Error("Erro ao tentar executar busca!", err)
		t.Fail()
	}

	if hasProd {
		log.Debug("None valid return was expected", returnedVal)
	}

	time.Sleep(time.Millisecond * 100)
}
