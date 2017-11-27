package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"net"
	"flag"
	"fmt"
	"time"

	"github.com/SAPHybrisGliwice/gopher-concurrency/cmd/site/middleware"
	"github.com/SAPHybrisGliwice/gopher-concurrency/cmd/site/limiter"
)



func main() {
	r := mux.NewRouter()
	r.HandleFunc("/status", delegateMiddleware.Limit(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(time.Second)
		rw.WriteHeader(http.StatusOK)
	}))

	fmt.Printf("Listening on port :%d\n", *port)
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		panic(err)
	}

	srv := http.Server{Handler: r}
	err = srv.Serve(l)
	if err != nil {
		panic(err)
	}
}

var delegateMiddleware middleware.LimiterMiddleware
var port *int

func init() {
	middleType := flag.String("middleware", "none", "Choose middleware type. Valid options are: none, global, leaky")
	maxConcurrentConnections := flag.Int("connections", 1000, "Max concurrent connections")
	maxConnectionPerUser := flag.Int("connPerUser", 20, "Max connections per user")
	timeout := flag.Duration("timeout", time.Second, "Request timeout")
	refillPeriod := flag.Duration("refillPeriod", time.Second, "Refill period")
	maxInactiveClientTime := flag.Duration("maxInactiveClientTime", time.Minute, "Max inactive client time")
	port = flag.Int("port", 5000, "Listen on specified port")

	flag.Parse()

	var concreteMiddleware middleware.LimiterMiddleware

	switch *middleType {
	case "none":
		fmt.Println("Running application without limiter")
		break
	case "global":
		fmt.Printf("Creating global limiter with max concurrentConnections: [%d]\n", *maxConcurrentConnections)
		concreteMiddleware = middleware.NewGlobalLimiter(*maxConcurrentConnections)
	case "leaky":
		fmt.Printf("Creating leaky bucket limiter, concurrent connections [%d], connection per user: [%d], timeout [%v] , "+
			"refill freq: [%v], inactive client time: [%v]\n",
			*maxConcurrentConnections, *maxConnectionPerUser, *timeout, *refillPeriod, *maxInactiveClientTime)
		bucketLimiter := limiter.NewLeakyBucketLimiter(*maxConcurrentConnections, *maxConnectionPerUser, *refillPeriod, *maxInactiveClientTime)
		// TODO start bucket limiter
		concreteMiddleware = middleware.NewLeakyBucketLimiterMiddleware(bucketLimiter, *timeout)

	default:
		flag.PrintDefaults()
		return
	}

	delegateMiddleware = middleware.NewDelegatingMiddleware(concreteMiddleware)
}