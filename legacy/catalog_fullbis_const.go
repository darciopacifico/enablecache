package legacy

import (
	"errors"
	"fmt"
	"gitlab.wmxp.com.br/bis/biro/rest"
	"gitlab.wmxp.com.br/bis/biro/utilsbiro"
	"net/http"
)

type FullBisRequestTemplate struct {
	rest.RequestTemplate
}

//compose uri for a rest service call, based on RequestPayload
//simple uri composition. there is no other dynamic displays
func (r FullBisRequestTemplate) GetURI(params map[string]string) string {

	log.Debug("URI template %v", r.Uri)
	log.Debug("Params %v", params)

	id, hasId := params["id"]

	if !hasId {
		panic(errors.New("id item map not informed!!"))
	}

	arrId := []string{params[id]}

	interfaceParam := utilsbiro.ToInterfaceArr(arrId)

	uri := fmt.Sprintf(r.Uri, interfaceParam...)

	return uri
}

func (r FullBisRequestTemplate) FillAuthentication(httpRequest *http.Request) {
	// no authentication on bis repo
}

//template for offer request in bis repo
var RequestItemOfferV1 = FullBisRequestTemplate{
	rest.RequestTemplate{
		Uri:  conf.Config("URIFullBisItemOfferV1", "http://localhost:9090/item/offer/%[1]v"),
		User: conf.Config("userBisItemOfferV1", ""),
		Pass: conf.Config("passBisItemOfferV1", ""),
	},
}

//template for offers by products request in bis repo
var RequestItemV1 = FullBisRequestTemplate{
	rest.RequestTemplate{
		Uri:  conf.Config("URIFullBisItemV1", "http://localhost:9090/item/%[1]v"),
		User: conf.Config("userBisItemV1", ""),
		Pass: conf.Config("passBisItemV1", ""),
	},
}
