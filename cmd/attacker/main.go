package main

import (
	"net/http"
	"fmt"
	"flag"
	"time"
	"io/ioutil"
)

func main() {
	sleepTime := flag.Duration("wait", 0, "")
	concurrency := flag.Int("c", 1, "")
	repeat := flag.Int("repeat", 1, "how many times repeat")
	endpoint := flag.String("endpoint", "http://localhost:5000/status", "endpoint to call")
	flag.Parse()

	tr := &http.Transport{
		MaxIdleConns:        *concurrency,
		MaxIdleConnsPerHost: *concurrency,
		DisableKeepAlives:   false,
	}
	httpClient := &http.Client{Transport: tr}

	outRes := NewResult()
	results := make(chan *result, *concurrency)

	for i := 0; i < *concurrency; i++ {

		go func(k int) {
			performGet(httpClient, fmt.Sprintf("client_%d", k), *endpoint, *repeat, *sleepTime, results)

		}(i)
	}

	for i := 0; i < * concurrency * *repeat; i++ {
		outRes = outRes.Merge(<-results)
	}
	outRes.Print()

}

func performGet(client *http.Client, clientID, endpoint string, repeat int, sleepTime time.Duration, cliResult chan *result) chan *result {
	for i := 0; i < repeat; i++ {

		go func() {
			req, err := http.NewRequest(http.MethodGet, endpoint, nil)
			if err != nil {
				fmt.Println("Gor error on request creation" + err.Error())
				cliResult <- resultOnError
				return
			}

			res := NewResult()
			req.Header.Set("client_id", clientID)
			before := time.Now()
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("Got error on GET " + err.Error())
				cliResult <- resultOnError
				return
			}

			res.AddStatusCode(resp.StatusCode, time.Since(before))

			fmt.Printf("%s: %d\n", clientID, resp.StatusCode)
			if resp.Body != nil {
				ioutil.ReadAll(resp.Body)
				resp.Body.Close()
			}
			cliResult <- res

		}()
		<-time.After(sleepTime)

	}
	return cliResult
}

type result struct {
	statusCode map[int]int
	times      []time.Duration
	noResult   bool
}

func NewResult() *result {
	return &result{
		statusCode: make(map[int]int),
		times:      make([]time.Duration, 0),
		noResult:   false,
	}
}

var resultOnError = &result{
	noResult: true,
}

func (r *result) AddStatusCode(code int, taken time.Duration) {
	prev, ex := r.statusCode[code]
	if !ex {
		r.statusCode[code] = 1
	} else {
		r.statusCode[code] = prev + 1
	}
	r.times = append(r.times, taken)
}

func (r *result) Merge(another *result) *result {
	if another == nil || another.noResult {
		return r
	}
	out := NewResult()
	for k, v := range r.statusCode {
		out.statusCode[k] = v
	}

	for k, v := range another.statusCode {
		_, ex := out.statusCode[k]
		if ex {
			out.statusCode[k] = out.statusCode[k] + v
		} else {
			out.statusCode[k] = v
		}
	}

	out.times = append(out.times, r.times...)
	out.times = append(out.times, another.times...)
	return out
}

func (r *result) Print() {
	for k, v := range r.statusCode {
		fmt.Printf("Code %d: %d\n", k, v)
	}

	var total int64
	for _, d := range r.times {
		total += d.Nanoseconds()
	}
	if len(r.times) == 0 {
		fmt.Print("Avg cannot be computed: no data")
	} else {
		fmt.Printf("Avg: %v \n", time.Duration(total/int64(len(r.times))))
	}
}
