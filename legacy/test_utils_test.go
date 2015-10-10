package legacy

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"regexp"
)

//mock http requisition.
func ExecuteRequestMock(client *http.Client, request *http.Request) (*http.Response, error) {
	log.Debug(" ===== EXECUTOR DE REST MOCKADO =====")

	resp := http.Response{}
	resp.Status = "success"
	resp.StatusCode = 200

	matchesP, _ := regexp.MatchString("/ws/products/", request.URL.Path)
	matchesS, _ := regexp.MatchString("/ws/skus/", request.URL.Path)
	matchesO, _ := regexp.MatchString("/ws/offers/", request.URL.Path)

	switch {
	case matchesP:
		resp.Body = ioutil.NopCloser(bytes.NewReader([]byte(productResp)))

	case matchesS:
		resp.Body = ioutil.NopCloser(bytes.NewReader([]byte(skusResp)))

	case matchesO:
		resp.Body = ioutil.NopCloser(bytes.NewReader([]byte(offerResp)))

	default:
		resp.Body = ioutil.NopCloser(bytes.NewReader([]byte("PAGINA NAO ENCONTRADA")))
		resp.StatusCode = 404
		resp.Status = "pagina nao encontrada"
	}

	return &resp, nil
}
