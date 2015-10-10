package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"fmt"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"gitlab.wmxp.com.br/bis/biro/cache"
	"gitlab.wmxp.com.br/bis/biro/legacy"
	"gitlab.wmxp.com.br/bis/biro/schema"
	"gitlab.wmxp.com.br/bis/biro/utilsbiro"
	"net/url"
	"strings"
	"time"
)

//contrato generico de find checkout

type FindGenericType func(w http.ResponseWriter, r *http.Request, paramIds string, transformerFinder utilsbiro.TransformerFinder, legacyFinder utilsbiro.TypeGenericSourceFinder)

type CacheKeyBuilder func(id int) string

var (
	IsBIROLive = true

	MaxMultiGet     = conf.ConfigInt("MaxMultiGet", "40")
	MultiGetTimeout = conf.ConfigInt("MultiGetTimeout", "30000")

	log = logging.MustGetLogger("biro")
)

type MultigetResponseItem struct {
	Order       int         `json:"-"`
	RequestedId string      `json:"requestedId"`
	StatusCode  int         `json:"statusCode"`
	TTL         int         `json:"ttl"`
	Message     string      `json:"message"`
	Payload     interface{} `json:"payload,omitempty"`
}

//health check endpoint
func Healthcheck(w http.ResponseWriter, r *http.Request) {
	if IsBIROLive {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("LIVE"))

	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("DEAD"))
	}
}

//UpdateCache
func UpdateItemV1Cache(w http.ResponseWriter, r *http.Request) {
	defer func() { //assure for not panicking
		if rec := recover(); rec != nil {
			WriteResponseAsMessage(w, r, 500, fmt.Sprintf("Error trying to update cache of itemv1! a %v", rec))
			log.Error(fmt.Sprintf("Recover! Error trying to update cache of item v1! %v", rec))
		}
	}()

	itemId, errParse := ParamToInt("itemid", w, r)
	if errParse != nil {
		WriteResponseAsMessage(w, r, 500, fmt.Sprintf("Error trying parse id of request x %v", errParse))
		log.Error(fmt.Sprintf("Error trying parse id of request c %v", errParse), errParse)
		return
	}

	var itemv1 schema.ItemV1
	d := json.NewDecoder(r.Body)

	err := d.Decode(&itemv1)
	if err != nil {

		WriteResponseAsMessage(w, r, 500, fmt.Sprintf("Error trying to update cache of itemv1! f %v", err))
		log.Error(fmt.Sprintf("Error trying to update cache of itemv1! d"), err)

	} else {
		if itemId != itemv1.Id {
			WriteResponseAsMessage(w, r, 400, fmt.Sprintf("The endpoint item id is not the same ID of item in request body!"))
			log.Error(fmt.Sprintf("The endpoint item id is not the same ID of item in request body! %v %v ", itemId, itemv1.Id))

		} else {

			err := legacy.AngusServices.UpdateItemV1Cache(itemId, itemv1)
			if err == nil {
				writeResponseAsJSON(w, r, fmt.Sprintf("Cache was successfully updated!"), nil)
				log.Debug("Cache was successfully updated!")

			} else {
				WriteResponseAsMessage(w, r, 500, fmt.Sprintf("Error trying to update cache! err: %v", err))
				log.Error("Error trying to update cache! ", err)

			}
		}
	}
}


//UpdateCache
func UpdateItemCache(w http.ResponseWriter, r *http.Request) {
	defer func() { //assure for not panicking
		if rec := recover(); rec != nil {
			WriteResponseAsMessage(w, r, 500, fmt.Sprintf("Error trying to update cache! a %v", rec))
			log.Error(fmt.Sprintf("Recover! Error trying to update cache! b", rec))
		}
	}()

	itemId, errParse := ParamToInt("itemid", w, r)
	if errParse != nil {
		WriteResponseAsMessage(w, r, 500, fmt.Sprintf("Error trying parse id of request x %v", errParse))
		log.Error(fmt.Sprintf("Error trying parse id of request c %v", errParse), errParse)
		return
	}

	var product schema.ProductLegacy
	d := json.NewDecoder(r.Body)

	err := d.Decode(&product)
	if err != nil {

		WriteResponseAsMessage(w, r, 500, fmt.Sprintf("Error trying to update cache! f %v", err))
		log.Error(fmt.Sprintf("Error trying to update cache! d"), err)

	} else {
		if itemId != product.Id {
			WriteResponseAsMessage(w, r, 400, fmt.Sprintf("The endpoint item id is not the same ID of item in request body!"))
			log.Error(fmt.Sprintf("The endpoint item id is not the same ID of item in request body! %v %v ", itemId, product.Id))

		} else {

			err := legacy.AngusServices.UpdateItemCache(itemId, product)
			if err == nil {
				writeResponseAsJSON(w, r, fmt.Sprintf("Cache was successfully updated!"), nil)
				log.Debug("Cache was successfully updated!")

			} else {
				WriteResponseAsMessage(w, r, 500, fmt.Sprintf("Error trying to update cache! err: %v", err))
				log.Error("Error trying to update cache! ", err)

			}
		}
	}
}

