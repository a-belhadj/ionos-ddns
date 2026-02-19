# Contributing

Thanks for your interest in contributing to IONOS DynDNS Updater!

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/<your-username>/ionos-ddns.git`
3. Create a branch: `git checkout -b my-feature`
4. Make your changes
5. Run tests: `go test -v -race ./...`
6. Commit and push your changes
7. Open a pull request

## Development

```bash
# Install Go (if needed)
make setup

# Run locally
cp .env.example .env
# Edit .env with your credentials
make run

# Run tests
go test -v -race ./...

# Build binary
make build
```

## Guidelines

- Keep changes focused and minimal
- Add tests for new functionality
- Follow existing code style
- Use meaningful commit messages

## Reporting Issues

Open an issue on GitHub with:
- What you expected to happen
- What actually happened
- Steps to reproduce
