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

		i := 0

		for {
			log.Println("i:", i)
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
				log.Printf("action completed successfully: \n\t%+v\n", m)
			case breaker.ErrBreakerOpen:
				log.Println("open circuit")
				go func() {
					time.Sleep(timeout)
					log.Println("should be half-open")
				}()
			default:
				log.Printf("something else happened: \n\tbody: %+v\n\terror: %+v\n", m, result)
			}

			i++
		}
	}
}