//UpdateCache
func UpdateItemOfferV1Cache(w http.ResponseWriter, r *http.Request) {
	defer func() { //assure for not panicking
		if rec := recover(); rec != nil {
			WriteResponseAsMessage(w, r, 500, fmt.Sprintf("Error trying to update cache! %v", rec))
		}
	}()

	offerId, errParse := ParamToInt("offerid", w, r)
	if errParse != nil {
		WriteResponseAsMessage(w, r, 500, fmt.Sprintf("Error trying parse id of request %v", errParse))
		return
	}

	var itemOfferV1 schema.ItemOfferV1
	d := json.NewDecoder(r.Body)

	err := d.Decode(&itemOfferV1)
	if err != nil {
		WriteResponseAsMessage(w, r, 500, fmt.Sprintf("Error trying to update cache! %v", err))

	} else {

		if offerId != itemOfferV1.Id {
			WriteResponseAsMessage(w, r, 400, fmt.Sprintf("The endpoint offer id is not the same ID of offer in request body!"))
		} else {

			err := legacy.AngusServices.UpdateItemV1OfferCache(offerId, itemOfferV1)
			if err == nil {
				writeResponseAsJSON(w, r, fmt.Sprintf("Cache was successfully updated!"), nil)

			} else {
				WriteResponseAsMessage(w, r, 500, fmt.Sprintf("Error trying to update cache! err: %v", err))

			}
		}
	}

}


//UpdateCache
func UpdateOfferCache(w http.ResponseWriter, r *http.Request) {
	defer func() { //assure for not panicking
		if rec := recover(); rec != nil {
			WriteResponseAsMessage(w, r, 500, fmt.Sprintf("Error trying to update cache! %v", rec))
		}
	}()

	offerId, errParse := ParamToInt("offerid", w, r)
	if errParse != nil {
		WriteResponseAsMessage(w, r, 500, fmt.Sprintf("Error trying parse id of request %v", errParse))
		return
	}

	var offer schema.OfferLegacy
	d := json.NewDecoder(r.Body)

	err := d.Decode(&offer)
	if err != nil {
		WriteResponseAsMessage(w, r, 500, fmt.Sprintf("Error trying to update cache! %v", err))

	} else {

		if offerId != offer.Id {
			WriteResponseAsMessage(w, r, 400, fmt.Sprintf("The endpoint offer id is not the same ID of offer in request body!"))
		} else {

			err := legacy.AngusServices.UpdateOfferCache(offerId, offer)
			if err == nil {
				writeResponseAsJSON(w, r, fmt.Sprintf("Cache was successfully updated!"), nil)

			} else {
				WriteResponseAsMessage(w, r, 500, fmt.Sprintf("Error trying to update cache! err: %v", err))

			}
		}
	}

}

//kill BIRO, setting flag to false
func KillBIRO(w http.ResponseWriter, r *http.Request) {
	IsBIROLive = false
	Healthcheck(w, r)
}

//Resurrect BIRO, setting flag to true
func ResurrectBIRO(w http.ResponseWriter, r *http.Request) {
	IsBIROLive = true
	Healthcheck(w, r)
}

