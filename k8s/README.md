# Kubernetes Deployment

Deploy IONOS DynDNS updater on Kubernetes.

## Quick Start

1. **Create the secret** from the template. `k8s/secret.yaml` is git-ignored so
   your credentials are never committed:
   ```bash
   cp k8s/secret.example.yaml k8s/secret.yaml
   nano k8s/secret.yaml
   ```

   For production, prefer a secret manager over a local file, for example
   [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets) or
   [External Secrets](https://external-secrets.io/). In that case drop
   `secret.yaml` from `kustomization.yaml` and let the operator create the
   `ionos-dyndns-secret` Secret.

   You can also create it imperatively, with no file on disk:
   ```bash
   kubectl create secret generic ionos-dyndns-secret \
     --namespace ionos-ddns \
     --from-literal=IONOS_API_KEY=... \
     --from-literal=IONOS_DOMAINS=example.com
   ```

2. **Deploy with kustomize**:
   ```bash
   kubectl apply -k k8s/
   ```

   Use `kubectl apply -k` rather than `kubectl apply -f k8s/`, which would also
   pick up `secret.example.yaml` and the optional NetworkPolicy.

The deployment will be created in the `ionos-ddns` namespace.


## Configuration

### Using Secret (recommended for production)

Set your credentials in `k8s/secret.yaml` (created from the template above):
- `IONOS_API_KEY`: Your IONOS API key
- `IONOS_DOMAINS`: Comma-separated list of domains

### Using ConfigMap

Edit `k8s/configmap.yaml` to adjust settings:
- `UPDATE_INTERVAL_SECONDS`: Update interval (default: 300)
- `LOG_LEVEL`: Log level (DEBUG, INFO, WARN, ERROR)
- `HEARTBEAT_INTERVAL_SECONDS`: Heartbeat log interval (default: 21600 = 6h)
- `HEALTH_PORT`: Health check endpoint port (default: 8080)


### Network policy (optional)

`networkpolicy.yaml` restricts the pod to cluster DNS and outbound HTTPS only.
It is commented out in `kustomization.yaml` because it is silently ignored
unless your CNI plugin enforces NetworkPolicy (Cilium, Calico, Weave, ...).
Uncomment it once you have confirmed support.


## Uninstall

```bash
kubectl delete -k k8s/
```
