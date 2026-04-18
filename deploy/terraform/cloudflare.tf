resource "random_bytes" "tunnel_secret" {
  length = 32
}

resource "cloudflare_zero_trust_tunnel_cloudflared" "ecoplan" {
  account_id = var.cloudflare_account_id
  name       = "ecoplan"
  secret     = random_bytes.tunnel_secret.base64
}

resource "cloudflare_zero_trust_tunnel_cloudflared_config" "ecoplan" {
  account_id = var.cloudflare_account_id
  tunnel_id  = cloudflare_zero_trust_tunnel_cloudflared.ecoplan.id

  config {
    ingress_rule {
      hostname = "api.${var.domain}"
      service  = "https://ecoscan-service:50051"

      origin_request {
        http2_origin  = true
        no_tls_verify = true
      }
    }
    ingress_rule {
      service = "http_status:404"
    }
  }
}

resource "cloudflare_record" "api" {
  zone_id = var.cloudflare_zone_id
  name    = "api"
  content = "${cloudflare_zero_trust_tunnel_cloudflared.ecoplan.id}.cfargotunnel.com"
  type    = "CNAME"
  proxied = true
}
