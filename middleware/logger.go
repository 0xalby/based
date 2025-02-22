package middleware

import (
	"net/http"

	"github.com/charmbracelet/log"
)

// Buffered logging channel
var logChan = make(chan string, 100)

// Middleware asynchronously logging every request
func Logger(logger log.Logger) func(next http.Handler) http.Handler {
	go func() {
		for msg := range logChan {
			logger.Info(msg)
		}
	}()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logChan <- "recieved " + r.Method + " on " + r.URL.Path + " from " + r.RemoteAddr
			next.ServeHTTP(w, r)
		})
	}
}
