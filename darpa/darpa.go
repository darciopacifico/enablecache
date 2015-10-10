package darpa

import (
	"bytes"
	"encoding/json"
	"github.com/op/go-logging"
	"gitlab.wmxp.com.br/bis/biro/config"
	"gitlab.wmxp.com.br/bis/biro/rest"
	"gitlab.wmxp.com.br/bis/biro/schema"
	"gitlab.wmxp.com.br/bis/biro/utilsbiro"
	"net/http"
	"os"
	"reflect"
)

var (
	conf = config.CreateConfig()
	log  = logging.MustGetLogger("biro")

	TTLDefault = conf.Config("TTLDefault", "3600")

	productCachePrefix = reflect.TypeOf(schema.ProductLegacy{}).Name()
	offerCachePrefix   = reflect.TypeOf(schema.OfferLegacy{}).Name()

	darpaService = DarpaStub{
		Rest: rest.GenericRESTClient{
			HttpCaller: rest.ExecuteRequestHot,
			//HttpCaller: pool.ExecuteRequestPool,
		},
	}
)

type DarpaStub struct {
	Rest rest.RESTClient
}

// consulta offerta
func (d DarpaStub) FindPromotion(dr schema.DarpaRequest, param map[string]string) ([]schema.DarpaResponse, error) {
	if d.Rest == nil {
		log.Error("Rest instance for Darpa stub is nil! Cant load BIRO App!")
		os.Exit(0)
	}

	buildMyObj := func(resp *http.Response) interface{} {
		darpaResponse := []schema.DarpaResponse{}
		err := json.NewDecoder(resp.Body).Decode(&darpaResponse)
		if err != nil {
			log.Error("Error trying to decode Darpa response!", err)
		}
		return darpaResponse
	}

	b := bytes.Buffer{}
	enc := json.NewEncoder(&b)
	err := enc.Encode(dr)

	reader := bytes.NewReader(b.Bytes())

	if err != nil {
		log.Error("Error trying to call darpa service!", err)
		return []schema.DarpaResponse{}, schema.BIROError{"Erro trying to access Darpa service!", err, 500}
	}

	darpaResp, _, err := d.Rest.ExecuteRESTPost(darpaRequest, buildMyObj, reader, param)

	if err != nil {
		log.Error("Error trying to access Darpa service", err)

		return []schema.DarpaResponse{}, schema.BIROError{
			Message: "Error trying to access Darpa service!",
			Parent:  err,
			Code:    500,
		}
	}
	typedResp, _ := (*darpaResp).([]schema.DarpaResponse)

	log.Debug("Found %v darpa promotion items, with error =nil!", len(typedResp))

	return typedResp, nil

}

//decoreates a simple biro transformer results with darpa promotion information
func DarpaTransformer(originalFinder utilsbiro.TransformerFinder) utilsbiro.TransformerFinder {
	darpaTransformer := func(id int, params map[string]string, version string, finder utilsbiro.TypeGenericSourceFinder) (interface{}, bool, error) {

		log.Debug("Calling darpa transformer ")

		respObj, boolRet, err := originalFinder(id, params, version, finder)

		switch rType := respObj.(type) {
		case *schema.ItemV1:
			return complementItem(*rType, params, boolRet, err)

		case *schema.ItemOfferV1:
			return complementItemOffer(*rType, params, boolRet, err)

		case schema.ItemV1:
			return complementItem(rType, params, boolRet, err)

		case schema.ItemOfferV1:
			return complementItemOffer(rType, params, boolRet, err)

		default:
			return respObj, boolRet, err
		}

		return respObj, boolRet, err
	}

	return darpaTransformer
}

//complement BIRO response with darpa response
func complementItem(itemV1 schema.ItemV1, params map[string]string, boolRet bool, err error) (interface{}, bool, error) {

	if err == nil && boolRet {
		log.Debug("Complementing itemv1 id=%v with darpa results!", itemV1.Id)

		//create request for darpa service
		darpaRequest := createItemV1DarpaRequest(itemV1)

		//call darpa service
		darpaResponses, errDarpa := darpaService.FindPromotion(darpaRequest, params)

		//check for error
		if errDarpa != nil {
			return nil, false, schema.BIROError{
				Message: "Error trying to find promotion!",
				Parent:  errDarpa,
				Code:    500,
			}

		}

		applyDarpaResponseToItem(&itemV1, darpaResponses)

		return itemV1, boolRet, nil

	} else {
		// ignore error or bool, return as is
		return itemV1, boolRet, err
	}

}

