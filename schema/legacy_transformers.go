package schema

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/op/go-logging"
	"gitlab.wmxp.com.br/bis/biro/config"
)

// Images URL Prefix for Assets

var (
	conf            = config.CreateConfig()
	logger          = logging.MustGetLogger("biro")
	ImagesUrlPrefix = conf.Config("imagesUrlPrefix", "http://static.wmobjects.com.br/imgres/arquivos/ids/")
)

func (p ProductLegacy) ToItemV1() *ItemV1 {

	i := &ItemV1{}
	i.Headers = make(map[string]string)
	i.Headers["Content-Type"] = "application/vnd.walmart.item.v1+json; charset=utf-8"

	i.Id = p.Id
	i.Name = p.Name
	i.Description = p.Description
	i.Keywords = strings.Split(p.KeyWords, ", ")
	i.Brand.Name = p.Brand.Name
	i.Brand.Id = p.Brand.Id
	i.LinkText = p.LinkText

	categories := make([]Category, 0)
	categories = transformCategories(&p.Category, categories)
	i.Categories = categories

	i.Attributes, i.Filters = transformFieldGroupsAttributes(p.FieldGroups)

	// For now Assets consists only of extracted "Product.FieldGroup[?].Field.Name == "URL Vídeo"
	// carried by legacy products
	videos := extractVideoAssetFromFieldGroups(p.FieldGroups)
	if len(videos) > 0 {
		assetsMap := make(map[string][]Asset)
		assetsMap["video"] = videos

		i.Assets = &assetsMap

	}

	i.Variations = transformSkus(p.Skus)

	return i
}

func (s SkuLegacy) ToVariationV1() VariationV1 {

	v := VariationV1{}

	//set header for variation v1
	v.Headers = make(map[string]string)
	v.Headers["Content-Type"] = "application/vnd.walmart.item-variation.v1+json; charset=utf-8"

	//put the variation itself
	v.Variation = transformVariation(s)

	//basic item information
	v.BasicItem.Id = s.ProductId
	v.BasicItem.Name = s.Product.Name

	//some brand information
	v.Brand.Id = s.Product.Brand.Id
	v.Brand.Name = s.Product.Brand.Name

	//category information
	categories := make([]Category, 0)
	categories = transformCategories(&s.Product.Category, categories)
	v.Categories = categories

	//ttl of cache registry
	v.Ttl = s.Ttl

	return v
}

func (p ProductLegacy) ToItemOfferV1() (ItemOfferV1, bool) {

	i := ItemOfferV1{}
	i.Headers = make(map[string]string)
	i.Headers["Content-Type"] = "application/vnd.walmart.item-offer.v1+json; charset=utf-8"

	status := p.Status // by default, consider product status

	i.Id = p.Id
	i.Name = p.Name
	i.Brand.Id = p.BrandId
	i.Brand.Name = p.Brand.Name

	if len(p.Skus) > 0 {
		offerImage := Asset{} //a default initial dummy asset

		for index, skuFile := range p.Skus[0].SkuFileList.SkuFile {
			if skuFile.Main {
				//if is the main image, its ok. Set as image for offer and break the lace.
				offerImage = skuFileImageToAsset(&skuFile)
				break
			}

			if index == 0 {
				//by default, if no main image was found, set the first one as offer image.
				offerImage = skuFileImageToAsset(&skuFile)
			}
		}

		//finally, set offer image...
		i.Image = offerImage
		i.VariationName = p.Skus[0].Name
		i.VariationId = p.Skus[0].Id
		_, _, i.Specializations = transformAttributes(p.Skus[0].Attributes)

		//status
		status = p.Skus[0].Status
	}

	bestOffer, hasSomeOffer := findBestOffer(p.Skus)
	if hasSomeOffer {
		status = bestOffer.Status

		i.Offer.Id = bestOffer.Id
		i.Offer.Price.Current = int(bestOffer.ListPrice * 100)
		i.Offer.Price.Original = int(bestOffer.Price * 100)
		i.Offer.Price.ImportTax = int(bestOffer.ImportTax * 100)
		i.Offer.Seller.Id = bestOffer.Seller.Id
		i.Offer.Seller.Name = bestOffer.Seller.Name
	}

	categories := make([]Category, 0)
	categories = transformCategories(&p.Category, categories)
	i.Categories = categories

	logger.Debug("Product %d, consistency: %b", i.Id, status)
	return i, status
}

func (o OfferLegacy) ToItemOfferV1() (ItemOfferV1, bool) {
	//invert reference to parent->child
	s := o.Sku
	s.OfferList.Offer = []OfferLegacy{o}

	p := s.Product
	p.Skus = []SkuLegacy{s}

	return p.ToItemOfferV1()
}

// -- Helper functions --

