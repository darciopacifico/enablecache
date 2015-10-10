package legacy

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strconv"

	"github.com/op/go-logging"
	"gitlab.wmxp.com.br/bis/biro/cache"
	"gitlab.wmxp.com.br/bis/biro/config"
	"gitlab.wmxp.com.br/bis/biro/rest"
	"gitlab.wmxp.com.br/bis/biro/schema"
	"gitlab.wmxp.com.br/bis/biro/utilsbiro"
)

var (
	conf = config.CreateConfig()
	log  = logging.MustGetLogger("biro")

	TTLDefault     = conf.Config("TTLDefault", "3600")
	TTLBiroDefault = conf.Config("TTLBIRODefault", "1728000")

	TTLProduct = conf.ConfigInt("TTLProduct", TTLDefault)
	TTLSKU     = conf.ConfigInt("TTLSKU", TTLDefault)
	TTLOffer   = conf.ConfigInt("TTLOffer", TTLDefault)

	TTLBIRO = conf.ConfigInt("TTLBIRO", TTLBiroDefault)


	itemV1CachePrefix = reflect.TypeOf(schema.ItemV1{}).Name()
	itemOfferV1CachePrefix = reflect.TypeOf(schema.ItemOfferV1{}).Name()

	productCachePrefix = reflect.TypeOf(schema.ProductLegacy{}).Name()
	offerCachePrefix   = reflect.TypeOf(schema.OfferLegacy{}).Name()

	//instantiate the catalog legacy component
	AngusServices = CatalogLegacy{
		Rest: rest.GenericRESTClient{ // deal with rest client operations
			HttpCaller: rest.ExecuteRequestHot, //put a hot executor on this. Can be a mock executor, for unit tests purposes...
			//HttpCaller: pool.ExecuteRequestPool,
		},
	}
)

type CatalogLegacy struct {
	Rest rest.RESTClient
}

//Update cache
func (c CatalogLegacy) UpdateItemV1Cache(productId int, itemv1 schema.ItemV1) error {

	cacheRegistry := cache.CacheRegistry{
		CacheKey: itemV1CachePrefix + ":" + strconv.Itoa(itemv1.Id),
		Payload:  itemv1,
		Ttl:      TTLBIRO,
		HasValue: true,
	}

	return BiroCacheManager.SetCache(cacheRegistry)
}


//Update cache
func (c CatalogLegacy) UpdateItemV1OfferCache(offerId int, itemOffer schema.ItemOfferV1) error {
	if offerId != itemOffer.Id {
		return errors.New("The endpoint offer id is not the same ID of offer in request body!")
	}

	cacheRegistry := cache.CacheRegistry{
		CacheKey: itemOfferV1CachePrefix + ":" + strconv.Itoa(itemOffer.Id),
		Payload:  itemOffer,
		Ttl:      TTLBIRO,
		HasValue: true,
	}

	return BiroCacheManager.SetCache(cacheRegistry)
}


//Update cache
func (c CatalogLegacy) UpdateItemCache(productId int, product schema.ProductLegacy) error {

	cacheRegistry := cache.CacheRegistry{
		CacheKey: productCachePrefix + ":" + strconv.Itoa(product.Id),
		Payload:  product,
		Ttl:      TTLProduct,
		HasValue: true,
	}

	return BiroCacheManager.SetCache(cacheRegistry)
}

//Update cache
func (c CatalogLegacy) UpdateOfferCache(offerId int, offer schema.OfferLegacy) error {
	if offerId != offer.Id {
		return errors.New("The endpoint offer id is not the same ID of offer in request body!")
	}

	cacheRegistry := cache.CacheRegistry{
		CacheKey: offerCachePrefix + ":" + strconv.Itoa(offer.Id),
		Payload:  offer,
		Ttl:      TTLOffer,
		HasValue: true,
	}

	return BiroCacheManager.SetCache(cacheRegistry)
}

