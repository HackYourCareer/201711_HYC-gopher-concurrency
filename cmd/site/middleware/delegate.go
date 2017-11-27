package middleware

import "net/http"

type LimiterMiddleware interface {
	Limit(next http.HandlerFunc) http.HandlerFunc
}

func NewDelegatingMiddleware(concrete LimiterMiddleware) *DelegatingMiddleware {
	return &DelegatingMiddleware{concrete: concrete}
}

type DelegatingMiddleware struct {
	concrete LimiterMiddleware
}

func (m *DelegatingMiddleware) Limit(next http.HandlerFunc) http.HandlerFunc {
	if m.concrete == nil {
		return next
	}
	return m.concrete.Limit(next)
}
