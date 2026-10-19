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
| `IONOS_API_KEY` | Yes | — | Your IONOS API key (full `<public prefix>.<secret>` string, see below) |
| `IONOS_DOMAINS` | Yes | — | Comma-separated list of domains |
| `UPDATE_INTERVAL_SECONDS` | No | `300` | Update interval in seconds |
| `LOG_LEVEL` | No | `INFO` | `DEBUG`, `INFO`, `WARN`, `ERROR` |
| `HEARTBEAT_INTERVAL_SECONDS` | No | `21600` | Heartbeat log interval (default: 6h) |
| `HEALTH_PORT` | No | `8080` | Health check endpoint port |

## Example `.env`

```bash
IONOS_API_KEY=3f2a9c1b7d4e48a6b5c0d9e8f7a61234.Qm9xV2tYcEw3ZGhOa1JzTHpBdUZqQ0VpU0hnb1BuMTIzNDU2Nzg5MGFiY2RlZg
IONOS_DOMAINS=example.com,sub.example.com
UPDATE_INTERVAL_SECONDS=300
LOG_LEVEL=INFO
```

## Getting your IONOS API Key

1. Log in to your [IONOS Developer Portal](https://developer.hosting.ionos.com/)
2. Navigate to **API Keys**
3. Create a new key. IONOS shows a **public prefix** and a **secret**; the key to use is both joined by a dot: `<public prefix>.<secret>` (example above). The secret is only displayed once.
