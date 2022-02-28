package profiler

import (
	log "ledfx/logger"
	"net/http"
	_ "net/http/pprof" //nolint:gosec
)

func Start() {
	go func() {
		log.Logger.WithField("category", "Performance Profiler").Errorf("Error starting PPROF: %v", http.ListenAndServe("localhost:6060", nil))
	}()

}
