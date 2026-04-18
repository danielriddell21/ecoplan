# ecoplan deployment

GitOps-managed deployment of the ecoscan gRPC service to a Hetzner k3s node,
exposed as `api.ecoplan.uk` via a dedicated Cloudflare tunnel. A full
Prometheus / Loki / Tempo / Grafana observability stack runs alongside it.

## Layout

```
deploy/
├── terraform/    # Hetzner server + Cloudflare tunnel + DNS
├── argocd/       # ArgoCD Application CRDs (bootstrap)
└── manifests/    # what ArgoCD watches (service + obs stack)
```

## First-time bootstrap

### 1. Provision infra

```sh
cd deploy/terraform
cp terraform.tfvars.example terraform.tfvars   # fill in real values
terraform init -backend-config=backend.hcl
terraform apply
```

Outputs a post-apply checklist; follow it.

### 2. Install ArgoCD + sealed-secrets

```sh
export KUBECONFIG=~/.kube/hetzner-ecoplan.yaml

kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

kubectl apply -f deploy/argocd/sealed-secrets.yaml
# Wait for the controller pod to be Running in kube-system before sealing anything.
```

### 3. Seal the three secrets

The Bitnami chart installs the controller as `sealed-secrets` in `kube-system`,
which isn't kubeseal's default (`sealed-secrets-controller`), so every
invocation needs `--controller-name` and `--controller-namespace`.

```sh
# cloudflared tunnel token
kubectl create secret generic cloudflared-token \
  --from-literal=token="$(cd deploy/terraform && terraform output -raw tunnel_token)" \
  --dry-run=client -o yaml -n ecoplan \
  | kubeseal \
      --controller-name sealed-secrets \
      --controller-namespace kube-system \
      --format yaml > deploy/manifests/cloudflared/sealed-secret.yaml

# ecoscan service keys
kubectl create secret generic ecoscan-secret \
  --from-literal=anthropic-api-key="sk-ant-..." \
  --from-literal=ecoscan-api-key="..." \
  --dry-run=client -o yaml -n ecoplan \
  | kubeseal \
      --controller-name sealed-secrets \
      --controller-namespace kube-system \
      --format yaml > deploy/manifests/service/sealed-secret.yaml

# Grafana admin password
kubectl create secret generic grafana-secret \
  --from-literal=admin-password="$(openssl rand -base64 24)" \
  --dry-run=client -o yaml -n ecoplan \
  | kubeseal \
      --controller-name sealed-secrets \
      --controller-namespace kube-system \
      --format yaml > deploy/manifests/grafana/sealed-secret.yaml
```

Commit the three encrypted files to `trunk`.

### 4. Apply the ArgoCD Application

```sh
kubectl apply -f deploy/argocd/application.yaml
```

ArgoCD now syncs everything in `deploy/manifests/` into the `ecoplan` namespace.

## Verifying

```sh
# gRPC reachable via the tunnel
grpcurl -H "authorization: Bearer $ECOSCAN_API_KEY" \
  -d '{"barcode":"5000112546415","council_id":"portsmouth"}' \
  api.ecoplan.uk:443 recycling.v1.RecyclingService/CanItBeRecycledBarcode

# Grafana (port-forward — there is no public hostname for v1)
kubectl port-forward -n ecoplan svc/grafana 3000:3000
# Log in as admin with the sealed password.
# Prometheus datasource should go green; `up{job="ecoscan-service"}` should return 1.
```

## Release flow

1. Merge a conventional-commits feat/fix to `trunk`.
2. CI's `tag` job creates a new `vX.Y.Z` git tag.
3. `.github/workflows/release.yaml` builds and pushes `ghcr.io/danielriddell21/ecoplan:vX.Y.Z`
   and `:edge` to GHCR, then cuts a GitHub Release.
4. `.github/workflows/promote.yaml` retags `:latest`, rewrites
   `deploy/manifests/service/deployment.yaml`, commits `chore: deploy vX.Y.Z`
   to `trunk`. ArgoCD picks up the commit and rolls the pod.

Rollback: `git revert` the `chore: deploy vX.Y.Z` commit. ArgoCD syncs backwards.
