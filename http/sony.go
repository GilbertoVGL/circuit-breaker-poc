package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/sony/gobreaker"
)

func sonyBreaker(c http.Client) func(http.ResponseWriter, *http.Request) {
	// configuração do circuit breaker
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "sony/gobreaker",          // nome do circuit breaker
		MaxRequests: uint32(halfOpenThreshold), // numero de requests permitidas no estado half-open
		Interval:    timeout,                   // intervalo em que as contagens são resetadas no estado fechado
		Timeout:     timeout,                   // tempo de espera para tentar fechar o circuito
		ReadyToTrip: func(counts gobreaker.Counts) bool { // função que triga a mudança de estado
			return counts.ConsecutiveFailures > errThreshold
		},
		OnStateChange: func(name string, from, to gobreaker.State) { // função que triga quando o estado muda
			log.Printf("circuit breaker %s going\n\tfrom: %s\n\tto: %s", name, from, to)
		},
	})

	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("sony/gobreaker")

		resp, err := cb.Execute(func() (interface{}, error) {
			m := ""
			req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:6060/get", nil)
			if err != nil {
				return m, err
			}

			resp, err := c.Do(req)
			if err != nil {
				return m, err
			}
			defer resp.Body.Close()
			json.NewDecoder(resp.Body).Decode(&m)

			if resp.StatusCode >= 400 && resp.StatusCode <= 599 {
				return m, errors.New(fmt.Sprintf("request problem: %v", resp.StatusCode))
			}

			return m, nil
		})

		log.Printf("current cb state:\ntotal: \t%v\nsuccess: \t%v\nerrors: \t%v\nconsecutive success: \t%v\nconsecutive errors: \t%v\n", cb.Counts().Requests, cb.Counts().TotalSuccesses, cb.Counts().TotalFailures, cb.Counts().ConsecutiveSuccesses, cb.Counts().ConsecutiveFailures)

		if err != nil {
			log.Printf("something else happened: \n\tbody: %+v\n\terror: %+v\n", resp, err)
			return
		}

		log.Printf("action completed successfully: \n\t%+v\n", resp)
	}
}
