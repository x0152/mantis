# Mantis Helm Chart

Kubernetes deployment for Mantis — Go backend, React/nginx frontend, PostgreSQL, and SSH sandboxes. Modeled after the Platform 3.0 chart but trimmed to what Mantis actually uses.

## What the chart installs

| Component | Workload | Exposed as |
|---|---|---|
| `app` | Deployment + Service | `app:8080` (ClusterIP) |
| `frontend` | Deployment + Service | `frontend:80` (ClusterIP) |
| `postgres` | Deployment + PVC + Service | `postgres:5432` (ClusterIP) |
| `migrate` | Job (`goose up`) | Helm post-install/post-upgrade hook |
| Sandboxes (base, browser, ffmpeg, python, db, netsec) | Deployment + Service per sandbox | `sandbox`, `browser-sandbox`, etc. (SSH) |
| `ingress` | Ingress | `/api → app`, `/ → frontend` |

Sandbox service names match `docker-compose.yml` hostnames (e.g. `browser-sandbox`, `ffmpeg-sandbox`, `netsec-sandbox`, `python-sandbox`, `db-sandbox`, `sandbox` for base), so any `Connection` entries created in the app continue to work unchanged.

## Prerequisites

- Kubernetes 1.24+
- An ingress controller (NGINX by default)
- A default `StorageClass` (or set `postgres.storageClassName` / `app.attachments.storageClassName`)
- Built and pushed images:
  - `mantis:<tag>` — from `./Dockerfile.prod`
  - `mantis-frontend:<tag>` — from `./frontend/Dockerfile.prod`
  - `mantis-sandbox`, `mantis-browser-sandbox`, `mantis-ffmpeg-sandbox`, `mantis-python-sandbox`, `mantis-db-sandbox`, `mantis-netsec-sandbox`

## Quick start

```bash
# Create namespace (or set global.createNamespace=true)
kubectl create namespace mantis

# Install
helm install mantis ./helm/mantis \
  --namespace mantis \
  --set ingress.host=mantis.example.com \
  --set app.image.tag=1.0.0 \
  --set frontend.image.tag=1.0.0
```

## Common overrides

```bash
# Point at your own registry
helm upgrade --install mantis ./helm/mantis \
  --namespace mantis \
  --set app.image.repository=registry.example.com/mantis \
  --set app.image.tag=$(git rev-parse --short HEAD) \
  --set frontend.image.repository=registry.example.com/mantis-frontend \
  --set frontend.image.tag=$(git rev-parse --short HEAD) \
  --set global.imagePullSecrets[0].name=registry-creds \
  --set ingress.host=mantis.example.com \
  --set ingress.tls[0].secretName=mantis-tls \
  --set ingress.tls[0].hosts[0]=mantis.example.com
```

## Secrets

`mantis-secrets` is generated from `values.yaml` (`secrets.postgres.*`). For real environments, either:

1. Override via `--set secrets.postgres.password=<strong>` (not recommended).
2. Pre-create a secret named `mantis-secrets` with keys `postgres-user`, `postgres-password`, `postgres-db`, and disable the template (TODO: add `secrets.create: false` if needed).

## External PostgreSQL

Disable the bundled DB and point the app at an external host:

```yaml
postgres:
  enabled: false
secrets:
  postgres:
    user: myuser
    password: mypassword
    db: mantis
```

Then edit `templates/app-deployment.yaml` `DATABASE_URL` — or override via a Secret with a full `DATABASE_URL` value. (The current chart builds the URL from user/password/db against service `postgres`; for external DBs prepare a custom `DATABASE_URL` secret.)

## Sandboxes

Every sandbox has a `enabled` flag in `values.yaml`. Disable the ones you don't need:

```yaml
sandboxes:
  browser:
    enabled: false
  netsec:
    enabled: false
```

SSH ports:
- `base` uses `2222` (linuxserver/openssh-server default)
- all others use `22`

Set up `Connection` entries inside the app using the Kubernetes service names as hosts.
