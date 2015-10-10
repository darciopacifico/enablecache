package legacy

import (
	"testing"
	"time"

	"gitlab.wmxp.com.br/bis/biro/schema"
)

/*
func TestCachedFindOfferById(t *testing.T) {
	offerId := 877220

	offer, hasFind, err := CachedFindOfferBisById(offerId, true)
	time.Sleep(time.Millisecond * 100)

	if err != nil {
		log.Error("Error trying to find an offer %v", err)
		log.Error("4 Is bis catalog repository up and reacheable?")
		t.Fail()
		return
	}

	if !hasFind {
		log.Error("Offer not found! %v", offerId)
		log.Error("3 Is bis catalog repository up and reacheable?")
		t.Fail()
		return
	}

	log.Debug("Offer search succeed ! %v", offer)

}
*/

/*
func TestCachedFindOffersByProduct(t *testing.T) {
	prodId := 2000395

	log.Debug("Testing cached find offers by product")
	offersByProd, hasFind, err := CachedFindOffersBisByProduct(prodId, true)

	time.Sleep(time.Millisecond * 100)

	if err != nil {
		log.Error("Error trying to find an offers by product %v", err)
		log.Error("1 Is bis catalog repository up and reacheable?")
		t.Fail()
		return
	}

	if !hasFind || len(offersByProd.Offers) == 0 {
		log.Error("Offers by product not found! %v", prodId)
		log.Error("2 Is bis catalog repository up and reacheable?")
		t.Fail()
		return
	}

	log.Debug("Offers by prod search succedd ! %v", offersByProd)

}
*/

func BenchmarkFindOffersByProductCached(b *testing.B) {
	prodId := 2000395

	log.Debug("Testing cached find offers by product")

	var offersByProd schema.OffersByProduct
	var hasFind bool
	var err error

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		offersByProd, hasFind, err = CachedFindOffersBisByProduct(prodId, true)
	}
	time.Sleep(time.Millisecond * 100)

	if err != nil {
		log.Error("Error trying to find an offers by product %v", err)
		log.Error("x Is bis catalog repository up and reacheable?")
		b.Fail()
		return
	}

	if !hasFind || len(offersByProd.Offers) == 0 {
		log.Error("Offers by product not found! %v", prodId)
		log.Error("y Is bis catalog repository up and reacheable?")
		b.Fail()
		return
	}

	log.Debug("Offers by prod search succedd ! %v", offersByProd)

}

func BenchmarkFindOffersByProductNoCache(b *testing.B) {
	prodId := 2000395

	log.Debug("Testing cached find offers by product")

	var offersByProd schema.OffersByProduct
	var hasFind bool
	var err error

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		offersByProd, hasFind, err = catalogBis.FindOffersByProduct(prodId, true)
	}
	time.Sleep(time.Millisecond * 100)

	if err != nil {
		log.Error("Error trying to find an offers by product %v", err)
		log.Error("z Is bis catalog repository up and reacheable?")
		b.Fail()
		return
	}

	if !hasFind || len(offersByProd.Offers) == 0 {
		log.Error("Offers by product not found! %v", prodId)
		log.Error("h Is bis catalog repository up and reacheable?")
		b.Fail()
		return
	}

	log.Debug("Offers by prod search succedd ! %v", offersByProd)

}
