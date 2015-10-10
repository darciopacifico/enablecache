package legacy

import (
	"testing"

	"gitlab.wmxp.com.br/bis/biro/schema"
)

var ()

/*
func TestMixBisLegacy(t *testing.T) {
	idProd := 2000395

	prodLegacyNoOffers, hasProd, errProd := FindProductNoOfferByIdCache(idProd, true)
	offersByProd, hasOffers, errOff := FindOffersByProdCache(idProd, true)

	if !hasOffers || !hasProd || errProd != nil || errOff != nil {
		log.Error("Test precondition failed hasProd %v hasOffers %v errProd %v errOff %v", hasProd, hasOffers, errProd, errOff)
		t.Fail()
	}

	prodLegacy := MixResults(prodLegacyNoOffers, offersByProd)

	if len(prodLegacy.Skus) == 0 {
		log.Error("Nenhuma sku encontrada!")
		t.Fail()
	}

	for _, sku := range prodLegacy.Skus {
		if len(sku.OfferList.Offer) == 0 {
			log.Error("Nenhuma offer encontrada, sku %v!", sku.Id)
		}

		if len(sku.OfferList.Offer) == 0 {
			log.Error("SKU %v ", sku.Id)
		} else {
			log.Debug("SKU %v possui %v offers", sku.Id, len(sku.OfferList.Offer))
		}

	}

	_ = spew.Dump
}
*/

/*
func TestMixBisLegacy2(t *testing.T) {
	idProd := 2000395

	prodLegacyNoOffers, hasProd, errProd := FindProductNoOfferByIdCache(idProd, true)
	offersByProd, hasOffers, errOff := FindOffersByProdCache(idProd, true)

	if !hasOffers || !hasProd || errProd != nil || errOff != nil {
		log.Error("Test precondition failed hasProd %v hasOffers %v errProd %v errOff %v", hasProd, hasOffers, errProd, errOff)
		t.Fail()
	}

	prodLegacy := MixResults(prodLegacyNoOffers, offersByProd)

	if len(prodLegacy.Skus) == 0 {
		log.Error("Nenhuma sku encontrada!")
		t.Fail()
	}

	for _, sku := range prodLegacy.Skus {
		if len(sku.OfferList.Offer) == 0 {
			log.Error("Nenhuma offer encontrada, sku %v!", sku.Id)
		}

		if len(sku.OfferList.Offer) == 0 {
			log.Error("SKU %v ", sku.Id)
		} else {
			log.Debug("SKU %v possui %v offers", sku.Id, len(sku.OfferList.Offer))
		}

	}

	_ = spew.Dump
}
*/

//TODO fix with Thrift repo
//func TestMixFindOffersProd(t *testing.T) {
//	idProd := 2274193
//
//	prodLegacy, hasProd, err := FindItemByProductMix(idProd, false)
//
//	if !hasProd {
//		log.Error("Produto nao encontrado!")
//		t.Fail()
//	}
//
//	if err != nil {
//		log.Error("Erro ao tentar encontrar produto!")
//		t.Fail()
//	}
//
//	if len(prodLegacy.Skus) == 0 {
//		log.Error("Nenhuma sku encontrada!")
//		t.Fail()
//
//	}
//
//	for _, sku := range prodLegacy.Skus {
//
//		if len(sku.OfferList.Offer) == 0 {
//			log.Error("Nenhuma offer encontrada, sku %v!", sku.Id)
//		}
//
//	}
//
//	_ = spew.Dump
//	//spew.Dump(prodLegacy.Skus[0])
//
//}

func BenchmarkCopyDeep(b *testing.B) {
	idProd := 2000395

	prodLegacyNoOffers, hasProd, errProd := FindProductNoOfferByIdCache(idProd, true)

	if !hasProd {
		b.Fail()
		log.Error("Erro ao tentar consultar produto %v para teste de deepcopy!", idProd, errProd)
		return
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var prodNew schema.ProductLegacy
		DeepCopy(prodLegacyNoOffers, &prodNew)
		_ = prodNew
	}

}

func BenchmarkCopyMem(b *testing.B) {
	idProd := 2000395

	prodLegacyNoOffers, hasProd, errProd := FindProductNoOfferByIdCache(idProd, true)

	if !hasProd {
		b.Fail()
		log.Error("Erro ao tentar consultar produto %v para teste de deepcopy!", idProd, errProd)
		return
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		var prodNew schema.ProductLegacy
		prodNew = prodLegacyNoOffers
		_ = prodNew
	}

}
