---
id: deployment
title: Deployment
sidebar_position: 3
---

# Deployment

## Docker Compose / Podman

```bash
make up       # Start
make down     # Stop
make status   # Container status
make logs     # Follow logs
```

## Kubernetes

```bash
kubectl apply -k k8s/
```

Includes namespace, deployment, configmap, secret, health probes, and security context. See the [Kubernetes README](https://github.com/a-belhadj/ionos-ddns/tree/main/k8s) for details.

## Binary

Download from [Releases](https://github.com/a-belhadj/ionos-ddns/releases) (Linux/macOS, amd64/arm64).

```bash
./ionos-ddns
```

## Health Check

The service exposes a `/healthz` endpoint on `HEALTH_PORT` (default `8080`):

```bash
curl http://localhost:8080/healthz
# {"status":"ok"}
```
