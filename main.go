package main

import (
	"flag"
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	_ "net/http/pprof" // Профилирование
)

func main() {
	var port int
	var gcPercent int

	flag.IntVar(&port, "port", 8080, "HTTP port")
	flag.IntVar(&gcPercent, "gc-percent", 100, "GC percent (debug.SetGCPercent)")
	flag.Parse()

	debug.SetGCPercent(gcPercent)
	fmt.Printf("GC Percent set to %d\n", gcPercent)

	memAlloc := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "go_mem_alloc_bytes",
			Help: "Current memory allocated (bytes)",
		},
		func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc)
		},
	)

	memSys := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "go_mem_sys_bytes",
			Help: "Total memory obtained from OS (bytes)",
		},
		func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Sys)
		},
	)

	totalAlloc := prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "go_mem_total_alloc_bytes",
			Help: "Total memory allocated (bytes)",
		},
		func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.TotalAlloc)
		},
	)

	mallocs := prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "go_mem_mallocs_total",
			Help: "Total number of mallocs",
		},
		func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Mallocs)
		},
	)

	frees := prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "go_mem_frees_total",
			Help: "Total number of frees",
		},
		func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Frees)
		},
	)

	gcCycles := prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "go_gc_cycles_total",
			Help: "Total number of completed GC cycles",
		},
		func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.NumGC)
		},
	)

	lastGCPause := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "go_gc_last_pause_seconds",
			Help: "Duration of last GC pause in seconds",
		},
		func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			if m.NumGC == 0 {
				return 0
			}
			return float64(m.PauseNs[(m.NumGC+255)%256]) / 1e9
		},
	)

	totalGCPause := prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "go_gc_pause_total_seconds",
			Help: "Total GC pause duration (seconds)",
		},
		func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			var total uint64
			for i := uint32(0); i < m.NumGC && i < 256; i++ {
				total += m.PauseNs[i]
			}
			return float64(total) / 1e9
		},
	)

	prometheus.MustRegister(memAlloc, memSys, totalAlloc, mallocs, frees, gcCycles, lastGCPause, totalGCPause)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "GC and Memory Monitoring Server\n")
		fmt.Fprintf(w, "Visit /metrics for Prometheus metrics\n")
		fmt.Fprintf(w, "Visit /debug/pprof/ for profiling\n")
	})

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Starting server at %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
