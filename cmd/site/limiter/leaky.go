package limiter

import (
	"time"
	"sync"
	"fmt"
)

type Token struct{}
type Bucket struct {
	tokens    chan Token
	lastUsed  time.Time
	maxTokens int
}

func NewBucket(maxTokens int) *Bucket {
	ch := make(chan Token, maxTokens)
	b := &Bucket{
		lastUsed:  time.Now(),
		tokens:    ch,
		maxTokens: maxTokens,
	}
	return b
}

func (b *Bucket) Refill(count int) {
	// TODO
}

func (b *Bucket) TryGet(timeout time.Duration) chan bool {
	b.lastUsed = time.Now()
	out := make(chan bool)
	// TODO
	return out
}

type ClientID string

type LeakyBucketLimiterOption func(limiter *LeakyBucketLimiter)

func WithCustomTimeAfter(timeAfter TimeAfter) LeakyBucketLimiterOption {
	return func(limiter *LeakyBucketLimiter) {
		limiter.timeAfter = timeAfter
	}
}

func NewLeakyBucketLimiter(totalTokens int, maxTokensPerClient int, refillPeriod time.Duration, maxInactiveClientTime time.Duration, ops ...LeakyBucketLimiterOption) *LeakyBucketLimiter {
	limiter := &LeakyBucketLimiter{
		clients:               make(map[ClientID]*Bucket),
		refillPeriod:            refillPeriod,
		totalTokens:           totalTokens,
		maxTokensPerClient:    maxTokensPerClient,
		maxInactiveClientTime: maxInactiveClientTime,
		timeAfter:             time.After,
	}
	for _, op := range ops {
		op(limiter)
	}
	return limiter
}

type LeakyBucketLimiter struct {
	mtx                   sync.Mutex
	clients               map[ClientID]*Bucket
	refillPeriod            time.Duration
	maxInactiveClientTime time.Duration
	totalTokens           int
	maxTokensPerClient    int
	timeAfter             TimeAfter
}

type TimeAfter func(duration time.Duration) <-chan time.Time

func (limiter *LeakyBucketLimiter) Start() {
	for {
		limiter.removeInactiveClients()
		limiter.distributeTokens()
		_, ok := <-limiter.timeAfter(limiter.refillPeriod)

		if !ok {
			fmt.Println("closed")
			return
		}
	}
}

func (limiter *LeakyBucketLimiter) removeInactiveClients() {
	// TODO
}

func (limiter *LeakyBucketLimiter) distributeTokens() {
	limiter.mtx.Lock()
	defer limiter.mtx.Unlock()

	if len(limiter.clients) == 0 {
		fmt.Println("No clients to distribute tokens")
		return
	}

	inUse := 0
	for _, v := range limiter.clients {
		inUse += len(v.tokens)
	}

	refill := limiter.totalTokens - inUse
	perClient := refill / len(limiter.clients)

	fmt.Printf("[%v] Distribute %d tokens for %d clients\n", time.Now(), perClient, len(limiter.clients))

	if perClient > 0 {
		for _, v := range limiter.clients {
			v.Refill(perClient)
		}
	} else if perClient == 0 && refill > 0 {
		for _, v := range limiter.clients {
			v.Refill(1)
			refill--
			if refill <= 0 {
				return
			}
		}
	}

}

func (limiter *LeakyBucketLimiter) GetToken(cli ClientID, timeout time.Duration) chan bool {
	limiter.mtx.Lock()
	defer limiter.mtx.Unlock()
	var bucket *Bucket
	bucket, ex := limiter.clients[cli]
	if !ex {
		bucket = NewBucket(limiter.maxTokensPerClient)
		limiter.clients[cli] = bucket
	}

	return bucket.TryGet(timeout)
}