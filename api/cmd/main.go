package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	rootHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_root_hits_total",
	})
	ticks = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_ticks_total",
	})
)

func handler(w http.ResponseWriter, r *http.Request) {
	rootHits.Inc()
	slog.Info("request", "service", "api", "path", r.URL.Path)
	_, _ = fmt.Fprintln(w, "api:"+r.URL.Path)
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	go func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for range t.C {
			ticks.Inc()
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	mux.Handle("/metrics", promhttp.Handler())

	slog.Info("starting", "service", "api", "addr", ":8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		slog.Error("server failed", "err", err)
	}
}
