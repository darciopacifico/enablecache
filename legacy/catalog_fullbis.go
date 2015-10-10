package legacy

import (
	"gitlab.wmxp.com.br/bis/biro/httphandler"
	"io"
	"net/http"
	"strconv"
)

func ProxyRepo(w http.ResponseWriter, r *http.Request, reqTemplate FullBisRequestTemplate, paramId string) {

	intId, err := httphandler.ParamToInt(paramId, w, r)
	if err != nil {
		log.Error("invalid parameter %v", paramId, err)
		httphandler.WriteResponseAsMessage(w, r, 400, "Invalid parameter!")
		return
	}
	log.Debug("value of parameter %v, %v", paramId, intId)

	requestURI := reqTemplate.GetURI(map[string]string{"id": strconv.Itoa(intId)})

	reqREST, err := http.NewRequest("GET", requestURI, nil)
	if err != nil {
		log.Error("Erro ao tentar criar objeto de request %v", err)
		httphandler.WriteResponseAsMessage(w, r, 500, "Internal server error!")
		return
	}

	client := &http.Client{}
	resp, err := client.Do(reqREST)
	if err != nil {
		log.Error("Erro ao tentar criar objeto de request %v", err)
		httphandler.WriteResponseAsMessage(w, r, 500, "Internal server error!")
		return
	}

	w.Header().Add("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)

	io.Copy(w, resp.Body)

}
