package schema

import (
	"strconv"

	"gitlab.wmxp.com.br/bis/biro/cache"
)

// Product Legacy obj and PartItself interface methods ----------
type ProductLegacy struct {
	Id                 int            `json:"id"`
	Name               string         `json:"name"`
	Cached             bool           `json:"cached"`
	CachedMembers      []string       `json:"cachedSkus"`
	CategoryId         int            `json:"categoryId"`
	Description        string         `json:"description"`
	EditionDate        int            `json:"editionDate"`
	KeyWords           string         `json:"keyWords"`
	LinkText           string         `json:"linkText"`
	MetaTagDescription string         `json:"metaTagDescription"`
	ReferenceCode      string         `json:"referenceCode"`
	ShowFlag           bool           `json:"showFlag"`
	SiteTitle          string         `json:"siteTitle"`
	BrandId            int            `json:"brandId"`
	Skus               []SkuLegacy    `json:"skus"`
	Category           CategoryLegacy `json:"category"`
	Brand              struct {
		DisplayedInHomeMenu bool     `json:"displayedInHomeMenu"`
		Id                  int      `json:"id"`
		ListKeywords        []string `json:"listKeywords"`
		Name                string   `json:"name"`
		SiteTitle           string   `json:"siteTitle"`
		Status              bool     `json:"status"`
		Text                string   `json:"text"`
	} `json:"brand"`
	FieldGroups FieldGroups `json:"fieldGroups"`
	Status      bool        `json:"status"`
	//cache.HasTTLGeneric
	Ttl int `json:"ttl"`
}

func (p ProductLegacy) SetTtl(ttl int) interface{} {
	p.Ttl = ttl
	return p
}
func (p ProductLegacy) GetTtl() int {
	return p.Ttl
}

// CachetearDown for Products returns the list of SKUs contained within
func (p ProductLegacy) CacheTearDown() (map[string]cache.CacheRegistry, interface{}) {
	tearDown := make(map[string]cache.CacheRegistry)

	/*
		p.CachedMembers = make([]string, 0)

		for _, sku := range p.Skus {
			key := strings.Join([]string{"SkuLegacy", strconv.Itoa(sku.Id)}, ":")
			p.CachedMembers = append(p.CachedMembers, key)
			tearDown[key] = sku
		}
		// Zeroes the attrib so we dont have duplicated caches
		p.Skus = nil
	*/
	return tearDown, p
}

func (p ProductLegacy) CacheToSearch() []string {
	return []string{} //p.CachedMembers
}

//   atrib   ids cache    valor retornado
func (p ProductLegacy) CacheBuildUp(map[string]cache.CacheRegistry) interface{} { // cache devolve para a estrutura os objetos solicitados na funcao acima
	return p
}

// SKU Legacy obj and PartItself interface methods ----------
type SkuLegacy struct {
	Id              int              `json:"id"`
	Name            string           `json:"name"`
	Cached          bool             `json:"cached"`
	CachedMembers   []string         `json:"cachedOffers"`
	CreatedAt       int              `json:"createdAt"`
	FreightTypeId   int              `json:"freightTypeId"`
	Weight          float32          `json:"weight"`
	Height          float32          `json:"height"`
	Width           float32          `json:"width"`
	Eans            []string         `json:"eans"`
	Length          float32          `json:"length"`
	WeightReal      float32          `json:"weightReal"`
	HeightReal      float32          `json:"heightReal"`
	WidthReal       float32          `json:"widthReal"`
	LengthReal      float32          `json:"lengthReal"`
	PostalVolumeFee float32          `json:"postalVolumeFee"`
	ProductId       int              `json:"productId"`
	Product         ProductLegacy    `json:"product"`
	Status          bool             `json:"status"`
	Attributes      Attributes       `json:"attributes"`
	OfferList       SKUOfferListType `json:"offerList"`
	SkuFileList     struct {
		Status     bool      `json:"status"`
		Weight     float32   `json:"weight"`
		WeightReal float32   `json:"weightReal"`
		Width      int       `json:"width"`
		WidthReal  int       `json:"widthReal"`
		SkuFile    []SkuFile `json:"skuFile"`
	} `json:"skuFileList"`
	Ttl int `json:"ttl"`
	//cache.HasTTLGeneric
}

type SKUOfferListType struct {
	Offer        []OfferLegacy `json:"offer"`
	TotalResults int           `json:"total_results"`
}

//define the cachekey for OfferBis struct
func (s SKUOfferListType) GetCacheKey() string {

	if len(s.Offer) > 0 {
		return "SKUOfferList:" + "firstOffer_" + strconv.Itoa(s.Offer[0].Id)
	} else {
		return "SKUOfferList:empty"
	}

}

//define the cachekey for OfferBis struct
func (i SkuLegacy) GetCacheKey() string {
	return "SkuLegacy:" + strconv.Itoa(i.Id)
}

func (s SkuLegacy) SetTtl(ttl int) interface{} {
	s.Ttl = ttl
	return s
}
func (s SkuLegacy) GetTtl() int {
	return s.Ttl
}

