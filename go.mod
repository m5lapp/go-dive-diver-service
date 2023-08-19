module github.com/m5lapp/go-dive-diver-service

go 1.21

require (
	github.com/julienschmidt/httprouter v1.3.0
	github.com/lib/pq v1.10.9
	github.com/m5lapp/go-service-toolkit v0.0.0-20230622235322-4a0256d062fc
	golang.org/x/exp v0.0.0-20230522175609-2e198f4a06a1
)

require (
	github.com/tomasen/realip v0.0.0-20180522021738-f0c99a92ddce // indirect
	golang.org/x/time v0.3.0 // indirect
)

replace github.com/m5lapp/go-service-toolkit => ../go-service-toolkit
