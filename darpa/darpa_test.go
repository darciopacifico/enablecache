package darpa

import (
	"gitlab.wmxp.com.br/bis/biro/rest"
)

var (
	darpaStub = DarpaStub{
		Rest: rest.GenericRESTClient{
			HttpCaller: rest.ExecuteRequestHot,
			//HttpCaller: pool.ExecuteRequestPool, // nao usar ate estabilizar
		},
	}
)
