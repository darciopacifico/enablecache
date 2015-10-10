package main

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"strconv"

	"github.com/op/go-logging"
	"gitlab.wmxp.com.br/bis/biro/cache"
	"gitlab.wmxp.com.br/bis/biro/config"
	"gitlab.wmxp.com.br/bis/biro/schema"

	_ "net/http/pprof"
)

// define log var
var conf = config.CreateConfig()

//var log = logging.MustGetLogger("biro")

func init() {
	// Initialize Serializable structs
	registerStructs()
}

//initiate BIRO applicationbiro
func main() {

	//just show biro configs
	printBiroSplash()

	//create a new gorilla rest router
	router := NewRouter()
	log := logging.MustGetLogger("biro")
	//expose biro mux router and stops here, until application stops
	log.Info("", http.ListenAndServe(":"+strconv.Itoa(config.Port), router))

}

//just show biro configs
func printBiroSplash() {
	fmt.Println("===============BIRO Ready!===============")
	fmt.Println("Version: v." + config.BiroVersion)
	fmt.Println("Log Level:", logging.GetLevel("biro").String())
	fmt.Println("Config File:", config.ConfigFile)
	fmt.Println("Config File Passwords:", config.ConfigFilePass)
	fmt.Println("HTTP/REST Ready on:", config.Port)
	if len(config.TelemetryHost) > 0 {
		fmt.Println("HTTP/REST Process Telemetry:", config.TelemetryHost+"/debug/vars")
	} else {
		fmt.Println("HTTP/REST Process Telemetry: disabled", config.TelemetryHost)
	}
	fmt.Println("=========================================")
}

// Registers all objetcs that will eventually end up in Redis
func registerStructs() {

	gob.Register(schema.ItemV1{})
	gob.Register(schema.ItemOfferV1{})

	gob.Register(schema.ProductLegacy{})
	gob.Register(schema.SkuLegacy{})
	gob.Register(schema.OfferLegacy{})
	gob.Register(schema.ItemOfferV1{})
	gob.Register(cache.CacheRegistry{})
	//gob.Register(cache.DefineTTLGeneric{})
}