func FindGenericMultGet(w http.ResponseWriter, r *http.Request, paramIds string, transformerFinder utilsbiro.TransformerFinder, legacyFinder utilsbiro.TypeGenericSourceFinder) {

	//protection defer. Avoid panic
	defer func(w http.ResponseWriter, req *http.Request) {
		if r := recover(); r != nil {
			log.Error("Recovered in FindGenericMultGet! %v", r)
			WriteResponseAsMessage(w, req, http.StatusInternalServerError, "Not expected Error!")
		}
	}(w, r)

	//try to parse param
	ids, err := ParamToArray(paramIds, w, r)

	if err != nil {
		WriteResponseAsMessage(w, r, http.StatusBadRequest, err)
		return //error already treated in paramToInt. Some error code and message was written
	}

	//version control
	version := getVersion(r) // TODO: review version mechanism

	respCha := make(chan MultigetResponseItem) // found itens

	for order, idStringo := range ids {

		go func(idString string, order int) {
			id, err := strconv.ParseInt(idString, 10, 0)

			if err != nil {
				log.Error("Error trying to format list of ids to int values! %v", err)
				respCha <- MultigetResponseItem{
					Order:       order,
					RequestedId: idString,
					StatusCode:  400,
					TTL:         -1,
					Message:     "Error trying to parse parameter to int value!",
					Payload:     nil,
				}
				return
			}

			r.ParseForm()

			//call finder that retrieve object as struct, status and error
			someInterface, found, err := transformerFinder(int(id), toMap(r.Form), version, legacyFinder)
			var responseWrap MultigetResponseItem

			if err != nil {

				switch err := err.(type) {
				case schema.BIROError:
					responseWrap = MultigetResponseItem{
						Order:       order,
						RequestedId: idString,
						StatusCode:  err.Code,
						TTL:         GetTTL(someInterface),
						Message:     err.Error(),
						Payload:     someInterface,
					}
				case error:
					responseWrap = MultigetResponseItem{
						Order:       order,
						RequestedId: idString,
						StatusCode:  500,
						TTL:         -1,
						Message:     err.Error(),
						Payload:     nil,
					}
				}

			} else {

				if found {
					responseWrap = MultigetResponseItem{
						Order:       order,
						RequestedId: idString,
						StatusCode:  getCode(someInterface),
						TTL:         GetTTL(someInterface),
						Message:     "OK",
						Payload:     someInterface,
					}
				} else {
					responseWrap = MultigetResponseItem{
						Order:       order,
						RequestedId: idString,
						StatusCode:  404,
						TTL:         -1,
						Message:     "Not found!",
						Payload:     nil,
					}
				}

			}

			respCha <- responseWrap

		}(idStringo, order)
	}

	responseArray := make([]MultigetResponseItem, len(ids))

	for i := 0; i < len(ids); i++ {

		select {
		case responseWrap := <-respCha:

			responseArray[responseWrap.Order] = responseWrap

		case <-time.After(time.Millisecond * time.Duration(MultiGetTimeout)):
			//timeout controll
			writeResponse(w, r, "Internal error!", false, schema.BIROError{"Operation timeout!", nil, 408})
			return
		}
	}

	//write final response into http.Request
	writeResponse(w, r, responseArray, true, nil) // return true for "found" param in multi get as default
}

//find generico para objeto checkout
func FindGeneric(w http.ResponseWriter, r *http.Request, paramId string, transformerFinder utilsbiro.TransformerFinder, legacyFinder utilsbiro.TypeGenericSourceFinder) {

	//protection defer. Avoid panic
	defer func(w http.ResponseWriter, req *http.Request) {
		if r := recover(); r != nil {
			log.Error("Recovered in FindGeneric! %v", r)
			WriteResponseAsMessage(w, req, http.StatusInternalServerError, "Not expected Error!")
		}
	}(w, r)

	//try to parse param
	id, err := ParamToInt(paramId, w, r)
	if err != nil {
		WriteResponseAsMessage(w, r, http.StatusBadRequest, "Error trying to parse int param!")
		//error already treated in paramToInt. Some error code and message was written
		return
	}

	//version control
	version := getVersion(r) // TODO: review version mechanism

	errParse := r.ParseForm()
	if errParse != nil {
		WriteResponseAsMessage(w, r, http.StatusInternalServerError, "Not expected Error! "+errParse.Error())
		return
	}

	//call finder that retrive object as struct, status and error
	someInterface, found, err := transformerFinder(id, toMap(r.Form), version, legacyFinder)

	//write final response into http.Request
	writeResponse(w, r, someInterface, found, err)

}

//TODO: CHANGE ALL DEPENDENCIES TO url.Values, not map[string]string
func toMap(form url.Values) map[string]string {

	res := make(map[string]string)

	for c, v := range form {
		if len(v) > 0 {

			res[c] = v[0]
		}
	}

	return res
}

