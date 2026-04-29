# Mantis Helm Chart

Kubernetes deployment for Mantis — Go backend, React/nginx frontend, PostgreSQL. Modeled after the Platform 3.0 chart but trimmed to what Mantis actually uses.

## What the chart installs

| Component | Workload | Exposed as |
|---|---|---|
| `app` | Deployment + Service | `app:8080` (ClusterIP) |
| `frontend` | Deployment + Service | `frontend:80` (ClusterIP) |
| `postgres` | Deployment + PVC + Service | `postgres:5432` (ClusterIP) |
| `migrate` | Job (`goose up`) | Helm post-install/post-upgrade hook |
| `ingress` | Ingress | `/api → app`, `/ → frontend` |

## Sandboxes

Sandboxes are **not** Helm-managed workloads. The `app` pod ships an embedded
set of built-in Dockerfiles (`base`, `python`, `browser`, `ffmpeg`, `db`,
`netsec`, `runtimectl`) and seeds them into the `connections` table on
startup. It then builds and runs each sandbox container itself through the
Docker API.

The chart provides two ways to expose a Docker daemon to the app:

| `runtime.mode` | What it does | When to use |
|---|---|---|
| `dind` (default) | Adds a privileged `docker:24-dind` sidecar to the app pod and shares its Unix socket via an `emptyDir`. Image layers are kept on a PVC (`mantis-dind-storage`) so rebuilds stay fast across restarts. | Any Kubernetes cluster that allows privileged pods. Self-contained, no node prerequisites. |
| `hostSocket` | Mounts the host's `/var/run/docker.sock` via `hostPath`. | Single-node setups whose nodes already run Docker (k3s on Docker, kind, dev VMs). |

Mantis dials sandboxes by IP — the docker adapter reports container IPs from
the runtime and the app overwrites the `host` field of each connection
after starting the container. There is no DNS plumbing to maintain.

The list and state of running sandboxes is visible in the **Runtimes** page
inside the app, and can be inspected or managed via the
`/api/runtime/sandboxes` API.

## Prerequisites

- Kubernetes 1.24+
- An ingress controller (NGINX by default)
- A default `StorageClass` (or set `postgres.storageClassName` /
  `app.attachments.storageClassName` / `runtime.dind.storage.storageClassName`)
- Built and pushed images:
  - `mantis:<tag>` — from `./Dockerfile.prod`
  - `mantis-frontend:<tag>` — from `./frontend/Dockerfile.prod`
- For `runtime.mode=dind`: the cluster must allow privileged pods (managed
  offerings like GKE Autopilot do not). Use `runtime.mode=hostSocket` on
  Docker-hosted single-node clusters instead.

## Quick start

```bash
kubectl create namespace mantis

helm install mantis ./helm/mantis \
  --namespace mantis \
  --set ingress.host=mantis.example.com \
  --set app.image.tag=1.0.0 \
  --set frontend.image.tag=1.0.0
```

## Common overrides

```bash
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

`mantis-secrets` is generated from `values.yaml` (`secrets.postgres.*`). For
real environments, pre-create a secret named `mantis-secrets` with keys
`postgres-user`, `postgres-password`, `postgres-db`, or override via
`--set secrets.postgres.password=<strong>`.
