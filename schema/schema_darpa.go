package schema

type DarpaRequest struct {
	Offers []DarpaOffer `json:"offers"`
}

type DarpaOffer struct {
	ID              int        `json:"id"`
	ItemVariationId int        `json:"itemVariationId"`
	Seller          Seller     `json:"seller"`
	Price           Price      `json:"price"`
	Itemid          int        `json:"itemId"`
	Categories      []Category `json:"categories"`
	Brand           Brand      `json:"brand"`
	Available       bool       `json:"available"`
}

type DarpaResponse struct {
	Requestid     string                 `json:"requestId"`
	Statuscode    int                    `json:"statusCode"`
	Statusmessage string                 `json:"statusMessage"`
	Payload       []DarpaResponsePayload `json:"payload"`
}

//
type DarpaResponsePayload struct {
	Offerid int    `json:"offerId"`
	Seller  Seller `json:"seller"`
	Price   Price  `json:"price"` // copy to biro offer
	DarpaRespAttachment
	ID int `json:"id"`
}

//part of Darpa response that can be attached to BIRO offer
type DarpaRespAttachment struct {
	Installments *struct {
		Bestcalculatedinstallment         *CalculatedInstallment         `json:"bestCalculatedInstallment"`
		Bestcalculatedinstallmentwithrate *CalculatedInstallmentWithRate `json:"bestCalculatedInstallmentWithRate,omitempty"`
	} `json:"installments,omitempty"`
	Utm *UTM `json:"utm,omitempty"`
}

type UTM struct {
	Partner  interface{} `json:"partner"`
	Medium   interface{} `json:"medium"`
	Campaign interface{} `json:"campaign"`
}

type CalculatedInstallment struct {
	Price               int    `json:"price"`
	Valueperinstallment int    `json:"valuePerInstallment"`
	Installmentamount   int    `json:"installmentAmount"`
	Currency            string `json:"currency"`
}

type CalculatedInstallmentWithRate struct {
	CalculatedInstallment
	Rate float32 `json:"rate"`
}
