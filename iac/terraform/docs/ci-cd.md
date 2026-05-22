# Repository CI/CD

GitHub Actions workflows at `.github/workflows/` in the repo root.

## Workflows

| Workflow | Trigger paths | Required secret |
|---|---|---|
| `ci.yml` | `bot/slack/**`, `.github/workflows/ci.yml`, `.github/workflows/release.yml` | None |
| `release.yml` | `bot/slack/**`, `.github/workflows/ci.yml`, `.github/workflows/release.yml` (push to `main` only) | Default `GITHUB_TOKEN` |
| `github-apply.yml` | `iac/terraform/github/**`, `iac/terraform/modules/github/**` | `GH_PAT` |
| `cloudflare-apply.yml` | `iac/terraform/cloudflare/**`, `iac/terraform/modules/cloudflare/**` | `CLOUDFLARE_API_TOKEN` |
| `doppler-apply.yml` | `iac/terraform/doppler/**`, `iac/terraform/modules/doppler/**` | `DOPPLER_TOKEN` |
| `oci-apply.yml` | `iac/terraform/oci/**` | OCI auth and stack secrets |

## Bot workflows

### `ci.yml`

The bot CI workflow mirrors the structure used in `jae-labs/flashcards` while targeting the nested Go module in `bot/slack/`.

1. Checks out code (`actions/checkout`, SHA-ratcheted)
2. Sets up Go from `bot/slack/go.mod`
3. Runs `golangci-lint`
4. Runs `go test -v -race -coverprofile=coverage.out ./...`
5. Uploads coverage to Codecov on a best-effort basis
6. Builds `cmd/concierge` across the Linux/macOS matrix used by the reference repo
7. Runs `gosec` and `trivy` against the bot subtree

### `release.yml`

The bot release workflow also mirrors `flashcards`, with monorepo path adjustments and no Homebrew publishing step.

1. Builds `cmd/concierge` artifacts for Linux and macOS
2. Packages tarballs and raw binaries
3. Uploads build artifacts between jobs
4. Computes the next patch release tag from the latest `v*.*.*` tag on `main`
5. Creates a versioned GitHub release plus a refreshed `latest` release

The workflow uses the default `GITHUB_TOKEN`; no extra release-publishing secrets are required.

## Reusable workflow

GitHub, Cloudflare, and Doppler call `terraform-reusable.yml` with inputs:

- `module-path` (string) -- path to root module
- `provider-token-name` (string) -- env var name for provider token

The reusable workflow:

1. Checks out code (`actions/checkout`, SHA-ratcheted)
2. Sets up Terraform (`hashicorp/setup-terraform`, ~> 1.5)
3. Writes GCP credentials to temp file from `GCP_SA_KEY` secret
4. Runs `terraform init` with `GOOGLE_APPLICATION_CREDENTIALS`
5. Runs `terraform plan` into a temporary local tfplan with stdout/stderr suppressed in GitHub Actions
6. Runs `terraform apply` from that temporary tfplan with stdout/stderr suppressed in GitHub Actions
7. Deletes the temporary tfplan, serializes applies per root module with a GitHub Actions concurrency group keyed by repository and `module-path`, and cleans up credentials (always runs)

## Dedicated OCI workflow

`oci-apply.yml` is separate because OCI needs multiple provider environment variables plus stack inputs.

The OCI workflow:

1. Checks out code (`actions/checkout`, SHA-ratcheted)
2. Sets up Terraform (`hashicorp/setup-terraform`, ~> 1.5)
3. Writes GCP backend credentials to a temp file from `GCP_SA_KEY`
4. Writes the OCI API private key to a temp PEM file from `OCI_PRIVATE_KEY`
5. Exports OCI provider env vars and `TF_VAR_*` stack inputs
6. Runs `terraform init`
7. Runs `terraform plan` into a temporary local tfplan with stdout/stderr suppressed in GitHub Actions
8. Runs `terraform apply` from that temporary tfplan with stdout/stderr suppressed in GitHub Actions
9. Deletes the temporary tfplan, cleans up temporary credential files, and serializes applies for `iac/terraform/oci`

## Trigger

Bot CI runs on path-scoped pushes to `main` and pull requests. Bot releases run on path-scoped pushes to `main`. Terraform applies run on pushes to `main` affecting module-specific paths (see table).

## Secrets

Stored in the `production` environment on the `conCIerge` repo.

| Secret | Value |
|---|---|
| `GCP_SA_KEY` | Raw JSON contents of GCP service account key (GCS state backend) |
| `GH_PAT` | Fine-grained PAT with org admin permissions |
| `CLOUDFLARE_API_TOKEN` | Cloudflare API token with zone/DNS edit |
| `DOPPLER_TOKEN` | Doppler personal token |
| `OCI_TENANCY_OCID` | OCI tenancy OCID used by the provider |
| `OCI_USER_OCID` | OCI user OCID used by the provider |
| `OCI_FINGERPRINT` | Fingerprint for the OCI API signing key |
| `OCI_REGION` | OCI region for provider operations |
| `OCI_PRIVATE_KEY` | PEM contents of the OCI API signing key |
| `OCI_COMPARTMENT_OCID` | Compartment OCID for the OCI stack |
| `OCI_AVAILABILITY_DOMAIN` | Exact tenancy-prefixed OCI availability-domain name for the compute instance (for example `tjxx:eu-amsterdam-1-AD-1`) |
| `OCI_SSH_AUTHORIZED_KEYS` | SSH authorized keys content injected into the instance |
| `OCI_SSH_INGRESS_CIDR` | CIDR allowed to reach the instance over SSH |

## Security

- Secrets scoped to `production` environment only
- Environment restricted to protected branches (`main`)
- Branch protection requires PR with 1 approval before merge
- No direct pushes to `main`
- Action SHAs are ratchet-pinned in workflows

## SHA ratcheting

Action versions are pinned by SHA (not tag) with ratchet comments for auditability:

```yaml
- uses: actions/checkout@de0fac2e... # ratchet:actions/checkout@v6
```
