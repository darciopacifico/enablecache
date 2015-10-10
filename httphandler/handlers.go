package httphandler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"gitlab.wmxp.com.br/bis/biro/cache"
	"gitlab.wmxp.com.br/bis/biro/schema"
)

var (
	log = logging.MustGetLogger("biro")
)

//write response into http.Request
func writeResponse(w http.ResponseWriter, r *http.Request, someInterface interface{}, found bool, err error) {

	//switch for
	switch errCaptured := err.(type) {
	case schema.BIROError:
		//FOUND A ERROR TREATED BY BIRO. SET HTTP STATUS CODE AS SPECIFIED IN BIROError.Code
		log.Debug("BIRO Error to be treated! Maybe some business inconsistency")
		writeResponseAsJSON(w, r, someInterface, &errCaptured)

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

//write response as json object writted on http.response
func writeResponseAsJSON(w http.ResponseWriter, r *http.Request, responseObject interface{}, biroError *schema.BIROError) {
	writeHeader(w, responseObject, biroError)

	if biroError != nil {
		log.Debug("HTTP status code = %v", biroError.Code)
		w.WriteHeader(biroError.Code)
	}

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
	log.Debug("Looking for parameter %v at url!", paramName)

	strId := vars[paramName]

	log.Debug("Value encountered %v!", strId)

	intId, err := strconv.ParseInt(strId, 10, 0)

	log.Debug("values as int %v", intId)

	if err != nil {
		return 0, err
	}
	return int(intId), nil
}
