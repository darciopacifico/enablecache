package legacy

import (
	"gitlab.wmxp.com.br/bis/biro/aop"
	"gitlab.wmxp.com.br/bis/biro/utilsbiro"
)

var (
	//cached functions
	C_ItemTransformer_412  utilsbiro.TransformerFinder
	C_OfferTransformer_412 utilsbiro.TransformerFinder
	C_ItemTransformer      utilsbiro.TransformerFinder
	C_OfferTransformer     utilsbiro.TransformerFinder
	C_SkuTransformer       utilsbiro.TransformerFinder

	//store cache only functions
	NC_ItemTransformer_412  utilsbiro.TransformerFinder
	NC_OfferTransformer_412 utilsbiro.TransformerFinder
	NC_SkuTransformer       utilsbiro.TransformerFinder
	NC_ItemTransformer      utilsbiro.TransformerFinder
	NC_OfferTransformer     utilsbiro.TransformerFinder
)

func init() {

	//create cached versions for all those functions
	aop.MakeSwapPrefix(&C_ItemTransformer_412, AngusServices.ItemTransformer_412, BiroCacheManager, true, "ItemV1")
	aop.MakeSwapPrefix(&C_OfferTransformer_412, AngusServices.OfferTransformer_412, BiroCacheManager, true, "ItemOfferV1")
	aop.MakeSwapPrefix(&C_ItemTransformer, AngusServices.ItemTransformer, BiroCacheManager, true, "ItemV1")
	aop.MakeSwapPrefix(&C_OfferTransformer, AngusServices.OfferTransformer, BiroCacheManager, true, "ItemOfferV1")
	aop.MakeSwapPrefix(&C_SkuTransformer, AngusServices.SkuTransformer, BiroCacheManager, true, "ItemV1")
	aop.MakeSwapPrefix(&NC_ItemTransformer_412, AngusServices.ItemTransformer_412, BiroCacheManager, false, "ItemV1")
	aop.MakeSwapPrefix(&NC_OfferTransformer_412, AngusServices.OfferTransformer_412, BiroCacheManager, false, "ItemOfferV1")
	aop.MakeSwapPrefix(&NC_ItemTransformer, AngusServices.ItemTransformer, BiroCacheManager, false, "ItemV1")
	aop.MakeSwapPrefix(&NC_OfferTransformer, AngusServices.OfferTransformer, BiroCacheManager, false, "ItemOfferV1")
	aop.MakeSwapPrefix(&NC_SkuTransformer, AngusServices.SkuTransformer, BiroCacheManager, false, "ItemV1")

}
