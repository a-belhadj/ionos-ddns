# IONOS DynDNS Updater

[![Build and Push Docker Image](https://github.com/a-belhadj/ionos-ddns/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/a-belhadj/ionos-ddns/actions/workflows/docker-publish.yml)
[![Docker Image](https://ghcr-badge.egpl.dev/a-belhadj/ionos-ddns/latest_tag?trim=major&label=latest)](https://github.com/a-belhadj/ionos-ddns/pkgs/container/ionos-ddns)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

Automatically update your IONOS DNS records with your current public IP address.

The service periodically sends POST requests to `https://api.hosting.ionos.com/dns/v1/dyndns` to update your domain records. See the [IONOS DynDNS API documentation](https://developer.hosting.ionos.com/docs/dns) for details.

## Quick Start

```bash
# Configure
cp .env.example .env
nano .env

# Run with Docker Compose
docker-compose up -d

# Or run locally
make run
```

## Configuration

Configure via `.env` file:

```bash
IONOS_API_KEY=prefix_public.secret         # Required: Your IONOS API key
IONOS_DOMAINS=example.com,sub.example.com  # Required: Domains to update
UPDATE_INTERVAL_SECONDS=300                # Optional: Update interval in seconds (default: 300)
LOG_LEVEL=INFO                             # Optional: DEBUG, INFO, WARN, ERROR (default: INFO)
```

## Building

```bash
make run      # Run locally
make build    # Build binary
```

## API

Uses the [IONOS DynDNS API](https://developer.hosting.ionos.com/docs/dns#tag/Dynamic-DNS).