// To be used by ItemOffer conversions
func transformSkuFilesImageToAsset(skuFiles []SkuFile) []Asset {
	assets := make([]Asset, 0)

	for _, s := range skuFiles {
		a := skuFileImageToAsset(&s)

		// Main Image must bet the first slice element
		if s.Main {
			assets = append([]Asset{a}, assets...)
		} else {
			assets = append(assets, a)
		}

	}
	return assets
}

// To be used by ItemOffer conversions
func skuFileImageToAsset(s *SkuFile) Asset {
	a := Asset{}
	a.Url = ImagesUrlPrefix + strconv.Itoa(s.FileId)

	extension := Attribute{Name: "extension", Values: []string{s.File.FileFormat.Extension}}
	mimeType := Attribute{Name: "mimeType", Values: []string{s.File.FileFormat.MimeType}}
	height := Attribute{Name: "height", Values: []string{fmt.Sprintf("%.2f", s.File.Height)}}
	width := Attribute{Name: "width", Values: []string{fmt.Sprintf("%.2f", s.File.Width)}}

	a.Attributes = append(a.Attributes, extension)
	a.Attributes = append(a.Attributes, mimeType)
	a.Attributes = append(a.Attributes, height)
	a.Attributes = append(a.Attributes, width)

	return a
}

// To be used by ItemOffer conversions
func findBestOffer(skus []SkuLegacy) (*OfferLegacy, bool) {

	if len(skus) > 0 && len(skus[0].OfferList.Offer) > 0 {
		bestOffer := skus[0].OfferList.Offer[0]
		for _, sku := range skus {
			for _, offer := range sku.OfferList.Offer {
				if offer.Price < bestOffer.Price {
					bestOffer = offer
				}
			}
		}

		return &bestOffer, true

	} else {
		return nil, false
	}

}

func transformSkus(legacySKUs []SkuLegacy) []Variation {
	v := make([]Variation, 0)
	for _, sku := range legacySKUs {

		// Only add valid SKUs
		if err := sku.IsValid(); err == nil {
			v = append(v, transformVariation(sku))

		}
	}
	return v
}

func transformVariation(sku SkuLegacy) Variation {

	newV := Variation{}

	newV.Id = sku.Id
	newV.Name = sku.Name

	newV.Dimensions = make(map[string]float32, 4)
	newV.Dimensions["height"] = sku.Height
	newV.Dimensions["length"] = sku.Length
	newV.Dimensions["width"] = sku.Width
	newV.Dimensions["weight"] = sku.Weight

	newV.CodeList = make(map[string][]string, 1)

	if len(sku.Eans) > 0 {
		newV.CodeList["ean"] = sku.Eans
	}

	newV.Attributes, newV.Filters, newV.Specializations = transformAttributes(sku.Attributes)

	newV.Assets = make(map[string][]Asset)
	newV.Assets["images"] = transformSkuFilesImageToAsset(sku.SkuFileList.SkuFile)

	newV.OutOfCatalog = isOutOfCatalog(sku)

	newV.Offers = transformOffers(sku.OfferList.Offer)

	return newV
}

func transformOffers(legacyOffers Offers) []Offer {
	offers := make([]Offer, 0)

	// Best offer (lowest price) first
	sort.Sort(legacyOffers)

	for _, legacy := range legacyOffers {

		// Only add Valid Offers
		if err := legacy.IsValid(); err == nil {
			offers = append(offers, transformOffer(legacy))
		}
	}
	return offers
}

func transformOffer(legacy OfferLegacy) Offer {
	offer := Offer{}
	offer.Id = legacy.Id
	offer.Price.Current = int(legacy.ListPrice * 100)
	offer.Price.Original = int(legacy.Price * 100)
	offer.Price.ImportTax = int(legacy.ImportTax * 100)
	offer.Seller.Id = legacy.Seller.Id
	offer.Seller.Name = legacy.Seller.Name
	offer.Available = legacy.Quantity > 0
	return offer
}

func transformCategories(legacy *CategoryLegacy, categories []Category) []Category {
	if legacy != nil {
		newCat := Category{IdNameTyped{Id: legacy.Id, Name: legacy.Name}}
		categories = append([]Category{newCat}, categories...)
		categories = transformCategories(legacy.Parent, categories)
	}

	return categories
}

