package limiter_test

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"fmt"
	"runtime"

	"github.com/SAPHybrisGliwice/gopher-concurrency/cmd/site/limiter"
)

func fixTimeout() time.Duration {
	return time.Millisecond
}

func TestBucketRefillTokensWhenOfferedMoreThanCapacity(t *testing.T) {
	b := limiter.NewBucket(3)
	b.Refill(100)

	for i := 0; i < 3; i++ {
		token := <-b.TryGet(fixTimeout())
		assert.True(t, token)
	}

	token := <-b.TryGet(fixTimeout())
	assert.False(t, token)

}

func TestBucketRefillTokensWhenOfferedLessThanCapacity(t *testing.T) {
	b := limiter.NewBucket(3)
	b.Refill(2)

	for i := 0; i < 2; i++ {
		token := <-b.TryGet(fixTimeout())
		assert.True(t, token)
	}

	token := <-b.TryGet(fixTimeout())
	assert.False(t, token)

}

func TestBucketRefillWhenNoCapacity(t *testing.T) {
	b := limiter.NewBucket(3)
	b.Refill(3)
	b.Refill(3)

	for i := 0; i < 3; i++ {
		token := <-b.TryGet(fixTimeout())
		assert.True(t, token)
	}

	token := <-b.TryGet(fixTimeout())
	assert.False(t, token)
}

func TestBucketTryGetWhenTokensDontExist(t *testing.T) {
	// GIVEN
	b := limiter.NewBucket(3)
	timeout := time.Millisecond
	// WHEN & THEN
	before := time.Now()
	select {
	case token := <-b.TryGet(timeout):
		assert.False(t, token)
	case <-time.After(time.Second):
		assert.Fail(t, "TryGet should return in %v", timeout)
	}
	// takes less than second
	assert.Zero(t, time.Since(before).Round(time.Second))
}

func fixTotalTokens() int {
	return 100
}

func fixMaxTokensPerClient() int {
	return 10
}

func mockTimeAfterProvider(t *testing.T, amount int, expectedDuration time.Duration) limiter.TimeAfter {
	cnt := 0
	return func(duration time.Duration) <-chan time.Time {
		cnt++
		assert.Equal(t, expectedDuration, duration)
		out := make(chan time.Time)
		go func(cnt int) {

			if cnt < amount {
				out <- time.Now()
			}
			close(out)
		}(cnt)
		return out
	}

}

func TestLimiterDistributeTokensWhenNoClients(t *testing.T) {
	//t.SkipNow()
	bucketLimiter := limiter.NewLeakyBucketLimiter(
		fixTotalTokens(),
		fixMaxTokensPerClient(),
		time.Hour,
		time.Hour,
		limiter.WithCustomTimeAfter(mockTimeAfterProvider(t, 3, time.Hour)))

	done := make(chan struct{})
	go func(bl *limiter.LeakyBucketLimiter) {
		bucketLimiter.Start()
		done <- struct{}{}
	}(bucketLimiter)
	<-done

}

