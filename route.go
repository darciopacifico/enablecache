package main

import (
	"fmt"

	"github.com/gorilla/mux"
	"github.com/rcrowley/go-metrics"
	"gitlab.wmxp.com.br/bis/biro/darpa"
	"gitlab.wmxp.com.br/bis/biro/legacy"
	"gitlab.wmxp.com.br/bis/biro/utilsbiro"
	"net/http"
	"time"
)

var (
// DOING OBVIOUS CONVERSION FOR A LAZY LANGUAGE!!

//hot calls
	buscaProductHot = func(id int) (interface{}, bool, error) {
		return legacy.AngusServices.FindProductById(id)
	}
	buscaOfferHot = func(id int) (interface{}, bool, error) {
		return legacy.AngusServices.FindOfferById(id)
	}
	buscaSKUHot = func(id int) (interface{}, bool, error) {
		return legacy.AngusServices.FindSKUById(id)
	}

//cached calls
	buscaProductCached = func(id int) (interface{}, bool, error) {
		return legacy.CachedFindProductById(id)
	}
	buscaOfferCached = func(id int) (interface{}, bool, error) {
		return legacy.CachedFindOfferById(id)
	}
	buscaSKUCached = func(id int) (interface{}, bool, error) {
		return legacy.CachedFindSKUById(id)
	}

//save cache only calls
	buscaProductNoCache = func(id int) (interface{}, bool, error) {
		return legacy.NoCacheFindProductById(id)
	}
	buscaOfferNoCache = func(id int) (interface{}, bool, error) {
		return legacy.NoCacheFindOfferById(id)
	}
	buscaSKUNoCache = func(id int) (interface{}, bool, error) {
		return legacy.NoCacheFindSKUById(id)
	}

	alpha = conf.ConfigFloat("metricsAlpha", "0.015")
	reservoirSize = conf.ConfigInt("metricsReservoirSize", "1024")
)