//write response into http.Request
func writeResponse(w http.ResponseWriter, r *http.Request, someInterface interface{}, found bool, err error) {

	//switch for
	switch errCaptured := err.(type) {
	case schema.BIROError:
		//FOUND A ERROR TREATED BY BIRO. SET HTTP STATUS CODE AS SPECIFIED IN BIROError.Code
		log.Debug("BIRO Error to be treated! Maybe some business inconsistency x")

		if errCaptured.Code == 412 {
			//expected biro error with status code
			writeResponseAsJSON(w, r, someInterface, &errCaptured)
		} else {
			//other not expected code
			WriteResponseAsMessage(w, r, http.StatusInternalServerError, "Unexpected error! "+errCaptured.Error())
		}

	case error:
		//found some not expected error. Will write a generic error message!
		log.Error("Error unexpected!")
		WriteResponseAsMessage(w, r, http.StatusInternalServerError, "Unexpected error! "+errCaptured.Error())

	case nil:
		//GREAT! NO error found. Write
		if found {
			log.Debug("Error=nil, found=true")
			writeResponseAsJSON(w, r, someInterface, nil)

		} else {
			// write 404 - not found
			log.Debug("Error=nil, found=false")
			WriteResponseAsMessage(w, r, http.StatusNotFound, "Not found!")
		}

	default:
		// Return as http status 500 as default
		log.Error("Default at Select")
		WriteResponseAsMessage(w, r, http.StatusInternalServerError, "Unexpected error! ")
	}
}

//return 200 as default
func getCode(some interface{}) int {
	switch holdStatusCode := some.(type) {
	case schema.HoldStatusCode:
		return holdStatusCode.GetStatusCode()

	case *schema.HoldStatusCode:
		return (*holdStatusCode).GetStatusCode()

	default:
		return 200
	}
}

//write response as json object writted on http.response
func writeResponseAsJSON(w http.ResponseWriter, r *http.Request, responseObject interface{}, biroError *schema.BIROError) {
	writeHeader(w, responseObject, biroError)

	var statusCode int

	if biroError != nil {
		log.Error("HTTP status code %v, URL %v%v", biroError.Code, r.URL.Host, r.URL.Path)
		statusCode = biroError.Code
	} else {
		statusCode = getCode(responseObject)
	}

	w.WriteHeader(statusCode)

	e := json.NewEncoder(w)
	e.Encode(responseObject)
}

//Write error object as a json output, sets the content-type accordingly
func WriteResponseAsMessage(w http.ResponseWriter, r *http.Request, httpErrorCode int, responseObject interface{}) {
	e := json.NewEncoder(w)

	w.Header().Add("Content-Type", "application/json")
	w.Header().Del("Cache-Control")

	w.WriteHeader(httpErrorCode)

	e.Encode(struct{ Error interface{} }{responseObject})
}

//write all about header
func writeHeader(w http.ResponseWriter, someInterface interface{}, err error) {

	//generic header information from interface
	headerTyped, isHeaderTyped := someInterface.(schema.HeaderTyped)
	if isHeaderTyped {
		for name, value := range headerTyped.GetHeaders() {
			w.Header().Set(name, value)
		}
	}

	//get ttl from object
	ttl := GetTTL(someInterface)

	//header about ttl
	putTTLHeader(ttl, &w)

}

//extract ttl from object, if object is a cache.ExposeTTL
func GetTTL(someInterface interface{}) int {

	exposeTTL, isExposeTTL := someInterface.(cache.ExposeTTL)

	ttl := -1
	if isExposeTTL {
		ttl = exposeTTL.GetTtl()
	}

	return ttl
}

//check whether implements cache.DefineTTL and contains a defined ttl value
func putTTLHeader(ttl int, w *http.ResponseWriter) {
	log.Debug("Setting TTL in header =" + strconv.Itoa(ttl))
	(*w).Header().Add("Cache-Control", "public, max-age="+strconv.Itoa(ttl))
}

func getVersion(r *http.Request) string {
	return ""
}

//parse the informed param to integer
func ParamToInt(paramName string, w http.ResponseWriter, r *http.Request) (int, error) {
	vars := mux.Vars(r)
	strId := vars[paramName]
	intId, err := strconv.ParseInt(strId, 10, 0)

	if err != nil {
		return 0, err
	}
	return int(intId), nil
}

//parse the informed param to integer
func ParamToArray(paramName string, w http.ResponseWriter, r *http.Request) ([]string, error) {

	var mPut map[string][]string
	d := json.NewDecoder(r.Body)
	d.Decode(&mPut)

	log.Debug("map of ids %v:", mPut)

	var arrayIds []string

	if ids, hasIds := mPut["ids"]; hasIds {
		arrayIds = ids
		log.Debug("Processing REST/POST: Taking %v posted ids to search!", len(arrayIds))

	} else {

		vars := mux.Vars(r)
		strIds := vars[paramName]
		arrayIds = strings.Split(strIds, ",")
		log.Debug("Processing REST/GET: Taking %v ids to search from path param!", len(arrayIds))
	}

	if len(arrayIds) > MaxMultiGet {
		return []string{}, schema.BIROError{Message: "Bad Request! The number of requested ids is more than permitted! (" + strconv.Itoa(MaxMultiGet) + " items)", Code: 400, Parent: nil}

	}

	return arrayIds, nil

}
