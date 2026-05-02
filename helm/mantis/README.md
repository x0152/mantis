# Mantis Helm Chart

Single-release deployment of the Mantis Go backend, React/nginx frontend,
in-cluster PostgreSQL, and DB migrations. Modeled after the
sapiens-expert chart used in production: same labels, same overlay
layout, same secret/config split.

## What the chart installs

| Component | Workload | Service |
|---|---|---|
| `app` | Deployment + PVC (attachments, optional dind storage) | `app:8080` (ClusterIP, fixed name)
| `frontend` | Deployment | `frontend:80` (ClusterIP, fixed name)
| `postgres` (optional) | Deployment + PVC | `postgres:5432` (ClusterIP, fixed name)
| `migrate` | Job (`goose up`) — regular resource, image-tag in name |
| `mantis-config` | ConfigMap (non-sensitive env)
| `mantis-secrets` | Secret (sensitive env)
| `mantis` | Ingress — `/api → app`, `/ → frontend`

Why service names are fixed (`app`, `frontend`, `postgres`):
- The frontend nginx image proxies `/api/` → `http://app:8080` (see
  `frontend/nginx.conf`).
- Code in `cmd/main.go` defaults `DATABASE_URL` to `postgres:5432`.
Renaming would require rebuilding both images.

Deployment, PVC, ConfigMap, Secret, Job, Ingress names use the standard
`<release>-mantis-*` prefix so multiple releases can coexist in the same
cluster (different namespaces).

## File layout

```
helm/mantis/
├─ Chart.yaml
├─ values.yaml             # prod-ready base
├─ values-stage.yaml       # stage overlay (deltas only)
├─ values-prod.yaml        # prod overlay (currently empty)
├─ secrets.example.yaml    # template for prod Secret shipped out-of-band
├─ README.md
└─ templates/
   ├─ _helpers.tpl         # fullname, labels, commonEnv (used everywhere)
   ├─ NOTES.txt
   ├─ configmap.yaml       # non-sensitive env
   ├─ secrets.yaml         # bootstrap Secret rendered from values.secrets.*
   ├─ app-deployment.yaml  # backend pod (+ optional dind sidecar)
   ├─ app-service.yaml
   ├─ app-pvc.yaml
   ├─ app-hpa.yaml         # rendered only when app.autoscaling.enabled
   ├─ dind-pvc.yaml        # rendered only in runtime.mode=dind
   ├─ frontend-deployment.yaml
   ├─ frontend-service.yaml
   ├─ postgres-deployment.yaml  # rendered only when postgres.enabled
   ├─ postgres-service.yaml
   ├─ postgres-pvc.yaml
   ├─ migration-job.yaml   # goose up, regular Job
   └─ ingress.yaml
```

## Quick start

```bash
helm lint ./helm/mantis -f ./helm/mantis/values.yaml -f ./helm/mantis/values-prod.yaml
helm template mantis ./helm/mantis -n mantis -f ./helm/mantis/values.yaml -f ./helm/mantis/values-prod.yaml > /tmp/manifest.yaml

# prod
helm upgrade --install mantis ./helm/mantis \
  --namespace mantis --create-namespace \
  -f ./helm/mantis/values.yaml \
  -f ./helm/mantis/values-prod.yaml \
  --set ingress.host=mantis.example.com \
  --set app.image.repository=registry.example.com/mantis \
  --set app.image.tag=$(git rev-parse --short HEAD) \
  --set frontend.image.repository=registry.example.com/mantis-frontend \
  --set frontend.image.tag=$(git rev-parse --short HEAD)

# stage
helm upgrade --install mantis ./helm/mantis \
  --namespace mantis-stage --create-namespace \
  -f ./helm/mantis/values.yaml \
  -f ./helm/mantis/values-prod.yaml \
  -f ./helm/mantis/values-stage.yaml
```

`--atomic --cleanup-on-fail --timeout=15m` is safe here: migrations run as
a regular Job (not a Helm hook), so `helm install` and the Postgres
Deployment race together — the Job's init container waits for Postgres
ready, the app pod waits for the `goose_db_version` table created by
the migration, and the entire stack converges before `--wait` returns.

## Secrets

Two ways:

**Default — Secret rendered from `values.secrets.*` in `templates/secrets.yaml`.**
Convenient for local installs, never for prod.

**Prod — out-of-band Secret.** Two flavors, pick one:

