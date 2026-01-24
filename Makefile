.PHONY: run build clean setup

setup:
	go install golang.org/dl/go1.25.6@latest
	go1.25.6 download

run:
	export $$(cat .env | xargs) && go run ./cmd/dyndns

build:
	go build -o bin/dyndns ./cmd/dyndns

clean:
	rm -rf bin/
