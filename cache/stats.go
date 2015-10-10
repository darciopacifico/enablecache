package cache

import (
	nativeLog "log"
	"os"

	"github.com/rcrowley/go-metrics"
)

//basic structure for statistics service
type Stats struct {
	hit, miss metrics.Counter
}

//create new stats object for cache statistics
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
func StartMetrics() {
	go metrics.Log(metrics.DefaultRegistry, 30e9, nativeLog.New(os.Stderr, "metrics: ", nativeLog.Lmicroseconds))
}