1. *CI replaces the file.* Ship `secrets-prod.yaml` (or `secrets-stage.yaml`)
   from a secure store (e.g. GitLab Secure Files / Vault / SealedSecrets);
   CI copies it to `templates/secrets.yaml` before `helm upgrade`. The
   chart's own `templates/secrets.yaml` is overwritten and the Secret is
   applied as a `pre-install,pre-upgrade` hook (see `secrets.example.yaml`).
   This matches the sapiens-expert flow exactly.

2. *External Secret.* Pre-create the Secret in the cluster (e.g. via
   SealedSecrets / ExternalSecrets) and pass `--set
   global.secretName=mantis-secrets` so the chart skips rendering its
   own.

Required keys (used by `_helpers.tpl` `mantis.commonEnv`):

| Key | Purpose |
|---|---|
| `POSTGRES_USER` | DB user |
| `POSTGRES_PASSWORD` | DB password |
| `POSTGRES_DB` | DB name |
| `AUTH_TOKEN` | Bootstrap token for the admin user (optional) |
| `GONKA_PRIVATE_KEY` | Optional preset Gonka wallet key |

## ConfigMap

Non-sensitive env (host/port/SSL mode, AUTH_*, GONKA_*, ASR/OCR/TTS API URLs)
lives in `mantis-config`. Override via `config.*` in values, or replace
externally via `--set global.configName=<existing-cm>`.

`POSTGRES_HOST` / `POSTGRES_PORT` fall back to the in-cluster Postgres
Service when `config.postgres.host` / `config.postgres.port` are empty
and `postgres.enabled=true`.

## Sandbox runtime

Sandboxes are **not** Helm-managed — the app provisions them through the
Docker API at runtime. The chart only chooses how the daemon is exposed:

| `runtime.mode` | What it does | When to use |
|---|---|---|
| `dind` (default) | Adds a privileged `docker:24-dind` sidecar to the app pod and shares its Unix socket via an `emptyDir`. Layers persist on `mantis-dind-storage` (PVC). | Any cluster that allows privileged pods. |
| `hostSocket` | Mounts the host's `/var/run/docker.sock` via `hostPath`. | k3s on Docker / kind / single-node dev VMs. |

Sandbox containers are dialed by IP returned by the Docker adapter — no
in-cluster DNS plumbing required. State is visible at the **Hosts** page
inside the app and via `/api/runtime/sandboxes`.

## Gonka wallet provisioning

Same as before — production image bundles the official `inferenced` binary
(`Dockerfile.prod`, pinned via `INFERENCED_VERSION` build arg, default
`v0.2.11`). Wallet creation runs the binary in an isolated keyring; only
address, raw hex private key, and BIP-39 mnemonic are returned to the
browser. The mnemonic is shown once and never persisted.

Configurable via `config.gonka.*` (ConfigMap) and `secrets.gonkaPrivateKey`
(Secret):

| Key | Purpose |
|---|---|
| `GONKA_DEFAULT_NODE_URL` | Default Source URL prefilled in the wallet step. |
| `GONKA_PRIVATE_KEY` | Optional preset private key. If set, the wizard prefills "Use existing wallet". |
| `GONKA_NODE_URL` | Optional preset Source URL for the same form. |
| `GONKA_INFERENCED_BIN` | Override the binary path. Defaults to `/usr/local/bin/inferenced`. |

## Prerequisites

- Kubernetes 1.24+
- An ingress controller (NGINX assumed)
- A default `StorageClass` (or set `postgres.storageClassName` /
  `app.attachments.storageClassName` / `runtime.dind.storage.storageClassName`)
- Built and pushed images:
  - `mantis:<tag>` — from `./Dockerfile.prod`
  - `mantis-frontend:<tag>` — from `./frontend/Dockerfile.prod`
- For `runtime.mode=dind`: cluster must allow privileged pods (managed
  offerings like GKE Autopilot do not). Use `runtime.mode=hostSocket`
  instead on Docker-hosted single-node clusters.

## Operational tips

```bash
kubectl -n mantis get pods,svc,ingress,pvc,job
kubectl -n mantis logs -l app.kubernetes.io/instance=mantis --tail=200
kubectl -n mantis logs job/$(kubectl -n mantis get jobs -o name | grep migrate | head -1)

# Re-run migration on same image tag (Job is immutable, must delete first)
kubectl -n mantis delete job -l app.kubernetes.io/component=migrate
helm upgrade mantis ./helm/mantis -n mantis --reuse-values
```