// consulta offerta
func (c CatalogLegacy) FindOfferById(id int) (schema.OfferLegacy, bool, error) {

	//this function specify how to decode the rest response in a struct
	buildMyObj := func(resp *http.Response) interface{} {
		offer := schema.OfferLegacy{}
		json.NewDecoder(resp.Body).Decode(&offer)

		//offer.StatusWalmartMarketPlace = true
		return offer
	}

	//call rest service, as specified at requestoffer template object
	resultVal, hasFind, err := c.Rest.ExecuteRESTGet(RequestOfferLegacy, buildMyObj, map[string]string{"id": strconv.Itoa(id)})

	//formal treatment for errors
	if !hasFind && err == nil {
		log.Debug("Offer não enconrtada %d!", id)
		return schema.OfferLegacy{}, hasFind, nil
	}

	if !hasFind && err != nil {
		bErr := schema.BIROError{Message: "Internal server error! ", Code: 500, Parent: err}
		return schema.OfferLegacy{}, hasFind, bErr
	}

	//formal treatment for nil
	if resultVal == nil {
		log.Error("Erro ao tentar consultar offer %d! Return null!", id)
		return schema.OfferLegacy{}, false, nil
	}

	//just formally confirm that result is offer type
	offer, isOffer := (*resultVal).(schema.OfferLegacy)

	if isOffer {
		// put the TTL at cache
		offer.Ttl = TTLOffer
		return offer, hasFind, nil

	} else {
		return schema.OfferLegacy{}, hasFind, errors.New("Erro ao tentar consultar offer. Tipo retornado nao reconhecido!")

	}
}

//consulta offerta
func (c CatalogLegacy) FindSKUById(id int) (schema.SkuLegacy, bool, error) {
	//sku
	buildMyObj := func(resp *http.Response) interface{} {

		sku := schema.SkuLegacy{}

		//pipe reader to json decoder. on demand, low footprint to decode
		json.NewDecoder(resp.Body).Decode(&sku)

		return sku
	}

	//call rest service, as specified at requestsku template object
	resultVal, hasFind, err := c.Rest.ExecuteRESTGet(RequestSKU, buildMyObj, map[string]string{"id": strconv.Itoa(id)})

	//formal treatment for errors
	if !hasFind && err == nil {
		log.Debug("Sku não enconrtada %d!", id)
		return schema.SkuLegacy{}, hasFind, nil
	}

	if !hasFind && err != nil {
		bErr := schema.BIROError{Message: "Internal server error! ", Code: 500, Parent: err}
		return schema.SkuLegacy{}, hasFind, bErr
	}

	//formal treatment for nil
	if resultVal == nil {
		log.Error("Erro ao tentar consultar sku %d! Return null!", id)
		return schema.SkuLegacy{}, false, nil
	}

	//just formally confirm that result is sku type
	sku, isSku := (*resultVal).(schema.SkuLegacy)

	if isSku {
		// put the TTL at cache
		sku.Ttl = TTLSKU
		return sku, hasFind, nil

	} else {
		return schema.SkuLegacy{}, hasFind, errors.New("Erro ao tentar consultar sku. Tipo retornado nao reconhecido!")

	}

}

//consulta produto no legado
func (c CatalogLegacy) FindProductById(id int) (schema.ProductLegacy, bool, error) {

	buildMyObj := func(resp *http.Response) interface{} {

		prod := schema.ProductLegacy{}
		decoder := json.NewDecoder(resp.Body)
		decoder.Decode(&prod)

		return prod
	}

	resultVal, hasFind, err := c.Rest.ExecuteRESTGet(RequestProduct, buildMyObj, map[string]string{"id": strconv.Itoa(id)})

	defaultReturn := schema.ProductLegacy{}

	if !hasFind && err == nil {
		log.Debug("Produto não enconrtada %d!", id)
		return schema.ProductLegacy{}, hasFind, nil
	}

	if !hasFind && err != nil {

		var bErr error
		switch tErr := err.(type) {
		case schema.BIROError:
			bErr = tErr
		case error:
			bErr = schema.BIROError{Message: "Internal server error: ", Code: 500, Parent: tErr}
		default:
			bErr = errors.New("Non identified error type!")
		}

		return schema.ProductLegacy{}, hasFind, bErr
	}

	if resultVal == nil {
		log.Error("erro ao tentar consultar product! %v ", err, id)
		return defaultReturn, false, errors.New("Product not found")
	}

	product, isProduct := (*resultVal).(schema.ProductLegacy)

	if isProduct {

		newSkus := make([]schema.SkuLegacy, 0)
		for _, s := range product.Skus {

			s.Product = product

			newOffers := make([]schema.OfferLegacy, 0)
			for _, o := range s.OfferList.Offer {
				o.Sku = s
				newOffers = append(newOffers, o)
			}
			s.OfferList.Offer = newOffers

			newSkus = append(newSkus, s)
		}
		product.Skus = newSkus

		product.Ttl = TTLProduct
		return product, hasFind, nil

	} else {

		return defaultReturn, false, errors.New("Erro ao tentar consultar product. Tipo retornado nao reconhecido!")
	}
}

