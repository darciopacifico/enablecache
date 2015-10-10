package pool

// In this example we'll look at how to implement
// a _worker pool_ using goroutines and channels.

import (
	"errors"
	"fmt"
	"github.com/op/go-logging"
	"github.com/rcrowley/go-metrics"
	"gitlab.wmxp.com.br/bis/biro/config"
	"gitlab.wmxp.com.br/bis/biro/rest"
	"gitlab.wmxp.com.br/bis/biro/schema"
	"net/http"
	"os"
	"regexp"
	"time"
)

var (
	//default log
	log = logging.MustGetLogger("biro")

	rounTripTOAdvance = conf.ConfigInt("roundTripTimeoutAdvance", "10") //round trip timeout must happens before external

	conf = config.CreateConfig()

	// channels for this.
	workerPools = []*config.WorkerPoolConfig{}

	//default worker attributes
	name    = "Internal Default Worker Pool"
	pattern = ".*"

	//for metrics histogram
	reservoirSize = 2048
	alpha         = 0.015

	//default worker
	defaultWorker = &config.WorkerPoolConfig{
		Name:             name,
		RegexPattern:     pattern,
		QueueDepth:       1,
		Workers:          100,
		BurstPreventTime: 10,
		Timeout:          50000,
		Regexp:           nil,
		MetricsHistogram: GetHistogram(name, pattern, reservoirSize, alpha),
		MetricsErrors:    GetMeter("ERROR", name, pattern),
		MetricsSuccess:   GetMeter("SUCCESS", name, pattern),
		InputChanel:      nil,
	}
)

//creates a new histogram object and register on metrics
func GetHistogram(name string, pattern string, reservoirSize int, alpha float64) *metrics.Histogram {
	s := metrics.NewExpDecaySample(reservoirSize, alpha)
	h := metrics.NewHistogram(s)

	fullName := fmt.Sprintf("'%v', URI Pattern='%v'", name, pattern)

	metrics.Register(fullName, h)

	return &h
}

//creates a new histogram object and register on metrics
func GetMeter(meterType string, name string, pattern string) *metrics.Meter {
	m := metrics.NewMeter()
	fullName := fmt.Sprintf("Meter %v '%v', URI Pattern='%v'", meterType, name, pattern)
	metrics.Register(fullName, m)
	m.Rate1()
	return &m
}

//overload fuction for overloadless language :-(
//creates a new histogram object and register on metrics
func GetHistogramFromWorker(w config.WorkerPoolConfig) (*metrics.Histogram, *metrics.Meter, *metrics.Meter) {

	var MetricsReservoirSize int = 1024
	var MetricsAlpha float64 = 0.015

	if w.MetricsReservoirSize != nil && w.MetricsAlpha != nil {
		log.Debug("Generation metrics histogram for %v, pattern %v", w.Name, w.RegexPattern)
		MetricsReservoirSize = *w.MetricsReservoirSize
		MetricsAlpha = *w.MetricsAlpha

	} else {
		log.Warning("There is no metrics parameters registered for %v, pattern %v", w.Name, w.RegexPattern)
		log.Warning("Using default MetricsReservoirSize=%v and MetricsAlpha%v", MetricsReservoirSize, MetricsAlpha)
		log.Warning("Please define MetricsReservoirSize and MetricsAlpha attributes for %v, pattern %v", w.Name, w.RegexPattern)
	}

	return GetHistogram(w.Name, w.RegexPattern, MetricsReservoirSize, MetricsAlpha), GetMeter("SUCCESS", w.Name, w.RegexPattern), GetMeter("ERROR", w.Name, w.RegexPattern)

}

func _init() {

	defer func() { //assure for not panicking
		if r := recover(); r != nil {
			log.Error("Error trying to start the http worker pool! %v", r)
			panic(r)
		}
	}()

	// this configs mut come from a config file
	workerPools = conf.WPC

	//
	workerPools = append(workerPools, defaultWorker)

	for _, worker := range workerPools {

		regexp, err := regexp.Compile(worker.RegexPattern)

		if err != nil {
			log.Error("Error trying to spawn a worker '%v' for regex %v ", worker.Name, worker.RegexPattern, err)
			worker.Status = false
			break
		}

		worker.InputChanel = make(chan config.HTTPRoundTrip)
		//worker.InputChanel = make(chan config.HTTPRoundTrip, 1)
		worker.BurstPrevent = time.Tick(time.Millisecond * time.Duration(worker.BurstPreventTime))

		worker.MetricsHistogram, worker.MetricsSuccess, worker.MetricsErrors = GetHistogramFromWorker(*worker)
		worker.Regexp = regexp

		log.Debug("Instantiating %v workers, named '%v-N' for URL pattern '%v'",
			worker.Workers, worker.Name, worker.RegexPattern)

		for idWorker := 0; idWorker < worker.Workers; idWorker++ {
			go spawnWorker(worker, idWorker)
		}
	}
}

// execution hot of some http request.
// to mock some http request, use legacy/test_utils_test.go:11, ExecuteRequestMock
func ExecuteRequestPool_desativado(client *http.Client, req *http.Request) (*http.Response, error) {

	var selectedWP = defaultWorker

	for _, wp := range workerPools {
		if isFitWorker(req, wp) {
			selectedWP = wp
			break
		}
	}

	<-selectedWP.BurstPrevent // wait for bust prevent tick
	return execute(req, selectedWP)
}

