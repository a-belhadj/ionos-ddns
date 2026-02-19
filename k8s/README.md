# Kubernetes Deployment

Deploy IONOS DynDNS updater on Kubernetes.

## Quick Start

1. **Edit the secret** with your credentials:
   ```bash
   # Edit k8s/secret.yaml and replace with your actual values
   nano k8s/secret.yaml
   ```

2. **Deploy with kubectl**:
   ```bash
   kubectl apply -f k8s/
   ```

3. **Or deploy with kustomize**:
   ```bash
   kubectl apply -k k8s/
   ```

The deployment will be created in the `ionos-ddns` namespace.


## Configuration

### Using Secret (recommended for production)

Edit `k8s/secret.yaml` to set your credentials:
- `IONOS_API_KEY`: Your IONOS API key
- `IONOS_DOMAINS`: Comma-separated list of domains

### Using ConfigMap

Edit `k8s/configmap.yaml` to adjust settings:
- `UPDATE_INTERVAL_SECONDS`: Update interval (default: 300)
- `LOG_LEVEL`: Log level (DEBUG, INFO, WARN, ERROR)
- `HEARTBEAT_INTERVAL_SECONDS`: Heartbeat log interval (default: 21600 = 6h)


## Uninstall

```bash
kubectl delete -f k8s/
# or
kubectl delete -k k8s/
```