func applyDarpaResponseToItem(itemV1 *schema.ItemV1, darpaResponses []schema.DarpaResponse) {
	//map darpa results
	darpaResMap := map[int]schema.DarpaResponsePayload{}
	for _, darpaRes := range darpaResponses {
		for _, darpaResPayl := range darpaRes.Payload {
			darpaResMap[darpaResPayl.Offerid] = darpaResPayl
		}
	}

	for iv, v := range itemV1.Variations {
		for io, o := range v.Offers {
			darpaResp, hasResp := darpaResMap[o.Id]

			if hasResp {

				applyPromotionToOffer(&o, darpaResp)
				/*
					o.DarpaRespAttachment = darpaResp.DarpaRespAttachment
					log.Debug("darpa prices %+v", darpaResp.Price)
					o.Price = darpaResp.Price
				*/
				v.Offers[io] = o
			}
		}
		itemV1.Variations[iv] = v
	}

}

//apply promotion to offer
func applyPromotionToOffer(o *schema.Offer, darpaResp schema.DarpaResponsePayload) {
	o.DarpaRespAttachment = darpaResp.DarpaRespAttachment
	o.Price = darpaResp.Price

	if darpaResp.Price.DiscountedPrice != nil && *darpaResp.Price.DiscountedPrice > 0 {
		o.Price.Current = *darpaResp.Price.DiscountedPrice
		o.Price.DiscountedPrice = nil
	} else {
		log.Warning("Darpa doesn't show DiscountedPrice! for offer %v", o.Id)
	}

	if o.DarpaRespAttachment.Installments != nil && o.DarpaRespAttachment.Installments.Bestcalculatedinstallmentwithrate != nil {
		if o.DarpaRespAttachment.Installments.Bestcalculatedinstallmentwithrate.Installmentamount == 0 {
			log.Warning("Bestcalculatedinstallmentwithrate returned from darpa is not valid! Nulling!")
			o.DarpaRespAttachment.Installments.Bestcalculatedinstallmentwithrate = nil
		}
	}

	if o.Utm != nil && (o.Utm.Campaign == nil && o.Utm.Medium == nil && o.Utm.Partner == nil) {
		log.Warning("UTM returned from darpa is not valid! Nulling!")
		o.Utm = nil
	}
}

//apply darpa response to offer
func applyDarpaResponseToItemOffer(itemOfferV1 *schema.ItemOfferV1, darpaResponses []schema.DarpaResponse) {
	//map darpa results
	darpaResMap := map[int]schema.DarpaResponsePayload{}
	for _, darpaRes := range darpaResponses {
		for _, darpaResPayl := range darpaRes.Payload {
			darpaResMap[darpaResPayl.Offerid] = darpaResPayl
		}
	}

	o := itemOfferV1.Offer

	darpaResp, hasResp := darpaResMap[o.Id]

	if hasResp {
		applyPromotionToOffer(&o, darpaResp)
		itemOfferV1.Offer = o
	}
}

//create request for darpa service, based on item v1
func createItemV1DarpaRequest(itemV1 schema.ItemV1) schema.DarpaRequest {
	darpaOffers := []schema.DarpaOffer{}

	for _, variation := range itemV1.Variations {
		for _, offer := range variation.Offers {
			darpaOffer := schema.DarpaOffer{
				ID:              offer.Id,
				ItemVariationId: variation.Id,
				Seller:          offer.Seller,
				Price:           offer.Price,
				Itemid:          itemV1.Id,
				Categories:      itemV1.Categories,
				Brand:           itemV1.Brand,
				Available:       offer.Available,
			}
			darpaOffers = append(darpaOffers, darpaOffer)
		}
	}

	return schema.DarpaRequest{
		Offers: darpaOffers,
	}
}

//create request for darpa service, based on item v1
func createItemOfferV1DarpaRequest(itemOfferV1 schema.ItemOfferV1) schema.DarpaRequest {

	darpaOffer := schema.DarpaOffer{
		ID:              itemOfferV1.Offer.Id,
		ItemVariationId: itemOfferV1.VariationId,
		Seller:          itemOfferV1.Offer.Seller,
		Price:           itemOfferV1.Offer.Price,
		Itemid:          itemOfferV1.BasicItem.Id,
		Categories:      itemOfferV1.Categories,
		Brand:           itemOfferV1.Brand,
		Available:       itemOfferV1.Offer.Available,
	}

	return schema.DarpaRequest{
		Offers: []schema.DarpaOffer{darpaOffer},
	}
}

func complementItemOffer(itemOfferV1 schema.ItemOfferV1, params map[string]string, boolRet bool, err error) (interface{}, bool, error) {

	if err == nil && boolRet {
		log.Debug("Complementing itemOfferV1 id=%v with darpa results!", itemOfferV1.Id)

		//create request for darpa service
		darpaRequest := createItemOfferV1DarpaRequest(itemOfferV1)

		//call darpa service
		darpaResponses, errDarpa := darpaService.FindPromotion(darpaRequest, params)

		//check for error
		if errDarpa != nil {
			return nil, false, schema.BIROError{
				Message: "Error trying to find promotion!",
				Parent:  errDarpa,
				Code:    500,
			}
		}

		applyDarpaResponseToItemOffer(&itemOfferV1, darpaResponses)

		return itemOfferV1, boolRet, nil

	} else {
		// ignore error or bool, return as is
		return itemOfferV1, boolRet, err
	}
}
