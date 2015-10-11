package cache

import (
	nativeLog "log"
	"os"

	"github.com/rcrowley/go-metrics"
	"time"
)

//basic structure for statistics service
type Stats struct {
	hit, miss metrics.Counter
}

//create new stats object for cache statistics
func NewStats(name string) *Stats {

	stats := Stats{
		metrics.NewCounter(),
		metrics.NewCounter(),
	}

	metrics.Register(name+"-hit", stats.hit)
	metrics.Register(name+"-miss", stats.miss)

	return &stats
}

//increment cache hit
func (st *Stats) Hit() {
	//update histogram
	st.hit.Inc(1)

}

//increment cache Miss
func (st *Stats) Miss() {
	//update histogram
	st.miss.Inc(1)
}

//initialize metrics "daemon"
func init() {
	go metrics.Log(metrics.DefaultRegistry, time.Minute*2, nativeLog.New(os.Stderr, "metrics: ", nativeLog.Lmicroseconds))
}
