package schema

//check SKUs and Offers associated to this product
func (p ProductLegacy) IsValid() error {
	//product is active
	if !p.Status {
		logger.Error("Product is not active! %v", p.Id)
		return BIROError{"Product is not active!", nil, 412}
	}

	//has some variation
	if len(p.Skus) == 0 {
		logger.Error("Product has no item varations! %v", p.Id)
		return BIROError{"Product has no item varations!", nil, 412}
	}

	//variation is valid
	for _, sku := range p.Skus {
		err := sku.IsValid()
		if err == nil { //if only one sku is valid, it is ok for product tree at all
			return nil
		}
	}

	logger.Error("Product has no valid SKUs! %v", p.Id)
	return BIROError{"Product has no valid SKUs!", nil, 412}
}

// check SKUs validity
func (s SkuLegacy) IsValid() error {

	if !s.Status {
		logger.Debug("SKU is not active! %v", s.Id)
		return BIROError{"Item variation is not active!", nil, 412}
	}

	//  Deprecated: Offerless Variations is indeed valid

	//	if s.OfferList.Offer == nil || len(s.OfferList.Offer) == 0 {
	//		logger.Error("SKU is not valid! Has no offers!")
	//		return BIROError{"Item variation has no offers!", nil, 412}
	//	}

	//	for _, offer := range s.OfferList.Offer {
	//		err := offer.IsValid()
	//		if err == nil { // if some offer is valid, it is ok for sku at all
	//			return nil
	//		}
	//	}

	//	logger.Error("SKU has no valid offers! %v", s.Id)
	//	return BIROError{"SKU has no valid offers!", nil, 412}
	return nil
}

// check Offer validity
func (o OfferLegacy) IsValid() error {

	if !o.Status {
		logger.Debug("Offer is not active! %v. Status %v, statusSeller %v", o.Id, o.Status)
		return BIROError{"Offer is not active!", nil, 412}
	}

	if !o.Seller.Status {
		logger.Debug("Seller is not active! Offer: %v", o.Id)
		return BIROError{"Seller is not active!", nil, 412}
	}
	return nil
}
