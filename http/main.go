package main

import (
	"net/http"
	"time"
)

const (
	errThreshold      = 5
	successThreshold  = 5
	halfOpenThreshold = 5
	timeout           = time.Duration(10 * time.Second)
	numberOfRequests  = 1000
)

func main() {
	c := http.Client{Timeout: time.Duration(10 * time.Second)}
	http.HandleFunc("/go-resiliency/breaker", goResiliencyBreaker(c))
	http.HandleFunc("/sony/gobreaker", sonyBreaker(c))

	http.ListenAndServe("127.0.0.1:8080", nil)
}