// Extracts Attributes from legacy Product's FieldGroups and transforms in IdentAttributes
func transformFieldGroupsAttributes(fieldGroups FieldGroups) ([]IdentAttribute, []IdentAttribute) {
	attributes := make([]IdentAttribute, 0)
	searchFilters := make([]IdentAttribute, 0)

	// Return array must be ordered by Priority,Id -- sort before transformation
	sort.Sort(fieldGroups)
	for _, fg := range fieldGroups {

		legacyAttrCol := fg.Attributes

		// Attributes inside each FieldGroup must also be ordered by Priority,Id
		sort.Sort(legacyAttrCol)
		for _, legacyAttr := range legacyAttrCol {

			if legacyAttr.Field.DisplayedAsFilter || legacyAttr.Field.DisplayedAsSpecification {
				attr := transformFGAttributeToIdAtribute(legacyAttr)
				if legacyAttr.Field.DisplayedAsSpecification {
					attributes = append(attributes, attr)
				}
				if legacyAttr.Field.DisplayedAsFilter {
					searchFilters = append(searchFilters, attr)
				}
			}

		}
	}

	return attributes, searchFilters
}

func transformFGAttributeToIdAtribute(legacyAttr FieldGroupAttributeLegacy) IdentAttribute {
	attr := IdentAttribute{}
	attr.Id = legacyAttr.Field.Id
	attr.Name = legacyAttr.Field.Name
	attr.Values = make([]string, 0)
	for _, value := range legacyAttr.Values {
		attr.Values = append(attr.Values, value.Name)
	}
	return attr
}

func transformAttributes(attributesLegacy Attributes) ([]IdentAttribute, []IdentAttribute, []Attribute) {

	attributes := make([]IdentAttribute, 0)
	searchFilters := make([]IdentAttribute, 0)
	specStaging := make([]IdentAttribute, 0)

	// Return array must be ordered by Priority,FieldId -- sort before transformation
	sort.Sort(attributesLegacy)

	lastFId := -1
	for _, attrLegacy := range attributesLegacy {
		// Repeated FieldID values need to be grouped in a string array
		// Take advantage of the fact that they are already ordered to make it happen
		create := attrLegacy.Field.Id != lastFId

		if attrLegacy.Field.DisplayedAsSpecification {
			attributes = appendValuetoAttr(create, attributes, attrLegacy)
		}
		if attrLegacy.Field.DisplayedAsFilter {
			searchFilters = appendValuetoAttr(create, searchFilters, attrLegacy)
		}

		if attrLegacy.Field.Selector {
			specStaging = appendValuetoAttr(create, specStaging, attrLegacy)
		}

		lastFId = attrLegacy.Field.Id

	}

	// We dont need ID in specializations for grouping anymore. Transform it into a
	// regular idless []Attribute
	specializations := make([]Attribute, len(specStaging))
	for i, a := range specStaging {
		specializations[i] = a.Attribute
	}

	return attributes, searchFilters, specializations
}

func appendValuetoAttr(create bool, slice []IdentAttribute, attr AttributesLegacy) []IdentAttribute {

	if !create {
		if len(slice) > 0 {
			i := len(slice) - 1
			if slice[i].Id == attr.Field.Id {
				slice[i].Values = append(slice[i].Values, attr.Value)
				return slice
			}
		}
	}

	// Stars misaligned! First checking failed, need to create the first element
	slice = append(slice, transformAttributesLegacyToAttribute(attr))

	return slice
}

func transformAttributesLegacyToAttribute(attrLegacy AttributesLegacy) IdentAttribute {
	attr := IdentAttribute{}
	attr.Id = attrLegacy.Field.Id
	attr.Name = attrLegacy.Field.Name
	attr.Values = []string{attrLegacy.Value}
	return attr
}

// Scours legacy Fieldgroups look
// ing for attribute called "URL Vídeo" and add it to
// received assets map under "video" key
// Assumptions: there will be at most 1 attribute with right name; our Asset will have no Attributes
func extractVideoAssetFromFieldGroups(fieldGroups FieldGroups) []Asset {
	for _, fg := range fieldGroups {
		for _, attr := range fg.Attributes {
			if attr.Field.Name == "URL Vídeo" && len(attr.Values) > 0 {
				a := Asset{Url: attr.Values[0].Name} // Ignore attributes
				return []Asset{a}                    // We're done here.
			}
		}
	}
	return []Asset{}
}

// Check conditions for OutOfCatalog bool flag
func isOutOfCatalog(sku SkuLegacy) bool {

	// False if sku has no offers (may have in the future)
	if len(sku.OfferList.Offer) == 0 {
		return false
	}

	// False If at least 1 offer is valid and (quantity > 0 && Seller != 1)
	for _, o := range sku.OfferList.Offer {

		// At least one valid offer with stock == false
		err := o.IsValid()
		if err == nil && (o.Quantity > 0 && o.Seller.Id != "1") {
			return false
		}

		// if any seller has stock, or it  exists in Walmart catalog == false
		if o.Quantity > 0 || o.Seller.Id == "1" {
			return false
		}
	}

	// offer invalida (seller ativo, status ativa),
	// tem offer,  todas as offers nao validas,
	//
	//offer valida= quantity > 0 && seller != "1"
	//há offer

	// All conditions failed, so return OutOfCatalog
	return true
}
