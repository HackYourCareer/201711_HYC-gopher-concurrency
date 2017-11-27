package middleware

import (
	"time"
	"net/http"

	"github.com/SAPHybrisGliwice/gopher-concurrency/cmd/site/limiter"
)

func NewLeakyBucketLimiterMiddleware(l *limiter.LeakyBucketLimiter, timeout time.Duration) *LeakyBucketLimiterMiddleware {
	return &LeakyBucketLimiterMiddleware{
		timeout: timeout,
		l:       l,
	}
}

type LeakyBucketLimiterMiddleware struct {
	l       *limiter.LeakyBucketLimiter
	timeout time.Duration
}

func (m *LeakyBucketLimiterMiddleware) Limit(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		client := limiter.ClientID(req.Header.Get("client_id"))
		canContinue := <- m.l.GetToken(client, m.timeout)
		if !canContinue {
			rw.WriteHeader(http.StatusTooManyRequests)
			return
		}
		next(rw, req)
	}
}
