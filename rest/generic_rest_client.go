package rest

import (
	"net/http"

	"fmt"
	"github.com/op/go-logging"
	"gitlab.wmxp.com.br/bis/biro/schema"
	"io"
)

var (
	log = logging.MustGetLogger("biro")
)

type IRequestTemplate interface {
	FillAuthentication(httpRequest *http.Request)
	GetURI(params map[string]string) string
}

type RequestTemplate struct {
	Uri  string
	User string
	Pass string
}

//compose uri for a rest service call, based on RequestPayload
func (r RequestTemplate) FillAuthentication(httpRequest *http.Request) {
	if len(r.User) > 0 {
		httpRequest.SetBasicAuth(r.User, r.Pass)
	}
}

//compose uri for a rest service call, based on RequestPayload
//simple uri composition. there is no other dynamic displays
//can be overrided to some specific way of uri composition, like to include Displays of catalog legacy
func (r RequestTemplate) GetURI(params map[string]string) string {
	log.Debug("chamando o uri q ninguem chama!!")
	return r.Uri
}

//contract do generic buildMyObj function
//after the generic call of a rest service, some function implementation with this signature will be responsible to parse rest result to some specific structure
type BuildMyObjFunc func(resp *http.Response) interface{}

// maxima interface for rest calls
type RESTClient interface {
	ExecuteRESTGet(request IRequestTemplate, buildMyObjFunc BuildMyObjFunc, params map[string]string) (*interface{}, bool, error)
	ExecuteRESTPost(request IRequestTemplate, buildMyObj BuildMyObjFunc, body io.Reader, params map[string]string) (*interface{}, bool, error)
}

// execution hot of some http request.
// to mock some http request, use legacy/test_utils_test.go:11, ExecuteRequestMock
func ExecuteRequestHot(client *http.Client, request *http.Request) (*http.Response, error) {
	resp, err := client.Do(request)
	return resp, err
}

// define the signature to a http caller
type HttpCaller func(client *http.Client, request *http.Request) (*http.Response, error)

//implements a restclient
type GenericRESTClient struct {
	HttpCaller // can be mocked to test porpouses
}

//Execute a generic REST/JSON get and call BuildMyObj to build response object
func (r GenericRESTClient) ExecuteRESTGet(request IRequestTemplate, buildMyObj BuildMyObjFunc, params map[string]string) (*interface{}, bool, error) {
	return r.ExecuteREST("GET", request, buildMyObj, nil, params)
}

//Execute a generic REST/JSON get and call BuildMyObj to build response object
func (r GenericRESTClient) ExecuteRESTPost(request IRequestTemplate, buildMyObj BuildMyObjFunc, body io.Reader, params map[string]string) (*interface{}, bool, error) {
	return r.ExecuteREST("POST", request, buildMyObj, body, params)
}

//Execute a generic REST/JSON get and call BuildMyObj to build response object
func (r GenericRESTClient) ExecuteREST(method string, request IRequestTemplate, buildMyObj BuildMyObjFunc, body io.Reader, params map[string]string) (*interface{}, bool, error) {
	requestURI := request.GetURI(params)
	log.Debug("URL montada para consulta: %v", requestURI)

	reqREST, err := http.NewRequest(method, requestURI, body)
	reqREST.Header.Set("Content-Type", "application/json")

	if err != nil {
		log.Error("Erro ao tentar criar objeto de request %v", err)
		return nil, false, err
	}

	request.FillAuthentication(reqREST)

	client := &http.Client{}
	resp, err := r.HttpCaller(client, reqREST)
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if err != nil {

		log.Error("erro ao tentar invocar rest service %v", err.Error())

		if resp != nil {
			log.Error("HTTP STATUS CODE:", resp.StatusCode)
			log.Error("HTTP STATUS:", resp.Status)
		}

		biroErr := schema.BIROError{
			Message: fmt.Sprintf("Error trying to access external REST service!"),
			Parent:  err,
			Code:    500,
		}
		return nil, false, biroErr
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Debug("%d Registro nao encontrado! %s", resp.StatusCode, resp.Status)
		return nil, false, nil

	} else if resp.StatusCode != http.StatusOK {
		log.Error("Erro ao tentar consultar registro! ", resp.Status)
		return nil, false, schema.BIROError{
			Message: fmt.Sprintf("Erro ao tentar consultar registro! Status Code %v!", resp.StatusCode),
			Parent:  err,
			Code:    resp.StatusCode,
		}

	}

	destiny := buildMyObj(resp)

	return &destiny, true, nil
}
