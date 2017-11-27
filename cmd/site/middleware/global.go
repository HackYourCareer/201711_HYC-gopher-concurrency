package middleware

import "net/http"

func NewGlobalLimiter(maxConcurrentUsers int) *GlobalLimiter {
	return &GlobalLimiter{
		inProgress: make(chan struct{}, maxConcurrentUsers),
	}
}

type GlobalLimiter struct {
	inProgress chan struct{}
}

func (l *GlobalLimiter) Limit(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		l.inProgress <- struct{}{}
		next(rw, req)
		<-l.inProgress

	}
}

