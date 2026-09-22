.PHONY: run build clean setup lint vuln test up down status logs

setup:
	go install golang.org/dl/go1.27.1@latest
	go1.27.1 download

run:
	@# Previously `export $$(cat .env | xargs)`: xargs passes every value as
	@# argv to a child process, exposing the API key in ps / /proc/<pid>/cmdline.
	@# Sourcing keeps the values inside the shell.
	set -a; . ./.env; set +a; go run ./cmd/dyndns

build:
	go build -o bin/dyndns ./cmd/dyndns

clean:
	rm -rf bin/

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint not found. Install it with:"; \
		echo "  go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.2"; \
		exit 1; \
	}
	golangci-lint run ./...

vuln:
	@command -v govulncheck >/dev/null 2>&1 || { \
		echo "govulncheck not found. Install it with:"; \
		echo "  go install golang.org/x/vuln/cmd/govulncheck@v1.8.0"; \
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
