package legacy

import (
	"bytes"
	"encoding/gob"
	"gitlab.wmxp.com.br/bis/biro/schema"
)

//dummy struct to deal with multi value return
type ReturnDefault struct {
	Value  interface{}
	HasVal bool
	Error  error
}

//find product, mixing product from catalog legacy and offers from new bis repo
func FindItemByProductMix(prodId int, cached bool) (schema.ProductLegacy, bool, error) {
	var offersChan, prodChan ReturnDefault

	chanOffers := make(chan ReturnDefault)
	chanProd := make(chan ReturnDefault)

	//find offers by product in BIS repository
	go func() {
		retOffers := ReturnDefault{}
		retOffers.Value, retOffers.HasVal, retOffers.Error = FindOffersByProdCache(prodId, cached)
		chanOffers <- retOffers
	}()

	// find product+sku on catalog legacy, ingnoring any offer informartion.
	// Offers will be mixed with offers from bis repo
	go func() {
		retProd := ReturnDefault{}
		retProd.Value, retProd.HasVal, retProd.Error = FindProductNoOfferByIdCache(prodId, cached)
		chanProd <- retProd
	}()

	//wait for 2 responses, or one timeout...
	for i := 0; i < 2; i++ {
		select {
		case offersChan = <-chanOffers:
			log.Debug("Offers search arrive!")

		case prodChan = <-chanProd:
			log.Debug("Prod arrive!")
			/*
				case <-time.After(time.Millisecond * 4000):
					offers = ReturnDefault{schema.OffersByProduct{}, false, errors.New("Time out ao tentar consultar offers do produto!")}
					prod = ReturnDefault{schema.ProductLegacy{}, false, errors.New("Time out ao tentar consultar produto!")}
					break
			*/
		}
	}

	// do all parallel stuff above
	//////
	// do regular processing below

	offersByProd, hasOffer, errOffers := offersChan.Value.(schema.OffersByProduct), offersChan.HasVal, offersChan.Error
	if errOffers != nil {
		log.Error("Error trying to query legacy repo!", errOffers)
		return schema.ProductLegacy{}, false, errOffers
	}

	if !hasOffer {
		//just log
		log.Warning("No offer was found for product %v on bis repository!", prodId)
	}

	prodLegacyNoOffer, hasProd, errLeg := prodChan.Value.(schema.ProductLegacy), prodChan.HasVal, prodChan.Error
	if errLeg != nil {
		log.Error("Error trying to query bis repo!", errLeg)
		return schema.ProductLegacy{}, false, errLeg
	}
	if !hasProd {
		log.Error("Product not found %v! on legacy catalog ", prodId)
		return schema.ProductLegacy{}, false, nil
	}

	//mix product and offers
	newProd := MixResults(prodLegacyNoOffer, offersByProd)

	return newProd, true, nil
}

//mix product from catalog legacy and offers from new bis repo
func MixResults(productLegacyo schema.ProductLegacy, offers schema.OffersByProduct) schema.ProductLegacy {

	var productLegacy schema.ProductLegacy

	DeepCopy(productLegacyo, &productLegacy)
	//productLegacy = (productLegacyo)

	mapOffersBySku := make(map[int][]schema.OfferBis)
	for _, offer := range offers.Offers {
		mapOffersBySku[offer.SkuID] = append(mapOffersBySku[offer.SkuID], offer)
	}

	log.Debug(">>>>>> mapa de offers por sku %v", mapOffersBySku)

	log.Debug("qtd skus mapeadas %v", len(mapOffersBySku))
	for index, sku := range productLegacy.Skus {

		offers := mapOffersBySku[sku.Id]

		sku.OfferList.TotalResults = len(offers)

		offersLegConverted := convertToOffersLegacy(offers...)

		sku.OfferList.Offer = append(sku.OfferList.Offer, offersLegConverted...)

		productLegacy.Skus[index] = sku
	}

	return productLegacy
}

// convert
func convertToOffersLegacy(offersBis ...schema.OfferBis) []schema.OfferLegacy {
	resOffersLegacy := make([]schema.OfferLegacy, len(offersBis))

	for index, offerBis := range offersBis {
		resOffersLegacy[index] = ConvertToOfferLegacy(offerBis)
	}

	return resOffersLegacy
}

//
func DeepCopy(src interface{}, dest interface{}) error {
	var buffer bytes.Buffer

	e := gob.NewEncoder(&buffer)
	d := gob.NewDecoder(&buffer)

	errEnc := e.Encode(src)
	if errEnc != nil {
		log.Error("Error trying to deep copy (encode) structure!", errEnc)
		return errEnc
	}

	errDec := d.Decode(dest)
	if errDec != nil {
		log.Error("Error trying to deep copy (decode) structure!", errDec)
		return errDec
	}

	return nil
}
