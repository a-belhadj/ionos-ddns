---
id: configuration
title: Configuration
sidebar_position: 2
---

# Configuration

All configuration is done via environment variables, typically in a `.env` file.

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `IONOS_API_KEY` | Yes | — | Your IONOS API key (`prefix_public.secret`) |
| `IONOS_DOMAINS` | Yes | — | Comma-separated list of domains |
| `UPDATE_INTERVAL_SECONDS` | No | `300` | Update interval in seconds |
| `LOG_LEVEL` | No | `INFO` | `DEBUG`, `INFO`, `WARN`, `ERROR` |
| `HEARTBEAT_INTERVAL_SECONDS` | No | `21600` | Heartbeat log interval (default: 6h) |
| `HEALTH_PORT` | No | `8080` | Health check endpoint port |

## Example `.env`

```bash
IONOS_API_KEY=prefix_public.secret
IONOS_DOMAINS=example.com,sub.example.com
UPDATE_INTERVAL_SECONDS=300
LOG_LEVEL=INFO
```

## Getting your IONOS API Key

1. Log in to your [IONOS Developer Portal](https://developer.hosting.ionos.com/)
2. Navigate to **API Keys**
3. Create a new key — the format is `prefix_public.secret`
