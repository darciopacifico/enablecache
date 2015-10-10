package config

import (
	"encoding/json"
	_ "expvar" //this anonymous import is used in telemetry exposing!
	"flag"
	"fmt"
	nativeLog "log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/op/go-logging"
	"github.com/rcrowley/go-metrics"
)

//basic structure for statistics service
type Stats struct {
	hit, miss metrics.Counter
}

//a http worker pool request template
type HTTPRoundTrip struct {
	Request  *http.Request
	Response chan (interface{})
}

//worker pool configuration model
type WorkerPoolConfig struct {
	Name             string
	RegexPattern     string
	QueueDepth       int
	Workers          int
	BurstPreventTime int
	Timeout          int
	Regexp           *regexp.Regexp
	InputChanel      chan HTTPRoundTrip
	BurstPrevent     <-chan time.Time
	Status           bool

	MetricsReservoirSize *int
	MetricsAlpha         *float64
	MetricsHistogram     *metrics.Histogram
	MetricsSuccess       *metrics.Meter
	MetricsErrors        *metrics.Meter
}

//default format to biro logging
var (
	format = logging.MustStringFormatter(
		"%{color}%{time:15:04:05.000} PID:%{pid} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}")

	//telemetry host:port
	TelemetryHost string = ""
	//port to expose http/rest services
	Port int
	//string representation of log level, fail at all if parse fail
	strLogLevel string
	//config file path
	logLevel logging.Level
	//config file path
	ConfigFile string

	//enable cpu profile
	CPUProfile string

	//config file path passwords
	ConfigFilePass string

	//config strutct. contains config state inside
	configBiro ConfigStruct

	BiroVersion = "2.8.1"
)

//structure specialized to load and retrieve generic config params
type ConfigStruct struct {
	Conf map[string]interface{}

	//config for worker pool
	WPC []*WorkerPoolConfig
}

//parse all entry flags and configure application settings one time
func parseFlags() {

	if !flag.Parsed() { //assure one and just one parsing
		nativeLog.Println("Setting BIRO for start!")

		//load flags from command line
		loadFlags()

		//start stats count
		//startMetrics()

		//expvar telemetry
		exposeTelemetry(TelemetryHost)

		profile()

		//load external configuration file
		loadConfigFile(ConfigFile, true)
		loadConfigFile(ConfigFilePass, false)

		//tune http listener
		tuneBiro()

		SetLogger()

	}
}

