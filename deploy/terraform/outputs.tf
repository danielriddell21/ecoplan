output "server_ip" {
  description = "Hetzner server public IP"
  value       = hcloud_server.ecoplan.ipv4_address
}

output "tunnel_token" {
  description = "Cloudflare tunnel token — pipe into kubectl create secret after apply"
  value       = cloudflare_zero_trust_tunnel_cloudflared.ecoplan.tunnel_token
  sensitive   = true
}

output "post_apply" {
  description = "Manual steps to complete after terraform apply"
  value       = <<-EOT

    ── After apply ────────────────────────────────────────────────────────────

    1. Wait for k3s (runs via cloud-init, takes ~2 min after server boots):
         ssh root@${hcloud_server.ecoplan.ipv4_address}
         systemctl status k3s

    2. Copy kubeconfig:
         scp root@${hcloud_server.ecoplan.ipv4_address}:/etc/rancher/k3s/k3s.yaml ~/.kube/hetzner-ecoplan.yaml
         sed -i 's/127.0.0.1/${hcloud_server.ecoplan.ipv4_address}/' ~/.kube/hetzner-ecoplan.yaml
         export KUBECONFIG=~/.kube/hetzner-ecoplan.yaml

    3. Create namespace and tunnel secret (bootstrap — sealed-secret takes over after):
         kubectl create namespace ecoplan
         kubectl create secret generic cloudflared-token \
           --from-literal=token=$(terraform output -raw tunnel_token) \
           -n ecoplan

    4. Install ArgoCD:
         kubectl create namespace argocd
         kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

    5. Apply sealed-secrets controller + Application:
         kubectl apply -f ../argocd/sealed-secrets.yaml
         kubectl apply -f ../argocd/application.yaml

    6. Get ArgoCD initial password:
         kubectl -n argocd get secret argocd-initial-admin-secret \
           -o jsonpath="{.data.password}" | base64 -d

    7. Seal the remaining secrets (ANTHROPIC_API_KEY, ECOSCAN_API_KEY, grafana-admin)
       and commit the encrypted SealedSecret manifests — ArgoCD picks them up.

    ───────────────────────────────────────────────────────────────────────────
  EOT
}
