package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

var (
	successReqCount = 0
	errReqCount     = 0
	mu              sync.Mutex
)

func main() {
	http.HandleFunc("/get", genericEndpoint)

	http.ListenAndServe("127.0.0.1:6060", nil)
}

func genericEndpoint(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	r.ParseForm()
	cycle := r.Form.Get("cycle")

	ne := json.NewEncoder(w)

	if successReqCount >= 10 {
		errReqCount++
		log.Printf("client cycle: %s\t-\terrReqCount: %d\n", cycle, errReqCount)

		if errReqCount >= 10 {
			successReqCount = 0
			errReqCount = 0
		}

		w.WriteHeader(http.StatusInternalServerError)
		ne.Encode(`{ "badFoo": { "badBar": "badBaz" } }`)
	} else {
		successReqCount++
		log.Printf("client cycle: %v\t-\tsuccessReqCount: %d\n", cycle, successReqCount)
		ne.Encode(`{ "foo": { "bar": "baz" } }`)
	}
}
