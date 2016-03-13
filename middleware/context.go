package middleware

import (
	"net/http"

	"github.com/gorilla/context"
)

func WithValue(key, val interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			context.Set(r, key, val)
			next.ServeHTTP(w, r)
		})
	}
}
