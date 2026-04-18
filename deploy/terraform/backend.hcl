# Cloudflare R2 backend for Terraform state
# Setup:
#   1. Cloudflare dashboard → R2 → Create bucket → name it "ecoplan-tfstate"
#   2. R2 → Manage R2 API tokens → Create token (Object Read & Write on ecoplan-tfstate)
#   3. Replace ACCOUNT_ID below with your Cloudflare account ID
#   4. Set AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY from the R2 token before running terraform
#
# Local usage: terraform init -backend-config=backend.hcl

endpoints = {
  s3 = "https://3f92ad6c3854252d14a0c92b85753658.r2.cloudflarestorage.com"
}
bucket                      = "ecoplan-tfstate"
key                         = "terraform.tfstate"
region                      = "auto"
skip_credentials_validation  = true
skip_metadata_api_check      = true
skip_region_validation       = true
skip_requesting_account_id   = true
use_path_style               = true