//spawn a worker for a  worker pool config
func spawnWorker(worker *config.WorkerPoolConfig, workerId int) {

	client := &http.Client{} // safe for concurrent use
	//client.Timeout = time.Duration(time.Millisecond * time.Duration(worker.Timeout+200))

	if worker.InputChanel == nil {
		log.Error("Worker %v has a nil inputChanel! Cant start workers for pool request!", worker.Name)
		os.Exit(0)
	}

	for rt := range worker.InputChanel {
		log.Debug("Worker '%v-%v (of %v)' take %v %v%v", worker.Name, workerId, worker.Workers, rt.Request.Method, rt.Request.URL.Host, rt.Request.URL.Path)

		doRoundTrip(&rt, client, worker, workerId)

	}
}

func doRoundTrip(rtp *config.HTTPRoundTrip, client *http.Client, workers *config.WorkerPoolConfig, workerId int) {
	rt := *rtp
	req := rtp.Request

	log.Debug("Worker '%v-%v (of %v)' take %v %v%v", workers.Name, workerId, workers.Workers, rt.Request.Method, rt.Request.URL.Host, rt.Request.URL.Path)

	//protection against panic events. Last man standing!
	defer func() { //assure for not panicking
		if r := recover(); r != nil {
			log.Error("Recovering! Error in Worker '%v-%v (of %v)' task %v %v%v, error %v ", workers.Name, workerId, workers.Workers, rt.Request.Method, rt.Request.URL.Host, rt.Request.URL.Path, r)

			rt.Response <- schema.BIROError{
				Message: fmt.Sprintf("Recovering! Error in Worker '%v-%v (of %v)' task %v %v%v, error %v ", workers.Name, workerId, workers.Workers, rt.Request.Method, rt.Request.URL.Host, rt.Request.URL.Path, r),
				Parent:  nil,
				Code:    500,
			}
		}
	}()

	chResp := make(chan interface{})

	//call external resource by rest
	go func() {
		response, err := rest.ExecuteRequestHot(client, rt.Request)
		if err != nil {
			log.Error("Error %v", err.Error())
			log.Warning("Request for %v/%v doesn't succeed! Error %v", rt.Request.URL.Host, rt.Request.URL.Path, err)
			chResp <- err
		} else {
			log.Debug("Request for %v/%v succeed, status %v!", rt.Request.URL.Host, rt.Request.URL.Path, response.StatusCode)
			chResp <- response
		}
	}()

	timeoutMax := time.Millisecond * time.Duration(workers.Timeout+workers.BurstPreventTime-rounTripTOAdvance)

	var resp interface{}
	select {
	case resp = <-chResp:
	case <-time.After(timeoutMax):
		log.Error("[RT] Worker '%v' timeout (%v ms) exceed for request %v%v (... + supressed params)", workers.Name, workers.Timeout, req.URL.Host, req.URL.Path)
		resp = schema.BIROError{
			Message: fmt.Sprintf("External resource timed out (expires > %v) ! Host %v%v(... + supressed params)", timeoutMax, req.URL.Host, req.URL.Path),
			Parent:  nil,
			Code:    408,
		}
	}

	rt.Response <- resp
}

//test request url with worker regex
func isFitWorker(req *http.Request, worker *config.WorkerPoolConfig) bool {

	if worker.Regexp == nil {
		log.Error("A regex for worker '%v', pattern %v is null! ", worker.Name, worker.RegexPattern)
		return false
	}

	return worker.Regexp.Match([]byte(req.URL.String()))
}

//Execute the request
func execute(req *http.Request, workers *config.WorkerPoolConfig) (*http.Response, error) {
	roundTrip := config.HTTPRoundTrip{
		//Request  *http.Request
		//Response chan (interface{})
		Request:  req,
		Response: make(chan interface{}, 1),
	}

	startRequest := time.Now()
	workers.InputChanel <- roundTrip

	timeoutMax := time.Millisecond * time.Duration(workers.Timeout+workers.BurstPreventTime)

	var resp interface{}

	select {
	case resp = <-roundTrip.Response:
	case <-time.After(timeoutMax):
		//close(roundTrip.Response)
		log.Error("Worker '%v' timeout (%v ms) exceed for request %v%v (... + supressed params)", workers.Name, workers.Timeout, req.URL.Host, req.URL.Path)
		resp = schema.BIROError{
			Message: fmt.Sprintf("External resource timed out (expires > %v) ! Host %v%v(... + supressed params)", timeoutMax, req.URL.Host, req.URL.Path),
			Parent:  nil,
			Code:    408,
		}
	}

	switch response := resp.(type) {
	case *http.Response:
		duration := time.Since(startRequest)

		if response.StatusCode == 200 {
			markSuccess(duration, workers)
		} else {
			markFail(workers)
		}

		return response, nil

	case error:
		markFail(workers)
		return nil, response

	default:
		markFail(workers)
		log.Error("Error trying to receive response! Response value %v", response)
		return nil, errors.New("Non itentified return!")
	}

}

func markSuccess(duration time.Duration, workers *config.WorkerPoolConfig) {
	if workers == nil {
		return
	}

	refWorker := *workers

	if workers.MetricsSuccess != nil {
		mSuccess := *workers.MetricsSuccess
		mSuccess.Mark(1)
	}

	if refWorker.MetricsHistogram != nil {
		m := *refWorker.MetricsHistogram
		m.Update(int64(duration / time.Millisecond))
	} else {
		log.Warning("There is no metrics histogram for %v", refWorker.Name)
	}
}

func markFail(workers *config.WorkerPoolConfig) {
	if workers == nil {
		return
	}

	if workers.MetricsErrors != nil {
		mError := *workers.MetricsErrors
		mError.Mark(1)
	}
}
