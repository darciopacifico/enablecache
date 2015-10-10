package schema

import "strconv"

//
type OfferBis struct {
	//											   //    COLUMNS ON CASSANDRA
	OfferID        int     `json:"offerId"`        //    offer_id, int PRIMARY KEY,
	IsAvailable    bool    `json:"isAvailable"`    //    available, boolean,
	CreatedAt      int     `json:"createdAt"`      //    created_at, timestamp,
	PriceCurrent   float32 `json:"priceCurrent"`   //    price_current, decimal,
	PriceImportTax float32 `json:"priceImportTax"` //    price_import_tax, decimal,
	PriceOriginal  float32 `json:"priceOriginal"`  //    price_original, decimal,
	ProductID      int     `json:"productId"`      //    product_id, int,
	Quantity       float32 `json:"quantity"`       //
	SellerID       string  `json:"sellerId"`       //    seller_id, int,
	SellerName     string  `json:"sellerName"`     //    seller_name, text,
	SellerOfferID  string  `json:"sellerOfferId"`  //    seller_offer_id, text,
	SellerStatus   bool    `json:"sellerStatus"`   //    seller_status, boolean,
	SkuID          int     `json:"skuId"`          //    sku_id, int,
	SkuName        string  `json:"skuName"`        //    sku_name, text,
	StatusSeller   bool    `json:"statusSeller"`   //    status_seller, boolean,
	StatusWalmart  bool    `json:"statusWalmart"`  //    status_walmart, boolean,
	UpdatedAt      int     `json:"updatedAt"`      //    updated_at, timestamp
	Ttl            int     `json:"-"`              //
}

/**

ESTE É O CONTRATO DE OFFER DO NOVO REPOSITORIO!
VAMOS MANTER HABILITADO O CONTRATO ACIMA ATE QUE O REPOSITORIO EM SCALA TENHA SIDO ALTERADO

type OfferBis struct {
	////
	OfferID                int    `json:"offerId"`                // offer_id
	VariationID            int    `json:"variationId"`            // sku_id
	ItemID                 int    `json:"itemId"`                 // product_id
	SellerOfferID          string `json:"sellerOfferId"`          // seller_offer_id
	SellerID               string `json:"sellerId"`               // seller_id
	SellerName             string `json:"sellerName"`             // seller_name
	SellerStatus           bool   `json:"sellerStatus"`           // seller_status
	PriceCurrent           int    `json:"priceCurrent"`           // price_current
	PriceOriginal          int    `json:"priceOriginal"`          // price_original
	Available              bool   `json:"available"`              // available
	OfferStatusVendor      bool   `json:"offerStatusVendor"`      // status_seller
	OfferStatusMarketplace bool   `json:"offerStatusMarketplace"` // status_walmart
	CreatedAt              int    `json:"createdAt"`              // created_at
	UpdatedAt              int    `json:"updatedAt"`              // updated_at
	Ttl                    int    `json:"-"`                      // transient

}
*/

//define the cachekey for OfferBis struct
func (o OfferBis) GetCacheKey() string {
	return "OfferBis:" + strconv.Itoa(o.OfferID)
}

//define ttl setter
func (o OfferBis) SetTtl(ttl int) interface{} {
	o.Ttl = ttl
	return o
}

//define ttl getter
func (o OfferBis) GetTtl() int {
	return o.Ttl
}

//DTO to find offers by products
type OffersByProduct struct {
	ProductId int        `json:"ProductId"`
	Offers    []OfferBis `json:"Offers"`
	Ttl       int        `json:"-"` // transient
}

//define the cachekey for OfferBis struct
func (o OffersByProduct) GetCacheKey() string {
	return "OffersByProduct:" + strconv.Itoa(o.ProductId)
}

//define ttl setter
func (o OffersByProduct) SetTtl(ttl int) interface{} {
	o.Ttl = ttl
	return o
}

//define ttl getter
func (o OffersByProduct) GetTtl() int {
	return o.Ttl
}
