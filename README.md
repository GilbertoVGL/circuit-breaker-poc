# circuit-breaker-poc
A simple POC to decide between two different golang circuit breaker libraries

# how to run

on a shell, in the project root directory, run the following:
 go run ./server/main.go

on another shell, run the cli application:
 go run ./cli/*.go --sony=true --runs=200 --concurrency=5

CLI parameters:
--sony run sony cb implementation
--resiliency run resiliency cb implementation
--runs total number of cb requests to be made using the selected implementation
--concurrency number of concurrent requests
