build:
	go build -o bin/stomper .

build-docker:
	docker build -t stomper-local:latest .

run:
	go run main.go

test:
	go test .

force-test:
	go clean -testcache
	go test .

race-test:
	go test -race .
