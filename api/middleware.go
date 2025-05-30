package api

import (
	"fmt"
	"net/http"
	"time"
)

func CacheFor(d time.Duration) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			now := time.Now()

			w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", d/time.Second))
			w.Header().Set("Expires", now.Add(d).Format(http.TimeFormat))
			w.Header().Set("Last-Modified", now.UTC().Format(http.TimeFormat))

			h.ServeHTTP(w, r)
		})
	}
}