//consulta produto no legado
func (c CatalogLegacy) FindProductNoOfferById(id int) (schema.ProductLegacy, bool, error) {

	buildMyObj := func(resp *http.Response) interface{} {

		prod := schema.ProductLegacy{}
		decoder := json.NewDecoder(resp.Body)
		decoder.Decode(&prod)

		return prod
	}

	resultVal, hasFind, err := c.Rest.ExecuteRESTGet(requestProductNoOffer, buildMyObj, map[string]string{"id": strconv.Itoa(id)})

	defaultReturn := schema.ProductLegacy{}

	if !hasFind && err == nil {
		log.Debug("Produto não enconrtada %d!", id)
		return schema.ProductLegacy{}, hasFind, nil
	}

	if !hasFind && err != nil {

		bErr := schema.BIROError{Message: "Internal server error! ", Code: 500, Parent: err}

		return schema.ProductLegacy{}, hasFind, bErr
	}

	if resultVal == nil {
		log.Error("erro ao tentar consultar product! %v ", err, id)
		return defaultReturn, false, errors.New("Product not found")
	}

	product, isProduct := (*resultVal).(schema.ProductLegacy)

	if isProduct {

		product.Ttl = TTLProduct
		return product, hasFind, nil

	} else {

		return defaultReturn, false, errors.New("Erro ao tentar consultar product. Tipo retornado nao reconhecido!")
	}
}

/*
func FindCheckoutByProduct(id int, version string) (schema.CheckoutV1, bool, error) {
	//prod, found, err:= findProductById(id)
	prod, found, err := cachedFindProductById(id)

	if err != nil {
		log.Error("erro ao tentar consultar checkout por product! ", id)
		return schema.SimpleItem{}, false, err
	}

	if found {
		return *(prod.ToSimpleItem()), true, nil
	} else {
		return schema.SimpleItem{}, false, nil
	}
}

func FindCheckoutBySKU(id int, version string) (schema.SimpleItem, bool, error) {
	//sku, found, err:= findSKUById(id)
	sku, found, err := cachedFindSKUById(id)

	if err != nil {
		log.Error("erro ao tentar consultar checkout por sku! ", id)
		return schema.SimpleItem{}, false, err
	}

	if found {
		return *(sku.ToSimpleItem()), true, nil
	} else {
		return schema.SimpleItem{}, false, nil
	}
}
*/
func (c CatalogLegacy) OfferTransformer_412(id int, param map[string]string, version string, finderLegacy utilsbiro.TypeGenericSourceFinder) (interface{}, bool, error) {

	value, found, err := finderLegacy(id)

	offer, _ := value.(schema.OfferLegacy)

	if err != nil {
		log.Error("erro ao tentar consultar checkout por offer %v ", id)
		return struct{ Error string }{Error: "Internal server error. Consult logs! parent:" + err.Error()}, false, err
	}

	if found {
		log.Debug("Transfering ttl value %d from offer %d to ItemOfferV1 ", offer.GetTtl(), id)
		itemOfferV1, status := offer.ToItemOfferV1()
		itemOfferV1.Ttl = TTLBIRO

		//status checks consistency of returned offer.
		if status { //totally consistent offer
			itemOfferV1.StatusCode = 200
			return itemOfferV1, status, nil

		} else { //offer with some inconsistency
			itemOfferV1.StatusCode = 412
			return itemOfferV1, true, schema.BIROError{Message: "Inconsistent Offer", Parent: nil, Code: schema.ERROR_INVALID_ENTITY}
		}
	}

	// Offer not found
	return schema.ItemOfferV1{}, false, nil

}

func (c CatalogLegacy) SkuTransformer(id int, param map[string]string, version string, finderLegacy utilsbiro.TypeGenericSourceFinder) (interface{}, bool, error) {

	value, found, err := finderLegacy(id)

	skuLegacy, _ := value.(schema.SkuLegacy)

	if err != nil {
		log.Error("erro ao tentar consultar checkout por sku! ", id)
		return struct{ Error string }{Error: "Internal server error. Consult logs. parent:" + err.Error()}, false, err
	}

	if found {
		log.Debug("Transfering ttl value %d from sku %d to VariationV1 ", skuLegacy.GetTtl(), id)

		productLegacy := skuLegacy.Product
		productLegacy.Skus = []schema.SkuLegacy{skuLegacy}

		itemV1 := productLegacy.ToItemV1()
		itemV1.Ttl = TTLBIRO
		itemV1.StatusCode = 200

		return itemV1, true, nil

	}

	// Sku not found
	return schema.ItemV1{}, false, nil

}

