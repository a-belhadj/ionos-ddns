.PHONY: run build clean setup lint vuln test up down status logs

setup:
	go install golang.org/dl/go1.25.14@latest
	go1.25.14 download

run:
	export $$(cat .env | xargs) && go run ./cmd/dyndns

build:
	go build -o bin/dyndns ./cmd/dyndns

clean:
	rm -rf bin/

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint not found. Install it with:"; \
		echo "  go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"; \
		exit 1; \
	}
	golangci-lint run ./...

vuln:
	@command -v govulncheck >/dev/null 2>&1 || { \
		echo "govulncheck not found. Install it with:"; \
		echo "  go install golang.org/x/vuln/cmd/govulncheck@latest"; \
		exit 1; \
	}
	govulncheck ./...

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
