package legacy

import (
	"errors"
	"fmt"
	"gitlab.wmxp.com.br/bis/biro/rest"
	"gitlab.wmxp.com.br/bis/biro/utilsbiro"
	"net/http"
)

type BISRequestTemplate struct {
	rest.RequestTemplate
}

//compose uri for a rest service call, based on RequestPayload
//simple uri composition. there is no other dynamic displays
func (r BISRequestTemplate) GetURI(params map[string]string) string {

	log.Debug("URI template %v", r.Uri)
	log.Debug("Params %v", params)

	id, hasId := params["id"]

	if !hasId {
		panic(errors.New("id item map not informed!!"))
	}

	strId := []string{params[id]}

	interfaceParam := utilsbiro.ToInterfaceArr(strId)

	uri := fmt.Sprintf(r.Uri, interfaceParam...)

	return uri
}

/*func (r BiroRequestTemplate) GetURI(params ...interface{}) string {
 */

func (r BISRequestTemplate) FillAuthentication(httpRequest *http.Request) {

}

/*func (r BiroRequestTemplate) FillAuthentication(httpRequest *http.Request) {
func (r BiroRequestTemplate) GetURI(params ...interface{}) string {
*/

//template for offer request in bis repo
var requestOfferBis = BISRequestTemplate{
	rest.RequestTemplate{
		Uri:  conf.Config("URIBisOffer", "http://10.134.114.95:8080/offer/%[1]s"),
		User: conf.Config("userBisOfferRepo", ""),
		Pass: conf.Config("passBisOfferRepo", ""),
	},
}

//template for offers by products request in bis repo
var requestOffersByProdBis = BISRequestTemplate{
	rest.RequestTemplate{
		Uri:  conf.Config("URIBisOffersByProd", "http://10.134.114.95:8080/offer/product/%[1]s"),
		User: conf.Config("userBisOfferByProdRepo", ""),
		Pass: conf.Config("passBisOfferByProdRepo", ""),
	},
}
