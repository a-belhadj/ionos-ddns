# Contributing

Thanks for your interest in contributing to IONOS DynDNS Updater!

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/<your-username>/ionos-ddns.git`
3. Create a branch: `git checkout -b my-feature`
4. Make your changes
5. Run the checks: `make test && make lint && make vuln`
6. Commit and push your changes
7. Open a pull request

## Development

```bash
# Install Go (if needed)
make setup

# Run locally
cp .env.example .env
chmod 600 .env
# Edit .env with your credentials
make run

# Run tests
make test

# Lint (golangci-lint, includes gosec) and scan dependencies
make lint
make vuln

# Build binary
make build
```

`make lint` and `make vuln` need tooling that is not vendored:

```bash
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
```

## Guidelines

- Keep changes focused and minimal
- Add tests for new functionality
- Follow existing code style
- Use meaningful commit messages
- Never commit secrets; `.env` and `k8s/secret.yaml` are git-ignored on purpose

## Reporting Issues

Open an issue on GitHub with:
- What you expected to happen
- What actually happened
- Steps to reproduce