func (c CatalogLegacy) OfferTransformer(id int, param map[string]string, version string, finderLegacy utilsbiro.TypeGenericSourceFinder) (interface{}, bool, error) {

	value, found, err := finderLegacy(id)

	offer, _ := value.(schema.OfferLegacy)

	if err != nil {
		log.Error("erro ao tentar consultar checkout por offer! ", id)
		return struct{ Error string }{Error: "Internal server error. Consult logs. parent:" + err.Error()}, false, err
	}

	if found {
		log.Debug("Transfering ttl value %d from offer %d to ItemOfferV1 ", offer.GetTtl(), id)
		itemOfferV1, _ := offer.ToItemOfferV1()
		itemOfferV1.Ttl = TTLBIRO
		itemOfferV1.StatusCode = 200

		return itemOfferV1, true, nil // mark as valid, true, always.
	}

	// Offer not found
	return schema.ItemOfferV1{}, false, nil

}

func (c CatalogLegacy) ItemTransformer_412(id int, param map[string]string, version string, funcaoDeBusca utilsbiro.TypeGenericSourceFinder) (interface{}, bool, error) {

	// find
	value, found, err := funcaoDeBusca(id)

	//must be product legacy there is no other possible option
	product, _ := value.(schema.ProductLegacy)

	if err != nil {
		log.Error("Erro ao tentar consultar ItemV1 por Product! ", id)
		return schema.ItemV1{}, false, err
	}

	if found {
		log.Debug("Transfering ttl value %d from offer %d to ItemOfferV1 ", product.GetTtl(), id)

		err := product.IsValid()

		itemV1 := product.ToItemV1()
		itemV1.Ttl = TTLBIRO

		//status checks consistency of returned product.
		if err == nil { //totally consistent product
			log.Debug("O produto consultado é consistente! Convertendo para itemV1! %v", id)
			itemV1.StatusCode = 200
			return *itemV1, true, nil

		} else { //product with some inconsistency

			log.Debug("O produto consultado nao esta consistente! Refornando apenas category! %v", id)
			// returns only an array of categories and the ttl information

			itemV1.StatusCode = 412
			return *itemV1, true, err
		}
	}

	// Product not found
	return schema.ItemV1{}, false, nil

}

//Transforms returned product legacy to itemV1, without 412 checking! used to multi ids REST/GETs...
func (c CatalogLegacy) ItemTransformer(id int, param map[string]string, version string, funcaoDeBusca utilsbiro.TypeGenericSourceFinder) (interface{}, bool, error) {

	// find
	value, found, err := funcaoDeBusca(id)

	//must be product legacy there is no other possible option
	product, _ := value.(schema.ProductLegacy)

	if err != nil {
		log.Error("erro ao tentar consultar ItemV1 por Product! ", id)
		return struct{ Error string }{Error: "Internal server error. Consult logs. parent:" + err.Error()}, false, err
	}

	if found {
		log.Debug("Transfering ttl value %d from offer %d to ItemOfferV1 ", product.GetTtl(), id)

		itemV1 := product.ToItemV1()
		itemV1.Ttl = TTLBIRO
		itemV1.StatusCode = 200

		log.Debug("O produto consultado é consistente! Convertendo para itemV1! %v", id)
		return itemV1, true, nil

	}

	// Product not found
	return schema.ItemV1{}, false, nil

}

//return ttl for a offer
func (c CatalogLegacy) OfferCacheKeyBuilder(id int) string {
	return reflect.TypeOf(schema.OfferLegacy{}).Name() + ":" + strconv.Itoa(id)
}

/*
Oferta OK (200) - http://10.134.113.70:8080/item/offer/192776
Oferta inexistente (Not found – 404) - http://10.134.113.70:8080/item/offer/348232
Oferta desativada - http://10.134.113.70:8080/item/offer/870055
SKU desativado - http://10.134.113.70:8080/item/offer/305941
Produto desativado - http://10.134.113.70:8080/item/offer/305944
*/
