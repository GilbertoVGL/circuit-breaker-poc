package main

import (
	"flag"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	errThreshold      = 5
	successThreshold  = 5
	halfOpenThreshold = 5
	timeout           = time.Duration(5 * time.Second)
	waitBetweenStates = time.Duration(4 * time.Second) // dorme esse tempo entre requests com circuit breaker aberto
)

var (
	sony        bool
	resiliency  bool
	runs        int
	concurrency int
)

// Como rodar na linha de comando ex.:
// 		go run ./cli/*.go --sony=true --runs=200 --concurrency=5
// 		go run ./cli/*.go --resiliency=true --runs=200 --concurrency=5
// 		go run ./cli/*.go --sony=true --resiliency=true --runs=200 --concurrency=5
func main() {
	flag.BoolVar(&sony, "sony", false, "run sony cb implementation")
	flag.BoolVar(&resiliency, "resiliency", false, "run resiliency cb implementation")
	flag.IntVar(&runs, "runs", 50, "total number of cb requests to be made")
	flag.IntVar(&concurrency, "concurrency", 1, "number of concurrent requests")
	flag.Parse()

	c := http.Client{Timeout: time.Duration(10 * time.Second)}

	if concurrency < 1 {
		concurrency = 1
	}

	if sony {
		runConcurrent(sonyBreaker(c))
	} else if resiliency {
		runConcurrent(goResiliencyBreaker(c))
	} else {
		log.Println("no valid implementation selected")
	}
}

type CbRunner interface {
	Run(cycle int) error
}

// Função para instrumentar as chamadas de maneira concorrente
func runConcurrent(cbR CbRunner) {
	// Instrumentação para rodar o teste de código com circuit breaker
	actualRuns := runs / concurrency
	remainder := runs % concurrency
	currentCount := 0 // current run count

	log.Printf("\ntotal runs: \t%d\nbatch runs size: \t%d\nconcurrency per batch requests: \t%d\n", runs, actualRuns, concurrency)

	// Loop que vai rodar as chamadas com o nível de concorrência passado como parâmetro
	for i := 0; i < actualRuns; i++ {
		var wg sync.WaitGroup
		var err error

		for j := 0; j < concurrency; j++ {
			currentCount++
			wg.Add(1)

			go func(j int, err *error) {
				defer wg.Done()

				// Chamada com circuit breaker
				if err2 := cbR.Run(j); err2 != nil {
					*err = err2
				}
			}(currentCount, &err)
		}

		wg.Wait()
		log.Println("err:", err)
		if err != nil {
			log.Println("waiting ", waitBetweenStates, "before other requests")
			time.Sleep(waitBetweenStates)
			err = nil
		}
		log.Printf("\n\n\n\n")
	}

	// Roda o resto das chamadas caso o número de chamadas não seja divisivel pelo nível de concorrência
	var wg sync.WaitGroup
	for i := 0; i < remainder; i++ {
		currentCount++
		wg.Add(1)

		go func(j int) {
			defer wg.Done()

			// Chamada com circuit breaker
			cbR.Run(j)
		}(currentCount)

	}
	wg.Wait()
}
