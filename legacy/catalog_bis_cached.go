package legacy

import (
	"encoding/gob"
	"reflect"

	"gitlab.wmxp.com.br/bis/biro/aop"
	"gitlab.wmxp.com.br/bis/biro/cache"
	"gitlab.wmxp.com.br/bis/biro/rest"
	"gitlab.wmxp.com.br/bis/biro/schema"
)

var (
	//define empty body functions to reveice cached swap method
	CachedFindOfferBisById       TypeFindOfferBisById
	CachedFindOffersBisByProduct TypeFindOffersBisByProduct

	NoCacheFindOfferBisById       TypeFindOfferBisById
	NoCacheFindOffersBisByProduct TypeFindOffersBisByProduct

	//instantiate a original new catalog bis. there is no cache here.
	// This component will be mixed with cacheManager to make the twin cached functions
	catalogBIS = CatalogBIS{
		rest.GenericRESTClient{
			rest.ExecuteRequestHot,
			//pool.ExecuteRequestPool,
		},
	}

	//isntantiate the cache storage redis
	bisRedisCacheStorage = cache.NewRedisCacheStorage(
		conf.Config("ipPortRedis", "localhost:6379"),
		conf.Config("passwordRedis", ""),
		conf.ConfigInt("maxIdle", "8"),
		"newrepo",
	)

	//instantiate a simple cacheManager
	/*	BISCacheManager = cache.SimpleCacheManager{
		Ps: bisRedisCacheStorage,
	}*/

	//instantiate an auto cacheManager
	BISCacheManager = cache.AutoCacheManager{
		Ps: bisRedisCacheStorage,
	}
)

//define types to be maked as cached functions
type TypeFindOfferBisById func(offerId int, cache bool) (schema.OfferBis, bool, error)
type TypeFindOffersBisByProduct func(prodId int, cache bool) (schema.OffersByProduct, bool, error)

//is not possible to infer the default values for TypeFindOffersByProduct functions.
//DefaultValues implementation is needed
func (t TypeFindOffersBisByProduct) DefaultValues(outTypes []reflect.Type) []reflect.Value {

	defVals := make([]reflect.Value, 3)

	defVals[1] = reflect.ValueOf(true)
	var err error = nil
	defVals[2] = reflect.ValueOf(&err).Elem()

	return defVals

}

//initialize the cached funtions
func init() {
	gob.Register(schema.OfferBis{})
	gob.Register(make([]schema.OfferBis, 0))
	gob.Register(schema.OffersByProduct{})
	gob.Register(schema.SKUOfferListType{})

	aop.MakeSwap(&CachedFindOfferBisById, catalogBIS.FindOfferBisById, BISCacheManager, true)
	aop.MakeSwap(&CachedFindOffersBisByProduct, catalogBIS.FindOffersByProduct, BISCacheManager, true)

	aop.MakeSwap(&NoCacheFindOfferBisById, catalogBIS.FindOfferBisById, BISCacheManager, false)
	aop.MakeSwap(&NoCacheFindOffersBisByProduct, catalogBIS.FindOffersByProduct, BISCacheManager, false)
}

func FindOffersByProdCache(offerId int, cache bool) (schema.OffersByProduct, bool, error) {
	if cache {
		//log.Warning("busca COM cache FindOffersByProdCache")
		return CachedFindOffersBisByProduct(offerId, cache)
	} else {
		//log.Warning("busca SEM cache FindOffersByProdCache")
		return NoCacheFindOffersBisByProduct(offerId, cache)
	}
}

func FindOfferByIdCache(offerId int, cache bool) (schema.OfferLegacy, bool, error) {

	//busca oferta no BIS
	offerBis, hasOffer, errOffer := FindOfferBisByIdCache(offerId, cache)
	if errOffer != nil {
		log.Error("Erro ao tentar consultar offer bis", errOffer)
		return schema.OfferLegacy{}, false, errOffer
	}
	if !hasOffer {
		log.Error("Offer nao encontrada! %v", offerId)
		return schema.OfferLegacy{}, false, nil
	}

	//procura produto no catalogo legado
	productLegacy, hasProd, errProd := FindProductNoOfferByIdCache(offerBis.ProductID, cache)
	if errProd != nil {
		log.Error("Erro ao tentar consultar produto %v no legado!", offerBis.ProductID)
		return schema.OfferLegacy{}, false, errProd
	}
	if !hasProd {
		log.Debug("Produto %v nao encontrado no legado!", offerBis.ProductID)
		return schema.OfferLegacy{}, false, nil
	}

	//compoe a offer associando o produto/sku encontrado
	sku, _ := getSku(productLegacy, offerBis.SkuID)
	offerLegacy := ConvertToOfferLegacy(offerBis)
	productLegacy.Skus = []schema.SkuLegacy{}
	sku.OfferList.Offer = []schema.OfferLegacy{}
	sku.Product = productLegacy
	offerLegacy.Sku = sku

	return offerLegacy, true, nil
}

func getSku(productLegacy schema.ProductLegacy, idSku int) (schema.SkuLegacy, bool) {
	for _, sku := range productLegacy.Skus {
		if sku.Id == idSku {
			return sku, true
		}
	}
	return schema.SkuLegacy{}, false
}

func FindOfferBisByIdCache(offerId int, cache bool) (schema.OfferBis, bool, error) {

	if cache {
		return CachedFindOfferBisById(offerId, cache)

	} else {
		return NoCacheFindOfferBisById(offerId, cache)
	}

}