//define and load routes
func NewRouter() *mux.Router {

	//srv := legacy.AngusServices
	router := mux.NewRouter().StrictSlash(true)
	urlPrefix := conf.Config("legacyCatPrefix", "")
	urlDarpa := conf.Config("legacyCatGCPrefix", "/darpa")
	//urlCTPrefix := conf.Config("cacheTransformer", "/ctdarpa")

	//urlCTDarpaPrefix := conf.Config("cacheTransformer", "/darpact")
	/*
		AppendBIRORoute(router, urlCTPrefix, false, "/item",  FindGeneric, darpa.DarpaTransformer(legacy.NC_ItemTransformer_412), buscaProductNoCache)
		AppendBIRORoute(router, urlCTPrefix, true,  "/item",  FindGeneric, darpa.DarpaTransformer(legacy.C_ItemTransformer_412), buscaProductCached)
	*/

	//===========================================
	//=== Consultas sem promocoes gescom =========
	//===== GET DE PRODUTOS, Ofertas, Skus =========
	AppendBIRORoute(router, urlPrefix, false, "/item", FindGeneric, legacy.NC_ItemTransformer_412, buscaProductNoCache)
	AppendBIRORoute(router, urlPrefix, true, "/item", FindGeneric, legacy.C_ItemTransformer_412, buscaProductCached)
	AppendBIRORoute(router, urlPrefix, false, "/item/offer", FindGeneric, legacy.NC_OfferTransformer_412, buscaOfferNoCache)
	AppendBIRORoute(router, urlPrefix, true, "/item/offer", FindGeneric, legacy.C_OfferTransformer_412, buscaOfferCached)
	AppendBIRORoute(router, urlPrefix, false, "/item/variation", FindGeneric, legacy.NC_SkuTransformer, buscaSKUNoCache)
	AppendBIRORoute(router, urlPrefix, true, "/item/variation", FindGeneric, legacy.C_SkuTransformer, buscaSKUCached)

	//===== MULTI GET DE PRODUTOS, Ofertas e SKU =====
	AppendBIRORoute(router, urlPrefix, false, "/items", FindGenericMultGet, legacy.NC_ItemTransformer, buscaProductNoCache)
	AppendBIRORoute(router, urlPrefix, true, "/items", FindGenericMultGet, legacy.C_ItemTransformer, buscaProductCached)
	AppendBIRORoute(router, urlPrefix, false, "/item/offers", FindGenericMultGet, legacy.NC_OfferTransformer, buscaOfferNoCache)
	AppendBIRORoute(router, urlPrefix, true, "/item/offers", FindGenericMultGet, legacy.C_OfferTransformer, buscaOfferCached)
	AppendBIRORoute(router, urlPrefix, false, "/item/variations", FindGenericMultGet, legacy.NC_SkuTransformer, buscaSKUNoCache)
	AppendBIRORoute(router, urlPrefix, true, "/item/variations", FindGenericMultGet, legacy.C_SkuTransformer, buscaSKUCached)

	//===========================================
	//=== Consultas com promocao gescom =========
	//===== GET DE PRODUTOS, Ofertas, Skus =========
	AppendBIRORoute(router, urlDarpa, false, "/item", FindGeneric, darpa.DarpaTransformer(legacy.NC_ItemTransformer_412), buscaProductNoCache)
	AppendBIRORoute(router, urlDarpa, true, "/item", FindGeneric, darpa.DarpaTransformer(legacy.C_ItemTransformer_412), buscaProductCached)
	AppendBIRORoute(router, urlDarpa, false, "/item/offer", FindGeneric, darpa.DarpaTransformer(legacy.NC_OfferTransformer_412), buscaOfferNoCache)
	AppendBIRORoute(router, urlDarpa, true, "/item/offer", FindGeneric, darpa.DarpaTransformer(legacy.C_OfferTransformer_412), buscaOfferCached)
	AppendBIRORoute(router, urlDarpa, false, "/item/variation", FindGeneric, darpa.DarpaTransformer(legacy.NC_SkuTransformer), buscaSKUNoCache)
	AppendBIRORoute(router, urlDarpa, true, "/item/variation", FindGeneric, darpa.DarpaTransformer(legacy.C_SkuTransformer), buscaSKUCached)

	//===== MULTI GET DE PRODUTOS, Ofertas e SKU =====
	AppendBIRORoute(router, urlDarpa, false, "/items", FindGenericMultGet, darpa.DarpaTransformer(legacy.NC_ItemTransformer), buscaProductNoCache)
	AppendBIRORoute(router, urlDarpa, true, "/items", FindGenericMultGet, darpa.DarpaTransformer(legacy.C_ItemTransformer), buscaProductCached)
	AppendBIRORoute(router, urlDarpa, false, "/item/offers", FindGenericMultGet, darpa.DarpaTransformer(legacy.NC_OfferTransformer), buscaOfferNoCache)
	AppendBIRORoute(router, urlDarpa, true, "/item/offers", FindGenericMultGet, darpa.DarpaTransformer(legacy.C_OfferTransformer), buscaOfferCached)
	AppendBIRORoute(router, urlDarpa, false, "/item/variations", FindGenericMultGet, darpa.DarpaTransformer(legacy.NC_SkuTransformer), buscaSKUNoCache)
	AppendBIRORoute(router, urlDarpa, true, "/item/variations", FindGenericMultGet, darpa.DarpaTransformer(legacy.C_SkuTransformer), buscaSKUCached)

	/*
		//===========================================
		//=== Consultas sem promocoes gescom =========
		//===== GET DE PRODUTOS, Ofertas, Skus =========
		AppendBIRORoute(router, urlPrefix, false, "/item", FindGeneric, srv.ItemTransformer_412, buscaProductNoCache)
		AppendBIRORoute(router, urlPrefix, true, "/item", FindGeneric, srv.ItemTransformer_412, buscaProductCached)
		AppendBIRORoute(router, urlPrefix, false, "/item/offer", FindGeneric, srv.OfferTransformer_412, buscaOfferNoCache)
		AppendBIRORoute(router, urlPrefix, true, "/item/offer", FindGeneric, srv.OfferTransformer_412, buscaOfferCached)
		AppendBIRORoute(router, urlPrefix, false, "/item/variation", FindGeneric, srv.SkuTransformer, buscaSKUNoCache)
		AppendBIRORoute(router, urlPrefix, true, "/item/variation", FindGeneric, srv.SkuTransformer, buscaSKUCached)

		//===== MULTI GET DE PRODUTOS, Ofertas e SKU =====
		AppendBIRORoute(router, urlPrefix, false, "/items", FindGenericMultGet, srv.ItemTransformer, buscaProductNoCache)
		AppendBIRORoute(router, urlPrefix, true, "/items", FindGenericMultGet, srv.ItemTransformer, buscaProductCached)
		AppendBIRORoute(router, urlPrefix, false, "/item/offers", FindGenericMultGet, srv.OfferTransformer, buscaOfferNoCache)
		AppendBIRORoute(router, urlPrefix, true, "/item/offers", FindGenericMultGet, srv.OfferTransformer, buscaOfferCached)
		AppendBIRORoute(router, urlPrefix, false, "/item/variations", FindGenericMultGet, srv.SkuTransformer, buscaSKUNoCache)
		AppendBIRORoute(router, urlPrefix, true, "/item/variations", FindGenericMultGet, srv.SkuTransformer, buscaSKUCached)

		//===========================================
		//=== Consultas com promocao gescom =========
		//===== GET DE PRODUTOS, Ofertas, Skus =========
		AppendBIRORoute(router, urlDarpa, false, "/item", FindGeneric, darpa.DarpaTransformer(srv.ItemTransformer_412), buscaProductNoCache)
		AppendBIRORoute(router, urlDarpa, true, "/item", FindGeneric, darpa.DarpaTransformer(srv.ItemTransformer_412), buscaProductCached)
		AppendBIRORoute(router, urlDarpa, false, "/item/offer", FindGeneric, darpa.DarpaTransformer(srv.OfferTransformer_412), buscaOfferNoCache)
		AppendBIRORoute(router, urlDarpa, true, "/item/offer", FindGeneric, darpa.DarpaTransformer(srv.OfferTransformer_412), buscaOfferCached)
		AppendBIRORoute(router, urlDarpa, false, "/item/variation", FindGeneric, darpa.DarpaTransformer(srv.SkuTransformer), buscaSKUNoCache)
		AppendBIRORoute(router, urlDarpa, true, "/item/variation", FindGeneric, darpa.DarpaTransformer(srv.SkuTransformer), buscaSKUCached)

		//===== MULTI GET DE PRODUTOS, Ofertas e SKU =====
		AppendBIRORoute(router, urlDarpa, false, "/items", FindGenericMultGet, darpa.DarpaTransformer(srv.ItemTransformer), buscaProductNoCache)
		AppendBIRORoute(router, urlDarpa, true, "/items", FindGenericMultGet, darpa.DarpaTransformer(srv.ItemTransformer), buscaProductCached)
		AppendBIRORoute(router, urlDarpa, false, "/item/offers", FindGenericMultGet, darpa.DarpaTransformer(srv.OfferTransformer), buscaOfferNoCache)
		AppendBIRORoute(router, urlDarpa, true, "/item/offers", FindGenericMultGet, darpa.DarpaTransformer(srv.OfferTransformer), buscaOfferCached)
		AppendBIRORoute(router, urlDarpa, false, "/item/variations", FindGenericMultGet, darpa.DarpaTransformer(srv.SkuTransformer), buscaSKUNoCache)
		AppendBIRORoute(router, urlDarpa, true, "/item/variations", FindGenericMultGet, darpa.DarpaTransformer(srv.SkuTransformer), buscaSKUCached)
	*/
	//========= update cache Angus ============
	//feed cache of itemv1 and itemofferv1
	router.
	Methods("PUT").
	Path("/itemv1/{itemid}").
	Handler(http.HandlerFunc(UpdateItemV1Cache))

	router.
	Methods("PUT").
	Path("/itemofferv1/{offerid}").
	Handler(http.HandlerFunc(UpdateItemOfferV1Cache))


	//feed cache of legacy
	router.
	Methods("PUT").
	Path("/item/{itemid}").
	Handler(http.HandlerFunc(UpdateItemCache))

	router.
	Methods("PUT").
	Path("/item/offer/{offerid}").
	Handler(http.HandlerFunc(UpdateOfferCache))

	// ======= GETS de manutencao da aplicacao
	router.
	Methods("GET").
	Path("/healthcheck").
	Name("Healthcheck").
	//Headers("version", "1").
	Handler(http.HandlerFunc(Healthcheck))

	router.
	Methods("GET").
	Path("/killbiro").
	Name("Killbiro").
	Handler(http.HandlerFunc(KillBIRO))

	router.
	Methods("GET").
	Path("/resurrectbiro").
	Name("ResurrectBIRO").
	Handler(http.HandlerFunc(ResurrectBIRO))

	return router
}

