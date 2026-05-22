# Bot CI and Release Mirror Design

## Summary

Mirror the `jae-labs/flashcards` Go automation model for the Slack bot in this monorepo by adding bot-scoped CI and release workflows that preserve existing Terraform workflow boundaries.

## Goals

| Goal | Decision |
|---|---|
| Match `flashcards` workflow model | Use separate `ci.yml` and `release.yml` workflows |
| Keep Terraform automation isolated | Scope triggers to `bot/slack/**` and bot workflow files |
| Release the bot automatically from `main` | Create the next patch release on every push to `main` and refresh `latest` |
| Keep implementation repo-correct | Run all Go commands from `bot/slack` or `bot/slack/cmd/concierge` |

## Non-Goals

| Item | Reason |
|---|---|
| Reworking Terraform apply workflows | Out of scope and unrelated to bot delivery |
| Adding Homebrew publishing | `flashcards` has a tap integration; this repo does not |
| Turning the monorepo into a single Go project | The bot remains a nested Go module under `bot/slack` |

## Current Context

| Area | Current state |
|---|---|
| Go module | `bot/slack/go.mod` |
| Main package | `bot/slack/cmd/concierge` |
| Existing bot workflow | `.github/workflows/concierge-build.yml` runs `go build ./...` and `go test ./...` from `bot/slack` |
| Existing Terraform workflows | Remain path-scoped and must stay untouched |
| Reference repo | `flashcards` has `ci.yml` and `release.yml` for lint, test, build, security, and auto-release |

## Proposed Workflow Design

### `ci.yml`

Add a bot-only CI workflow that mirrors the `flashcards` job split.

| Job | Behavior |
|---|---|
| `lint` | Run `golangci-lint` from `bot/slack` |
| `test` | Run `go test -v -race -coverprofile=coverage.out ./...` from `bot/slack` |
| `build` | Build `bot/slack/cmd/concierge` across the same OS/arch matrix pattern used in `flashcards` |
| `security` | Run `gosec` and `trivy` against the bot code |

Trigger design:

| Event | Scope |
|---|---|
| `push` to `main` | Only when files under `bot/slack/**` or the bot workflow files change |
| `pull_request` targeting `main` | Same path scope |

### `release.yml`

Add a bot-only release workflow that mirrors the `flashcards` release structure while adapting paths and artifact names.

| Job | Behavior |
|---|---|
| `build-linux` | Build `concierge` for Linux `amd64` and `arm64`, package tarballs, upload artifacts |
| `build-darwin` | Build `concierge` for macOS `amd64` and `arm64`, package tarballs, upload artifacts |
| `release` | Download artifacts, compute next patch tag, create versioned release, recreate `latest` |

Trigger design:

| Event | Scope |
|---|---|
| `push` to `main` | Release runs on merges to `main`, matching `flashcards` |

Artifact naming:

| Platform | Binary name |
|---|---|
| Linux amd64 | `concierge-linux-amd64` |
| Linux arm64 | `concierge-linux-arm64` |
| macOS amd64 | `concierge-darwin-amd64` |
| macOS arm64 | `concierge-darwin-arm64` |

## Monorepo Adaptations

| Concern | Decision |
|---|---|
| Working directory | Use `bot/slack` for module-wide commands and `bot/slack/cmd/concierge` for binary builds |
| Workflow file naming | Use `ci.yml` and `release.yml` to align with `flashcards` |
| Existing bot CI file | Replace or retire `concierge-build.yml` so there is one bot CI entrypoint |
| Terraform workflows | Leave unchanged |

## Permissions and Secrets

| Workflow | Permissions | Secrets |
|---|---|
| `ci.yml` | `contents: read` | None beyond default GitHub Actions environment |
| `release.yml` | `contents: write` | Default `GITHUB_TOKEN` only |

No new release-publishing secrets are required because the Homebrew update step from `flashcards` is intentionally excluded.

## Error Handling and Safety

| Risk | Mitigation |
|---|---|
| Bot and Terraform workflows overlap | Use path filters so bot automation does not trigger Terraform apply workflows |
| Wrong module path in workflows | Anchor all Go setup and commands to `bot/slack` and `bot/slack/cmd/concierge` |
| Release drift from `flashcards` | Preserve the same release semantics and job shape, adapting only path- and artifact-specific details |
| Confusing duplicate CI | Remove or supersede the older bot-only workflow file |

## Validation Plan

| Area | Check |
|---|---|
| Local parity | Confirm documented local commands still build and test the bot |
| Workflow syntax | Validate the new workflow files in-repo |
| Bot behavior safety | Keep the existing `go build ./...` and `go test ./...` commands as the baseline checks |

## Documentation Updates

| File | Change |
|---|---|
| `bot/slack/README.md` | Add CI/release workflow details and release artifact expectations |

## Implementation Outline

1. Add `.github/workflows/ci.yml` for bot lint/test/build/security.
2. Add `.github/workflows/release.yml` for bot build and release publishing.
3. Remove or replace `.github/workflows/concierge-build.yml`.
4. Update `bot/slack/README.md` to document CI and release behavior.
