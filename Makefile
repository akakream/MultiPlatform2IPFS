build:
	go build -o bin/app

run:
	./bin/app

test:
	go test -v ./... -count=1

vendor:
	@go mod vendor
