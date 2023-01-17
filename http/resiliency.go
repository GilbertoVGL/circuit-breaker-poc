package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/eapache/go-resiliency/breaker"
)

func goResiliencyBreaker(c http.Client) func(http.ResponseWriter, *http.Request) {
	// configuração do circuit breaker
	cb := breaker.New(errThreshold, successThreshold, timeout)

	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("go-resiliency/breaker")
		m := ""

		result := cb.Run(func() error {
			req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:6060/get", nil)
			if err != nil {
				return err
			}

			resp, err := c.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			json.NewDecoder(resp.Body).Decode(&m)

			if resp.StatusCode >= 400 && resp.StatusCode <= 599 {
				return errors.New(fmt.Sprintf("request problem: %v", resp.StatusCode))
			}

			return nil
		})

		switch result {
		case nil:
			log.Printf("action completed successfully\n")
		case breaker.ErrBreakerOpen:
			log.Printf("state open error: %v\n", result)
			go func() {
				time.Sleep(timeout)
				log.Println("circuit should be half-open")
			}()
			w.WriteHeader(http.StatusServiceUnavailable)
		default:
			log.Printf("something else happened: %v\n", result)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
