package middleware

import (
	"net/http"
	"time"

	"github.com/charmbracelet/log"
)

// customWriter is a wrapper around http.ResponseWriter to capture status
type customWriter struct {
	http.ResponseWriter
	statusCode int
}

// Custom WriteHeader function
func (ww *customWriter) WriteHeader(code int) {
	ww.statusCode = code
	ww.ResponseWriter.WriteHeader(code)
}

// Logger middleware that logs requests and responses
func Logger(logger log.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			logger.Infof("received %s on %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
			// Wrap the response writer to capture the response
			ww := &customWriter{ResponseWriter: w}
			next.ServeHTTP(ww, r)
			duration := time.Since(start)
			logger.Infof("responded with %d in %s", ww.statusCode, duration)
		})
	}
}
