package middleware

import "net/http"

func NewGlobalLimiter(maxConcurrentUsers int) *GlobalLimiter {
	return &GlobalLimiter{}
}

type GlobalLimiter struct {
}

func (l *GlobalLimiter) Limit(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		next(rw, req)
	}
}