// CacheTearDown for SKUs involves getting th eparent product entity
// and its list of Offers
func (s SkuLegacy) CacheTearDown() (map[string]cache.CacheRegistry, interface{}) {

	tearDown := make(map[string]cache.CacheRegistry)

	return tearDown, s /*
		s.CachedMembers = make([]string, 0)

		// Add parent product
		keyProduct := strings.Join([]string{"ProductLegacy", strconv.Itoa(s.Product.Id)}, ":")
		tearDown[keyProduct] = s.Product
		s.CachedMembers = append(s.CachedMembers, keyProduct)
		s.Product = ProductLegacy{}

		// Get the offers
		for _, offer := range s.OfferList.Offer {
			keyOffer := strings.Join([]string{"OfferLegacy", strconv.Itoa(offer.Id)}, ":")
			tearDown[keyOffer] = offer
			s.CachedMembers = append(s.CachedMembers, keyOffer)
		}
		s.OfferList.Offer = nil
	*/
}

func (s SkuLegacy) CacheToSearch() []string {
	//return s.CachedMembers
	return []string{}
}

//   atrib   ids cache    valor retornado
func (s SkuLegacy) CacheBuildUp(map[string]cache.CacheRegistry) interface{} { // cache devolve para a estrutura os objetos solicitados na funcao acima
	return s
	// Todo: write me
}

// Offer Legacy obj and PartItself interface methods ----------
type OfferLegacy struct {
	Id                    int     `json:"id"`
	CreatedAt             int     `json:"createdAt"`
	ImportTax             float32 `json:"importTax"`
	LastUpdatedAt         int     `json:"lastUpdatedAt"`
	LastUpdatedByIp       string  `json:"lastUpdatedByIp"`
	LastUpdatedByUser     string  `json:"lastUpdatedByUser"`
	ListPrice             float32 `json:"listPrice"`
	Price                 float32 `json:"price"`
	Quantity              int     `json:"quantity"`
	RequestedUpdateDate   int     `json:"requestedUpdateDate"`
	SellerExternalOfferId string  `json:"sellerExternalOfferId"`
	SellerId              string  `json:"sellerId"`
	SkuId                 int     `json:"skuId"`
	Status                bool    `json:"status"`
	//StatusWalmartMarketPlace bool    `json:"-"`
	//	Store                 string   `json:"store"`
	StoreId int `json:"storeId"`
	Version int `json:"version"`
	Seller  struct {
		Cnpj        string `json:"cnpj"`
		Description string `json:"description"`
		Email       string `json:"email"`
		Id          string `json:"id"`
		Name        string `json:"name"`
		Status      bool   `json:"status"`
	} `json:"seller"`
	Sku SkuLegacy `json:"sku"`
	//cache.HasTTLGeneric
	Ttl int `json:"-"`
}

//define the cachekey for OfferBis struct
func (o OfferLegacy) GetCacheKey() string {
	return "OfferLegacy:" + strconv.Itoa(o.Id)
}

// Offer Collection definition for sortable interface
type Offers []OfferLegacy

func (o OfferLegacy) SetTtl(ttl int) interface{} {
	o.Ttl = ttl
	return o
}

func (o OfferLegacy) GetTtl() int {
	return o.Ttl
}

// CacheTearDown for Offers returns the parent SKU of the Offer
func (o OfferLegacy) CacheTearDown() (map[string]cache.CacheRegistry, interface{}) {
	tearDown := make(map[string]cache.CacheRegistry)
	/*
		o.CachedMembers = make([]string, 0)

		// Add parent SKU
		keySku := strings.Join([]string{"skuLegacy", strconv.Itoa(o.Sku.Id)}, ":")
		tearDown[keySku] = o.Sku
		o.CachedMembers = append(o.CachedMembers, keySku)
		o.Sku = SkuLegacy{}
	*/
	return tearDown, o
}

func (o OfferLegacy) CacheToSearch() []string {
	//return o.CachedMembers
	return []string{}
}

//   atrib   ids cache    valor retornado
func (o OfferLegacy) CacheBuildUp(map[string]cache.CacheRegistry) interface{} { // cache devolve para a estrutura os objetos solicitados na funcao acima
	return o
}

// -- component structures --

type SkuFile struct {
	File struct {
		FileFormat struct {
			Extension string `json:"extension"`
			HtmlTag   string `json:"htmlTag"`
			Id        int    `json:"id"`
			MimeType  string `json:"mimeType"`
			Name      string `json:"name"`
		} `json:"fileFormat"`
		FileFormatId int    `json:"fileFormatId"`
		FileLocation string `json:"fileLocation"`
		FileType     struct {
			Height             int    `json:"height"`
			Id                 int    `json:"id"`
			Name               string `json:"name"`
			RequiredResizeFlag bool   `json:"requiredResizeFlag"`
			Size               int    `json:"size"`
			Type               string `json:"type"`
			Width              int    `json:"width"`
		} `json:"fileType"`
		FileTypeId        int     `json:"fileTypeId"`
		Height            float32 `json:"height"`
		HeightMeasureUnit string  `json:"heightMeasureUnit"`
		Id                int     `json:"id"`
		Name              string  `json:"name"`
		Width             float32 `json:"width"`
		WidthMeasureUnit  string  `json:"widthMeasureUnit"`
	} `json:"file"`
	FileId     int    `json:"fileId"`
	Id         int    `json:"id"`
	IdFileType int    `json:"idFileType"`
	IdSku      int    `json:"idSku"`
	Label      string `json:"label"`
	Main       bool   `json:"main"`
	Status     bool   `json:"status"`
	Tag        string `json:"tag"`
}

