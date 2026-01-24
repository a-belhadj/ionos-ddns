# Build stage
FROM docker.io/library/golang:1.25-alpine AS builder

WORKDIR /build
COPY go.mod ./
RUN go mod download
COPY cmd/ ./cmd/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dyndns ./cmd/dyndns

# Final stage
FROM scratch

LABEL org.opencontainers.image.source="https://github.com/a-belhadj/ionos-ddns"
LABEL org.opencontainers.image.description="Dynamic DNS updater for IONOS domains"
LABEL org.opencontainers.image.licenses="GPL-3.0"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/dyndns /dyndns
ENTRYPOINT ["/dyndns"]
