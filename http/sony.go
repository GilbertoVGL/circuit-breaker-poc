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
		MaxRequests: uint32(halfOpenThreshold), // numero de requests com sucesso para passar de half-open para closed
		Interval:    timeout,                   // intervalo em que as contagens são resetadas no estado fechado
		Timeout:     timeout,                   // tempo de espera para tentar passar de aberto para meio aberto o circuito
		ReadyToTrip: func(counts gobreaker.Counts) bool { // função que triga a mudança de estado
			// aqui a gente tem acesso ao struct Counts, como estratégia para virar a chave podemos:
			//  1 - verificar se o número de erros consecutivos ultrapassa um threshold;
			//  2 - verificar o número total de erros dentro de uma janela de tempo (usando o parâmetro Interval);
			//  3 - uma combinação de ambas as opções.
			log.Printf("requests so far: %d\t-\tready to trip check\t-\tConsecutiveFailures %d Vs. %d errThreshold\n", counts.Requests, counts.ConsecutiveFailures, errThreshold)
			return counts.ConsecutiveFailures > errThreshold || counts.TotalFailures > 300
		},
		OnStateChange: func(name string, from, to gobreaker.State) { // função que triga quando o estado muda
			log.Printf("\n**********************************************************\n* changing state\t-\tfrom: %s\tto: %s *\n**********************************************************\n", from, to)
		},
		IsSuccessful: func(err error) bool { // customiza o que o circuit breaker vai entender como uma requisição de sucesso
			return err == nil
		},
	})

	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("sony/gobreaker")

		_, err := cb.Execute(func() (interface{}, error) {
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

		if err != nil {
			switch err {
			case gobreaker.ErrOpenState:
				log.Printf("state open error: %v\n", err)
				w.WriteHeader(http.StatusServiceUnavailable)
			case gobreaker.ErrTooManyRequests:
				log.Printf("too many requests while half open error: %v\n", err)
				w.WriteHeader(http.StatusServiceUnavailable)
			default:
				log.Printf("something else happened: %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}

		log.Printf("action completed successfully\n")
	}
}
