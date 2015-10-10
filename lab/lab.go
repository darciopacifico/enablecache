package main

import (
	"gitlab.wmxp.com.br/bis/biro/legacy"
)

func main() {

	params := map[string]string{"id": "1111", "zuba": "2222"}

	uriLegacy := legacy.RequestProduct.GetURI(params)

	println("URI formada =", uriLegacy)

}
