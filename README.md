# circuit-breaker-poc
A simple POC to decide between two different golang circuit breaker libraries

# how to run
on a shell, in the project root directory, run the following:  
 `go run ./server/main.go`

on another shell, run the cli application:  
 `go run ./cli/*.go --sony=true --runs=200 --concurrency=5`

CLI parameters:  
`--sony` run sony cb implementation  
`--resiliency` run resiliency cb implementation  
`--runs` total number of cb requests to be made using the selected implementation  
`--concurrency` number of concurrent requests  

You can also run the client application as a HTTP API by doing the following:  
 `go run ./http/*.go`
 
This will start a HTTP server on `localhost:8080` with two [GET] endpoints available:  
 `/go-resiliency/breaker` when making requests to this endpoint will trigger a request using the *eapache/go-resiliency* implementation  
 `/sony/gobreaker` when making requests to this endpoint will trigger a request using the *sony/go-breaker* implementation  
