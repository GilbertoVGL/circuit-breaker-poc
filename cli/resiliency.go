package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/eapache/go-resiliency/breaker"
)

type Res struct {
	cb *breaker.Breaker
	c  http.Client
}

func goResiliencyBreaker(c http.Client) Res {
	log.Println("go-resiliency/breaker")

	// configuração do circuit breaker
	cb := breaker.New(errThreshold, successThreshold, timeout)

	// cria struct para auxiliar na execução dos testes
	return Res{cb: cb, c: c}
}

// Run roda a chamada envolta no circuit breaker, por conveniência ela só retornar erro de circuito aberto,
// para um melhor controler na execução do código
func (r Res) Run(cycle int) error {
	log.Printf("cycle: %d\t-\tstart\n", cycle)
	m := ""

	result := r.cb.Run(func() error {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:6060/get?cycle=%d", cycle), nil)
		if err != nil {
			return err
		}

		resp, err := r.c.Do(req)
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
		log.Printf("cycle: %d\t-\taction completed successfully\n", cycle)
		return nil
	case breaker.ErrBreakerOpen:
		log.Printf("cycle: %d\t-\tstate open error: %v\n", cycle, result)
		return result
	default:
		log.Printf("cycle: %d\t-\tsomething else happened: %v\n", cycle, result)
		return nil
	}
}
