package api

import (
	"log"
	"net/http"
	"time"
)

func Logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		uuid := time.Now().Nanosecond()
		log.Printf(
			"[%d] Started %s %s %s",
			uuid,
			r.Method,
			r.RequestURI,
			name,
		)

		inner.ServeHTTP(w, r)

		log.Printf(
			"[%d] Completed %s %s %s %s",
			uuid,
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)

		log.Println()

	})
}