//append a route to gorilla mux, using a default route template
func AppendBIRORoute(
router *mux.Router,
legacyCatPrefix string,
cache bool,
path string,
genericFinder FindGenericType,
transformer utilsbiro.TransformerFinder,
businessFinder utilsbiro.TypeGenericSourceFinder) {

	//crates a metrics histogram for biro path
	//histogram := createBiroHistogram(legacyCatPrefix, path, cache)

	paramPath := "ids" //alface

	//post config block
	route_p := router.Methods("POST")

	if !cache {
		route_p = route_p.Headers("Cache-Control", "no-cache")
	}

	route_p = route_p.Path(legacyCatPrefix + path).Handler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			genericFinder(w, r, paramPath, transformer, businessFinder)
		}))

	//get config block
	route_g := router.Methods("GET")
	if !cache {
		route_g = route_g.Headers("Cache-Control", "no-cache")
	}
	fullPath := legacyCatPrefix + path + "/{" + paramPath + "}"

	log.Debug("Creating route for %v", (fullPath))

	route_g = route_g.Path(fullPath).
	Handler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			//take request start time
			start := time.Now()

			genericFinder(w, r, paramPath, transformer, businessFinder)

			//calculates duration
			duration := time.Since(start)
			_ = duration
			//update histogram
			//histogram.Update(int64(duration / time.Millisecond))

		}))
}

//crates a metrics histogram for biro path
func createBiroHistogram(prefix string, path string, cache bool) metrics.Histogram {

	log.Debug("Creating a metrics histogram for path %v%v, cached %v", prefix, path, cache)
	s := metrics.NewExpDecaySample(reservoirSize, alpha)
	h := metrics.NewHistogram(s)
	fullName := fmt.Sprintf("BIRO endpoint path %v%v, cached %v", prefix, path, cache)
	metrics.Register(fullName, h)

	return h
}
