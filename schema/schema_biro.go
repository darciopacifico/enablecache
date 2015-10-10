package schema

import (
	"fmt"
	"strconv"
)

// -- Generic header'ed types
type GenericHeaders struct {
	Headers map[string]string `json:"-"`
}

type HeaderTyped interface {
	GetHeaders() map[string]string
}

type HoldStatusCode interface {
	GetStatusCode() int
}

func (g GenericHeaders) GetHeaders() map[string]string {
	return g.Headers
}

type VariationV1 struct {
	BasicItem
	Categories []Category `json:"categories"`
	Variation  Variation  `json:"variation"`
	GenericHeaders
	Ttl int `json:"-"`
}

func (v VariationV1) SetTtl(ttl int) interface{} {
	v.Ttl = ttl
	return v
}
func (v VariationV1) GetTtl() int {
	return v.Ttl
}

type ItemV1 struct {
	BasicItem
	Description string              `json:"description"`
	Keywords    []string            `json:"keywords"`
	LinkText    string              `json:"linktext"`
	Categories  []Category          `json:"categories"`
	Attributes  []IdentAttribute    `json:"attributes"`
	Filters     []IdentAttribute    `json:"filters"`
	Variations  []Variation         `json:"variations"`
	Assets      *map[string][]Asset `json:"assets,omitempty"`
	GenericHeaders
	Ttl        int `json:"-"`
	StatusCode int `json:"-"`
}

func (i ItemV1) GetStatusCode() int {
	return i.StatusCode
}

//define the cachekey for OfferBis struct
func (i ItemV1) GetCacheKey() string {
	return "ItemV1:" + strconv.Itoa(i.Id)
}

func (i ItemV1) SetTtl(ttl int) interface{} {
	i.Ttl = ttl
	return i
}
func (i ItemV1) GetTtl() int {
	return i.Ttl
}

type ItemOfferV1 struct {
	BasicItem
	VariationId     int         `json:"variationId"`
	VariationName   string      `json:"variationName"`
	Specializations []Attribute `json:"specializations"`
	Image           Asset       `json:"image"`
	Categories      []Category  `json:"categories"`
	Offer           Offer       `json:"offer"`
	GenericHeaders
	Ttl        int `json:"-"`
	StatusCode int `json:"-"` // holds status code in cache... avoid to cache error object also
}

func (i ItemOfferV1) GetStatusCode() int {
	return i.StatusCode
}

func (o ItemOfferV1) SetTtl(ttl int) interface{} {
	o.Ttl = ttl
	return o
}
func (o ItemOfferV1) GetTtl() int {
	return o.Ttl
}

type ArrayCategories struct {
	GenericHeaders
	Categories []Category `json:"categories"`
	ttl        int
}

func (c ArrayCategories) SetTtl(ttl int) interface{} {
	c.ttl = ttl
	return c
}
func (c ArrayCategories) GetTtl() int {
	return c.ttl
}

type BasicItem struct {
	IdNameTyped
	Brand Brand `json:"brand"`
}

type Variation struct {
	Id              int                 `json:"id"`
	Name            string              `json:"name"`
	Specializations []Attribute         `json:"specializations"`
	Dimensions      map[string]float32  `json:"dimensions"`
	CodeList        map[string][]string `json:"codeList"`
	Attributes      []IdentAttribute    `json:"attributes"`
	Filters         []IdentAttribute    `json:"filters"`
	Assets          map[string][]Asset  `json:"assets"`
	Offers          []Offer             `json:"offers"`
	OutOfCatalog    bool                `json:"OutOfCatalog"`
}

type Asset struct {
	Url        string      `json:"url"`
	Attributes []Attribute `json:"attributes"`
}

type IdentAttribute struct {
	Id int `json:"id"`
	Attribute
}

type Attribute struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

type BasicOffer struct {
	Id     int    `json:"id"`
	Seller Seller `json:"seller"`
	Price  Price  `json:"price"`
	DarpaRespAttachment
	Available bool `json:"available"`
}

type Price struct {
	Original        int  `json:"original"`
	Current         int  `json:"current"`
	DiscountedPrice *int `json:"discountedPrice,omitempty"`
	ImportTax       int  `json:"importTax"`
}

type Offer struct {
	BasicOffer
}

type Seller struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Category struct {
	IdNameTyped
}

type Brand struct {
	IdNameTyped
}

type IdNameTyped struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

//generic error messages
const ERROR_INVALID_ENTITY = 412

type BIROError struct {
	Message string
	Parent  error
	Code    int
}

func (e BIROError) Error() string {

	var paretErrMsg string = ""

	if e.Parent != nil {
		paretErrMsg = e.Parent.Error()
	}

	return fmt.Sprintf("%s. %s", e.Message, paretErrMsg)
}
