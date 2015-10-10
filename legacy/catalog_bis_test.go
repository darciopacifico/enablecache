package legacy

import (
	//	"testing"

	"gitlab.wmxp.com.br/bis/biro/rest"
)

var catalogBis = CatalogBIS{
	rest.GenericRESTClient{
		rest.ExecuteRequestHot,
	},
}
