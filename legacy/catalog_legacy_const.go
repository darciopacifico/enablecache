package legacy

import (
	"bytes"
	"errors"
	"fmt"
	"gitlab.wmxp.com.br/bis/biro/rest"
	"gitlab.wmxp.com.br/bis/biro/utilsbiro"
	"net/http"
)

type BiroRequestTemplate struct {
	rest.RequestTemplate
	Displays []string
}

func (r BiroRequestTemplate) FillAuthentication(httpRequest *http.Request) {
	if len(r.User) > 0 {
		httpRequest.SetBasicAuth(r.User, r.Pass)
	}
}

//compose uri for a rest service call, based on RequestPayload
func (r BiroRequestTemplate) GetURI(params map[string]string) string {

	uri := r.Uri

	//there are any display specified in search template?

	if len(r.Displays) > 0 {

		var buffer bytes.Buffer

		buffer.WriteString(uri)

		buffer.WriteString("?display=")

		//concatenate those displays
		for _, display := range r.Displays {
			buffer.WriteString(display)
			buffer.WriteString("&display=")
		}

		uri = buffer.String()
	}

	id, hasId := params["id"]
	if !hasId {
		panic(errors.New("id item map not informed!!"))
	}

	arrId := []string{id}
	interfaceParam := utilsbiro.ToInterfaceArr(arrId)

	uri = fmt.Sprintf(uri, interfaceParam...)

	return uri

}

var RequestOfferLegacy = BiroRequestTemplate{

	RequestTemplate: rest.RequestTemplate{
		Uri:  conf.Config("URIwalmartLegacyOffer", "http://vip-cat-serv.qa.vmcommerce.intra/ws/offers/%[1]s.json"),
		User: conf.Config("userLegacyOffer", "admin"),
		Pass: conf.Config("passLegacyOffe", "admin"),
	},
	Displays: []string{

		"offer_seller",
		"offer_store",
		"offer_additional",

		"offer_sku", //JOIN DISPLAY OFFER->SKU
		"sku_eans",
		"sku_skufile",
		"sku_freighttype",
		"sku_attributes",
		"skuattribute_field", // ADDITIONAL DISPLAY, TO SHOW ATTRIBUTE VALUES
		"sku_eans",
		"sku_skufile",
		"sku_freighttype",
		"sku_attributes",

		"sku_product", //JOIN DISPLAY SKU->OFFER
		"product_category",
		"product_brand",
		"product_attributes",
		"productattribute_field", // ADDITIONAL DISPLAY, TO SHOW ATTRIBUTE VALUES
		"product_department",
		"product_fieldgroups",
		"product_tags",

		"category_parent",
		"category_parent_deep",
		"category_children",
		"category_fieldgroups",

		//"product_sku", 		// FORBIDEN DISPLAY. AVOID INFINITY LOOP
		//"sku_offer",	 		// FORBIDEN DISPLAY. AVOID INFINITY LOOP
	},
}

//template to the find SKU legacy service
var RequestSKU = BiroRequestTemplate{

	RequestTemplate: rest.RequestTemplate{
		Uri:  conf.Config("URIwalmartLegacySKU", "http://vip-cat-serv.qa.vmcommerce.intra/ws/skus/%[1]s.json"),
		User: conf.Config("userLegacySKU", "admin"),
		Pass: conf.Config("passLegacySKU", "admin"),
	},
	Displays: []string{
		//sku attributes
		"sku_eans",
		"sku_skufile",
		"sku_freighttype",
		"sku_attributes",
		"skuattribute_field",

		"sku_product", // join display, to attach product
		//"product_sku", // AVOID CYCLICAL RELATIONSHIP AND STACK OVERFLOW!!!
		"productattribute_field",
		"product_category",
		"product_brand",
		"product_attributes",
		"product_department",
		"product_fieldgroups",
		"product_tags",

		"category_parent",
		"category_parent_deep",
		"category_children",
		"category_fieldgroups",

		"sku_offer", // join display, to attach offers
		//"offer_sku",// AVOID CYCLICAL RELATIONSHIP AND STACK OVERFLOW!!!
		"offer_seller",
		"offer_store",
		"offer_additional",
	},
}

//template to the find product legacy service
var RequestProduct = BiroRequestTemplate{
	RequestTemplate: rest.RequestTemplate{
		Uri:  conf.Config("URIwalmartLegacyProduct", "http://vip-cat-serv.qa.vmcommerce.intra/ws/products/%[1]s.json"),
		User: conf.Config("userLegacyProduct", "admin"),
		Pass: conf.Config("passLegacyProduct", "admin"),
	},
	Displays: []string{
		"product_category",
		"product_brand",
		"product_attributes",
		"productattribute_field",
		"product_department",
		"product_fieldgroups",
		"product_tags",

		"category_parent",
		"category_parent_deep",
		"category_children",
		"category_fieldgroups",

		"product_sku", // join display, to attach skus
		//"sku_product",// AVOID CYCLICAL RELATIONSHIP AND STACK OVERFLOW!!!
		"sku_eans",
		"sku_skufile",
		"sku_freighttype",
		"sku_attributes",
		"skuattribute_field",

		"sku_offer", // join display, to attach offers
		//"offer_sku",// AVOID CYCLICAL RELATIONSHIP AND STACK OVERFLOW!!!
		"offer_seller",
		"offer_store",
		"offer_additional",
	},
}

//template to the find product legacy service
var requestProductNoOffer = BiroRequestTemplate{
	RequestTemplate: rest.RequestTemplate{
		Uri:  conf.Config("URIwalmartLegacyProduct", "http://vip-cat-serv.qa.vmcommerce.intra/ws/products/%[1]s.json"),
		User: conf.Config("userLegacyProduct", "admin"),
		Pass: conf.Config("passLegacyProduct", "admin"),
	},
	Displays: []string{
		"product_category",
		"product_brand",
		"product_attributes",
		"productattribute_field",
		"product_department",
		"product_fieldgroups",
		"product_tags",

		"category_parent",
		"category_parent_deep",
		"category_children",
		"category_fieldgroups",

		"product_sku", // join display, to attach skus
		//"sku_product",// AVOID CYCLICAL RELATIONSHIP AND STACK OVERFLOW!!!
		"sku_eans",
		"sku_skufile",
		"sku_freighttype",
		"sku_attributes",
		"skuattribute_field",

		//"sku_offer", // join display, to attach offers
		//"offer_sku",// AVOID CYCLICAL RELATIONSHIP AND STACK OVERFLOW!!!
		//"offer_seller",
		//"offer_store",
		//"offer_additional",
	},
}