// Colletion to implemment the Sorting interface
type Attributes []AttributesLegacy

type AttributesLegacy struct {
	ReferenceCode string `json:"referenceCode"`
	Field         struct {
		DisplayedAsFilter        bool   `json:"displayedAsFilter"`
		DisplayedAsSpecification bool   `json:"displayedAsSpecification"`
		FromSku                  bool   `json:"fromSku"`
		Id                       int    `json:"id"`
		Name                     string `json:"name"`
		Priority                 int    `json:"priority"`
		Required                 bool   `json:"required"`
		SideMenu                 bool   `json:"sideMenu"`
		Status                   bool   `json:"status"`
		TopMenu                  bool   `json:"topMenu"`
		TypeId                   int    `json:"typeId"`
		Wizard                   bool   `json:"wizard"`
		Selector                 bool   `json:"selector"`
	} `json:"field"`
	FieldId int    `json:"fieldId"`
	Id      int    `json:"id"`
	Value   string `json:"value"`
}

// Colletion to implemment the Sorting interface
type FieldGroups []FieldGroupLegacy

type FieldGroupLegacy struct {
	Id         int                  `json:"id"`
	Name       string               `json:"name"`
	Priority   int                  `json:"priority"`
	Attributes FieldGroupAttributes `json:"attributes"`
}

// Colletion to implemment the Sorting interface
type FieldGroupAttributes []FieldGroupAttributeLegacy

type FieldGroupAttributeLegacy struct {
	Field struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
		Type struct {
			ID         int    `json:"id"`
			Name       string `json:"name"`
			Hasoptions bool   `json:"hasOptions"`
			Status     bool   `json:"status"`
		} `json:"type"`
		Typeid                   int  `json:"typeId"`
		Priority                 int  `json:"priority"`
		Required                 bool `json:"required"`
		DisplayedAsSpecification bool `json:"displayedAsSpecification"`
		DisplayedAsFilter        bool `json:"displayedAsFilter"`
		Wizard                   bool `json:"wizard"`
		Topmenu                  bool `json:"topMenu"`
		Sidemenu                 bool `json:"sideMenu"`
		Status                   bool `json:"status"`
		Fromsku                  bool `json:"fromSku"`
		Selector                 bool `json:"selector"`
	} `json:"field"`
	Values []*AttrValues `json:"values"`
}

type AttrValues struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type CategoryLegacy struct {
	Parent *CategoryLegacy `json:"parent"`
	//Children           []interface{}   `json:"children"`
	Flagcategorymirror bool        `json:"flagCategoryMirror"`
	Id                 int         `json:"id"`
	Iddepartment       int         `json:"idDepartment"`
	Name               string      `json:"name"`
	Fieldgroups        FieldGroups `json:"fieldGroups"`
	Status             bool        `json:"status"`
}

// Sort Interface Impls ----

// FieldGroups Sorting
func (slice FieldGroups) Len() int {
	return len(slice)
}

func (slice FieldGroups) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (slice FieldGroups) Less(i, j int) bool {
	if slice[i].Priority == slice[j].Priority {
		return slice[i].Id <= slice[j].Id
	}
	return slice[i].Priority <= slice[j].Priority
}

// Attributes from a FieldGroup Sorting
func (slice FieldGroupAttributes) Len() int {
	return len(slice)
}

func (slice FieldGroupAttributes) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (slice FieldGroupAttributes) Less(i, j int) bool {
	if slice[i].Field.Priority == slice[j].Field.Priority {
		return slice[i].Field.Id <= slice[j].Field.Id
	}
	return slice[i].Field.Priority <= slice[j].Field.Priority
}

// AttributeLegacy Sorting
func (slice Attributes) Len() int {
	return len(slice)
}

func (slice Attributes) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (slice Attributes) Less(i, j int) bool {
	if slice[i].Field.Priority == slice[j].Field.Priority {
		return slice[i].Field.Id <= slice[j].Field.Id
	}
	return slice[i].Field.Priority <= slice[j].Field.Priority
}

// OfferLegacy Sorting
func (slice Offers) Len() int {
	return len(slice)
}

func (slice Offers) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (slice Offers) Less(i, j int) bool {
	if slice[i].Price == slice[j].Price {
		return slice[i].ListPrice <= slice[j].ListPrice
	}
	return slice[i].Price <= slice[j].Price
}

// Consistency/Status Checking ----
