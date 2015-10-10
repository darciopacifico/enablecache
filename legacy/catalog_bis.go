package legacy

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strconv"

	"gitlab.wmxp.com.br/bis/biro/rest"
	"gitlab.wmxp.com.br/bis/biro/schema"
)

//Contains the main bis repository findings
type CatalogBIS struct {
	Rest rest.RESTClient
}

// find offer in BIS repository
func (c CatalogBIS) FindOfferBisById(id int, cache bool) (schema.OfferBis, bool, error) {

	log.Debug("Searching for offer code %v", id)

	//this function specify how to decode the rest response in a struct
	buildMyObj := func(resp *http.Response) interface{} {
		offerBis := schema.OfferBis{}
		json.NewDecoder(resp.Body).Decode(&offerBis)
		return offerBis
	}

	//call rest service, as specified at requestoffer template object

	resultVal, hasFind, err := c.Rest.ExecuteRESTGet(requestOfferBis, buildMyObj, map[string]string{"id": strconv.Itoa(id)})

	//formal treatment for errors
	if !hasFind && err == nil {
		log.Debug("Offer não enconrtada %d!", id)
		return schema.OfferBis{}, hasFind, nil
	}

	if !hasFind && err != nil {
		bErr := schema.BIROError{Message: "Internal server error!", Code: 500, Parent: err}
		return schema.OfferBis{}, hasFind, bErr
	}

	//formal treatment for nil
	if resultVal == nil {
		log.Error("Erro ao tentar consultar offer %d! Return null!", id)
		return schema.OfferBis{}, false, nil
	}

	//just formally confirm that result is offer type
	offer, isOffer := (*resultVal).(schema.OfferBis)

	if isOffer {
		// put the TTL at cache
		offer.Ttl = TTLOffer
		return offer, hasFind, nil

	} else {
		return schema.OfferBis{}, hasFind, errors.New("Erro ao tentar consultar offer. Tipo retornado nao reconhecido!")

	}
}

// find offer in BIS repository
func (c CatalogBIS) FindOffersByProduct(prodId int, cache bool) (schema.OffersByProduct, bool, error) {

	//this function specify how to decode the rest response in a struct
	buildMyObj := func(resp *http.Response) interface{} {
		offersByProd := schema.OffersByProduct{}

		json.NewDecoder(resp.Body).Decode(&(offersByProd.Offers))
		return offersByProd
	}

	//call rest service, as specified at requestoffer template object
	resultVal, hasFind, err := c.Rest.ExecuteRESTGet(requestOffersByProdBis, buildMyObj, map[string]string{"id": strconv.Itoa(prodId)})

	//formal treatment for errors
	if !hasFind && err == nil {
		log.Debug("Offers não encontrada %d!", prodId)
		return schema.OffersByProduct{}, hasFind, nil
	}

	if !hasFind && err != nil {
		bErr := schema.BIROError{Message: "Internal server error! ", Code: 500, Parent: err}
		return schema.OffersByProduct{}, hasFind, bErr
	}

	//formal treatment for nil
	if resultVal == nil {
		log.Error("Erro ao tentar consultar offer %d! Return null!", prodId)
		return schema.OffersByProduct{}, false, nil
	}

	//just formally confirm that result is offer type

	offers, _ := (*resultVal).(schema.OffersByProduct)
	offers.ProductId = prodId

	for _, offer := range offers.Offers {
		offer.Ttl = TTLOffer
	}

	return offers, true, nil
}

//return ttl for a offer
func (c CatalogBIS) OfferCacheKeyBuilder(id int) string {
	return reflect.TypeOf(schema.OfferBis{}).Name() + ":" + strconv.Itoa(id)
}

/*
Oferta OK (200) - http://10.134.113.70:8080/item/offer/192776
Oferta inexistente (Not found – 404) - http://10.134.113.70:8080/item/offer/348232
Oferta desativada - http://10.134.113.70:8080/item/offer/870055
SKU desativado - http://10.134.113.70:8080/item/offer/305941
Produto desativado - http://10.134.113.70:8080/item/offer/305944
*/
//
func ConvertToOfferLegacy(offerBis schema.OfferBis) schema.OfferLegacy {

	resOfferLegacy := schema.OfferLegacy{}

	resOfferLegacy.Id = offerBis.OfferID
	resOfferLegacy.CreatedAt = offerBis.CreatedAt
	//resOfferLegacy.ImportTax = int(offerBis.PriceImportTax)

	resOfferLegacy.LastUpdatedAt = offerBis.UpdatedAt
	//resOfferLegacy.LastUpdatedByIp = offersBis.
	//resOfferLegacy.LastUpdatedByUser = offersBis

	resOfferLegacy.ListPrice = (offerBis.PriceCurrent)
	resOfferLegacy.Price = (offerBis.PriceOriginal)

	//resOfferLegacy.Quantity = int(offerBis.Quantity)

	//log.Debug("quantidade da offer %v %v %v", offerBis.OfferID, offerBis.Quantity, resOfferLegacy.Quantity)

	//resOfferLegacy.RequestedUpdateDate = offersBis
	resOfferLegacy.SellerExternalOfferId = offerBis.SellerOfferID
	resOfferLegacy.SellerId = offerBis.SellerID
	resOfferLegacy.SkuId = offerBis.SkuID
	resOfferLegacy.Status = offerBis.StatusSeller
	//resOfferLegacy.StatusWalmartMarketPlace = offerBis.StatusWalmart
	//resOfferLegacy.StoreId = offersBis.
	//resOfferLegacy.Version  = offersBis
	resOfferLegacy.Ttl = offerBis.Ttl

	resOfferLegacy.Seller.Id = offerBis.SellerID
	resOfferLegacy.Seller.Name = offerBis.SellerName
	resOfferLegacy.Seller.Status = offerBis.SellerStatus

	sku := schema.SkuLegacy{}
	sku.Product.Status = true
	sku.Status = true
	sku.Id = offerBis.SkuID

	//sku.Name = offerBis.SkuName

	resOfferLegacy.Sku = sku

	return resOfferLegacy

}