func TestLimiterDistributeFairlyTokens(t *testing.T) {
	//t.SkipNow()
	totalTokens := 10
	maxTokensPerClient := 10
	bucketLimiter := limiter.NewLeakyBucketLimiter(totalTokens, maxTokensPerClient, time.Hour, time.Hour)
	client1 := limiter.ClientID("client1")
	client2 := limiter.ClientID("client2")

	perClient := 5

	results := make(chan result, totalTokens+2)

	token1 := bucketLimiter.GetToken(client1, time.Second)
	token2 := bucketLimiter.GetToken(client2, time.Second)

	go func(bl *limiter.LeakyBucketLimiter) {
		bucketLimiter.Start()
	}(bucketLimiter)
	done := make(chan int)
	go func() {
		done <- 0
	}()

	go func() {
		val := <-token1
		results <- result{clientID: client1, value: val}
		cnt := <-done
		done <- cnt + 1
	}()
	go func() {
		val := <-token2
		results <- result{clientID: client2, value: val}
		cnt := <-done
		done <- cnt + 1
	}()

	for i := 0; i < perClient; i++ {
		go func() {
			results <- result{clientID: client1, value: <-bucketLimiter.GetToken(client1, time.Second)}
			cnt := <-done
			done <- cnt + 1
		}()
		go func() {
			results <- result{clientID: client2, value: <-bucketLimiter.GetToken(client2, time.Second)}
			cnt := <-done
			done <- cnt + 1
		}()

	}

	for {
		val := <-done
		if val == totalTokens+2 {
			close(results)
			break
		}
		done <- val
		runtime.Gosched()
	}

	client1_ok := 0
	client1_failed := 0
	client2_ok := 0
	client2_failed := 0

	for res := range results {
		if res.clientID == client1 {
			if res.value {
				client1_ok++
			} else {
				client1_failed++
			}
		} else {
			if res.value {
				client2_ok++
			} else {
				client2_failed++
			}
		}
	}
	assert.Equal(t, 5, client1_ok)
	assert.Equal(t, 1, client1_failed)
	assert.Equal(t, 5, client2_ok)
	assert.Equal(t, 1, client2_failed)

}

type result struct {
	clientID limiter.ClientID
	value    bool
}

func TestLimiterDistributeTokensWhenMoreClientsThanTokens(t *testing.T) {
	//t.SkipNow()
	totalTokens := 10
	clientsCnt := 15
	maxTokensPerClient := 5
	bucketLimiter := limiter.NewLeakyBucketLimiter(totalTokens, maxTokensPerClient, time.Hour, time.Hour)

	clients := make([]limiter.ClientID, clientsCnt)
	results := make(chan result, clientsCnt)

	done := make(chan int)
	go func() { done <- 0 }()
	for i := 0; i < clientsCnt; i++ {
		clients[i] = limiter.ClientID(fmt.Sprintf("client_%d", i))
		ret := bucketLimiter.GetToken(clients[i], time.Second)
		go func(id limiter.ClientID) {
			v := <-ret
			fmt.Println("v",v)
			results <- result{clientID: id, value: v}
			val := <-done
			val++
			done <- val
		}(clients[i])
	}

	go func(bl *limiter.LeakyBucketLimiter) {
		bucketLimiter.Start()
	}(bucketLimiter)

	for {
		val := <-done
		if val == clientsCnt {
			close(results)
			break
		}
		done <- val
		runtime.Gosched()
	}

	var ok, failed int
	perClientMap := make(map[limiter.ClientID]bool)
	for res := range results {
		if res.value {
			ok++
		} else {
			failed++
		}
		_, exist := perClientMap[res.clientID]
		assert.False(t, exist)
		perClientMap[res.clientID] = true
	}

	assert.Equal(t, 10, ok)
	assert.Equal(t, 5, failed)

}

func TestRemoveInactiveClients(t *testing.T) {
	//t.SkipNow()
	refillPeriod := time.Second
	inactiveTime := time.Second * 3
	bucketLimiter := limiter.NewLeakyBucketLimiter(2, 2, refillPeriod, inactiveTime)
	client1 := limiter.ClientID("id_1")
	client2 := limiter.ClientID("id_2")

	res1 := bucketLimiter.GetToken(client1, time.Second)
	res2 := bucketLimiter.GetToken(client2, time.Second)

	go func() {
		fmt.Println("res1", <-res1)
		fmt.Println("res2", <-res2)

	}()
	go func() {
		bucketLimiter.Start()
	}()
	time.Sleep(refillPeriod)
	start := time.Now()
	for time.Since(start) < inactiveTime {
		// keep client 2 alive
		<-bucketLimiter.GetToken(client2, time.Second)
	}

	res5 := bucketLimiter.GetToken(client2, time.Second)
	res6 := bucketLimiter.GetToken(client2, time.Second)

	done := make(chan bool)
	go func() {
		done <- <-res5 && <-res6

	}()
	assert.True(t, <-done)

}
