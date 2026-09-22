# Build stage
FROM docker.io/library/golang:1.25-alpine AS builder

ENV GOFLAGS=-mod=readonly

WORKDIR /build
COPY go.mod ./
RUN go mod download
COPY cmd/ ./cmd/
# -trimpath strips local paths, -s -w drops symbol tables and DWARF data.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o dyndns ./cmd/dyndns

# Final stage
FROM scratch

LABEL org.opencontainers.image.source="https://github.com/a-belhadj/ionos-ddns"
LABEL org.opencontainers.image.description="Dynamic DNS updater for IONOS domains"
LABEL org.opencontainers.image.licenses="AGPL-3.0"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/dyndns /dyndns

# Run unprivileged (nobody); matches runAsUser in the Kubernetes manifests.
USER 65534:65534

ENTRYPOINT ["/dyndns"]
