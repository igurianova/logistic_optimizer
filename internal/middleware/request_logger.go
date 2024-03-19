package middleware

import (
	"fmt"
	"net/http"
)

// RequestLoggerMiddleware Метод логирующий каждый запрос
func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Log: ", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
