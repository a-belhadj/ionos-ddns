.PHONY: run build clean setup lint test up down status logs publish-doc

setup:
	go install golang.org/dl/go1.25.6@latest
	go1.25.6 download

run:
	export $$(cat .env | xargs) && go run ./cmd/dyndns

build:
	go build -o bin/dyndns ./cmd/dyndns

clean:
	rm -rf bin/

lint:
	go vet ./...

test:
	go test -v -race ./...

up:
	podman-compose up -d

down:
	podman-compose down

status:
	podman-compose ps

logs:
	podman-compose logs -f

publish-doc:
	cd docs && GIT_USER=a-belhadj npm run deploy