func profile() {

	if CPUProfile != "" {

		fmt.Println("Starting biro in profile mode! Profile output %v", CPUProfile)
		f, err := os.Create(CPUProfile)
		if err != nil {
			fmt.Println("Error trying to write a profiling output! " + err.Error())
		}
		pprof.StartCPUProfile(f)

		fHeap, errHeap := os.Create("heap_" + CPUProfile)
		if errHeap != nil {
			fmt.Println("Error trying to write a profiling output! " + err.Error())
		}
		pprof.WriteHeapProfile(fHeap)

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			for sig := range c {

				// sig is a ^C, handle it
				pprof.StopCPUProfile()

				fmt.Println("Profiling Stoped! ", sig.String())
				fmt.Println("Exiting utilsbiro.T..")
				os.Exit(0)
			}
		}()

		go func() {
			fmt.Println(">>>> Exposing profile at localhost:6060")
			fmt.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

}

//use expvar lib to expose telemetry of application process
func exposeTelemetry(hostPort string) {

	if len(hostPort) > 0 { // expose temeletry only if hostPort was set
		sock, err := net.Listen("tcp", hostPort)
		if err != nil {
			fmt.Println("Error trying to expose telemetry port!", err)
			os.Exit(1)
		}

		go func() {
			fmt.Println("BIRO telemetry on " + hostPort)
			http.Serve(sock, nil)
		}()
	}
}

//get config boolean param
func (c ConfigStruct) ConfigBool(name string, defValue bool) bool {

	strVal := c.Config(name, strconv.FormatBool(defValue))

	boolVal, err := strconv.ParseBool(strVal)

	if err != nil {
		fmt.Println("Error trying to parse parameter to boolean val!", name, "", strVal, err)
		os.Exit(1)
	}
	return boolVal
}

//get config int param
func (c ConfigStruct) ConfigFloat(name string, defValue string) float64 {
	strVal := c.Config(name, defValue)

	floatVal, err := strconv.ParseFloat(strVal, 64)

	if err != nil {
		fmt.Println("Error trying to parse parameter to float64 val!", name, "", strVal, err)
		os.Exit(1)
	}

	return floatVal
}

//get config int param
func (c ConfigStruct) ConfigInt(name string, defValue string) int {
	strVal := c.Config(name, defValue)

	intVal, err := strconv.ParseInt(strVal, 10, 0)

	if err != nil {
		fmt.Println("Error trying to parse parameter to int val!", name, "", strVal, err)
		os.Exit(1)
	}

	return int(intVal)
}

//get config param
func (c ConfigStruct) Config(name string, defValue string) string {
	if val, hasVal := c.Conf[name]; hasVal {

		return fmt.Sprint(val)
	}

	nativeLog.Printf("recuperando valor DEFAULT para propriedade %s, valor default: %s", name, defValue)

	return defValue
}

//parse and set log level. pan ic if parse fail
func GetLevelLog() logging.Level {

	//log level
	var err error
	logLevel, err = logging.LogLevel(strLogLevel)
	if err != nil {
		fmt.Println("Error trying to parse the specified log level!", err)
		os.Exit(1)
	}
	return logLevel
	//logging.SetLevel(logLevel, "biro")
}

//load and parse flags
func loadFlags() {
	//parse all flags

	// = flag.String("cpuprofile", "", "write cpu profile to file")

	flag.StringVar(&ConfigFile, "conf", "/etc/biro/biro.conf", "Inform the config file path.")
	flag.StringVar(&ConfigFilePass, "confPass", "/etc/biro/biroPass.conf", "Inform the config file path for passwords.")
	flag.IntVar(&Port, "p", 8080, "Port for HTTP/REST services exposing.")
	flag.StringVar(&strLogLevel, "l", "DEBUG", "Loglevel: DEBUG | INFO | NOTICE | WARNING | ERROR | CRITICAL")
	flag.StringVar(&TelemetryHost, "t", "", "Host and port to expose telemetry REST. Empty string means no telemetry.")
	flag.StringVar(&CPUProfile, "cpuprofile", "", "Write profile info into a file.")
	showVersion := flag.Bool("version", false, "Show version info and exit.")
	flag.Parse()

	if *showVersion {
		fmt.Println("=====================================================")
		fmt.Println("BIRO - Brazilian Item Retriever Orchestrator v." + BiroVersion)
		fmt.Println("=====================================================")
		fmt.Println("Exiting...")
		os.Exit(0)
	}

}

//tune http listener and runtime
func tuneBiro() {
	//get configs tunnings for http component
	duration := time.Duration(configBiro.ConfigInt("ResponseHeaderTimeout_seconds", "40"))
	maxConns := int(configBiro.ConfigInt("MaxIdleConnsPerHost", "1000"))

	//tunne http component for biro
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = maxConns
	http.DefaultTransport.(*http.Transport).ResponseHeaderTimeout = time.Second * duration

	//set runtime GOMAXPROCS
	numCPUs := strconv.FormatInt(int64(runtime.NumCPU()), 10)
	gomaxprocs := int(configBiro.ConfigInt("GOMAXPROCS", numCPUs))
	runtime.GOMAXPROCS(gomaxprocs)
}

//load external config file
func loadConfigFile(configFile string, required bool) {
	//file config file existence
	_, err := os.Stat(configFile)
	if err != nil {
		if required {
			nativeLog.Println("Arquivo de configuracao "+configFile+" inválido ou inexistente!", err)
			os.Exit(1) //do not continue application loading
		} else {
			nativeLog.Println("Arquivo de configuracao opcional " + configFile + " inválido ou inexistente!")
			return
		}
	}

	file, err := os.Open(configFile)
	defer file.Close()
	if err != nil {
		fmt.Println("Error trying to open config file!", err)
		os.Exit(1) //do not continue application loading
	}

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&configBiro.Conf)
	if err != nil {
		fmt.Println(err)
		fmt.Println(err.Error(), "Erro ao tentat abrir arquivo de configuracao '"+configFile+"'")
		os.Exit(1)
	}

	//DECODE CONFIGURATION FOR WORKER POOL

	type WorkerConfig struct {
		ExternalResourcesTunning []*WorkerPoolConfig `json:"externalResourcesTunning"`
	}

	wc := &WorkerConfig{}

	file2, err2 := os.Open(configFile)
	defer file2.Close()
	if err2 != nil {
		fmt.Println("Error trying to open config file!", err)
		os.Exit(1) //do not continue application loading
	}
	decoder2 := json.NewDecoder(file2)

	err = decoder2.Decode(&wc)
	if err != nil {
		fmt.Println(err)
		fmt.Println(err.Error(), "Erro ao tentat abrir arquivo de configuracao '"+configFile+"'")
		os.Exit(1)
	}

	configBiro.WPC = wc.ExternalResourcesTunning

}

//parse flags and return config object
func CreateConfig() ConfigStruct {
	parseFlags()

	return configBiro
}

//config a logger to specified module
func SetLogger() {
	//module := "biro"

	backend1 := logging.NewLogBackend(os.Stdout, "", 0)
	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	logging.SetBackend(backend1Formatter)

	logging.SetLevel(GetLevelLog(), "biro")

}

//create statistics object usgin metrics
func NewStats(name string) *Stats {
	stats := Stats{metrics.NewCounter(), metrics.NewCounter()}
	metrics.Register(name+"-hit", stats.hit)
	metrics.Register(name+"-miss", stats.miss)
	return &stats
}

//increment cache hit
func (st *Stats) Hit() {
	st.hit.Inc(1)
}

//increment cache Miss
func (st *Stats) Miss() {
	st.miss.Inc(1)
}

//initialize metrics "daemon"
func startMetrics() {
	go metrics.Log(metrics.DefaultRegistry, time.Second*60, nativeLog.New(os.Stderr, "metrics: ", nativeLog.Lmicroseconds))
}
