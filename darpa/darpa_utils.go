package darpa

import "net/http"

//instantiate a request template
var darpaRequest = DarpaRequestTemplate{
	//Uri:  conf.Config("URIRequestDarpa", "http://vip-darpa-offers.qa.vmcommerce.intra/services/offers"), // Darpa VIP estava fora. Seguindo recomendacao do Barenko, apontando direto pro nó 2
	Uri: conf.Config("URIRequestDarpa", "http://napsao-qa-nix-darpa-offers-1.qa.vmcommerce.intra:8080/services/offers"),

	User: conf.Config("userDarpa", "services"),
	Pass: conf.Config("passDarpa", "123456"),
}

//define darpa request template
type DarpaRequestTemplate struct {
	Uri  string
	User string
	Pass string
}

//fill authentication
func (r DarpaRequestTemplate) FillAuthentication(httpRequest *http.Request) {
	if len(r.User) > 0 {
		httpRequest.SetBasicAuth(r.User, r.Pass)
	}
}

//format utl using darpa params
func (r DarpaRequestTemplate) GetURI(params map[string]string) string {

	uri := r.Uri

	var symbol = "?"
	for c, v := range params {
		if len(v) > 0 {
			uri = uri + symbol + c + "=" + v
			symbol = "&"
		}
	}

	log.Warning("formando uma url darpa!", uri)

	return uri
}
