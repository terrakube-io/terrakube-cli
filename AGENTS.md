# Terrakube CLI - Agent Instructions

## Project Overview

Terrakube CLI (`terrakube`) is a command-line tool for managing Terrakube Server resources: organizations, workspaces, modules, variables, teams, and jobs. It handles the full Terraform lifecycle (plan, apply, destroy) through the Terrakube API.

- **Language**: Go 1.27
- **Module path**: `terrakube`
- **CLI framework**: [cobra](https://github.com/spf13/cobra) + [viper](https://github.com/spf13/viper)
- **API client**: [github.com/terrakube-io/terrakube-go](https://github.com/terrakube-io/terrakube-go) (JSON:API via [google/jsonapi](https://github.com/google/jsonapi))
- **Config**: `~/.terrakube-cli.yaml` or `--config` flag
- **Auth**: `terrakube login` stores `api_url` and `token`; commands read them from the config or `TERRAKUBE_API_URL` / `TERRAKUBE_TOKEN`
- **Output formats**: json (default), yaml, table, tsv, none

## Architecture

```
main.go              # Entry point, calls cmd.Execute()
cmd/
  root.go            # Root command, config init, env handling, --output validation, newClient()
  login.go logout.go # Write/remove credentials in the config file
  resolvers.go       # Name-to-ID resolvers for parent flags (resolveByName + quoted RSQL nameFilter)
  <resource>.go      # One file per resource, each calls resource.Register
  team_token.go      # Hand-written commands (no generic CRUD shape)
  operations.go      # Hand-written: JSON:API atomic operations from --file
internal/
  output/renderer.go # Render(w, data, format): json, yaml, table, tsv, none
  resource/
    resource.go      # Config[T], Register[T](), flag/field population
    resolve.go       # Parent scope resolution (UUID passthrough or Resolver)
testutil/            # HTTP test server, assertions, fixtures
tests/e2e_operations.bats  # End-to-end suite against a live server
```

### Resource framework

Every resource is registered through `resource.Register[T](rootCmd, resource.Config[T]{...})` in `cmd/<resource>.go`. See `cmd/template.go` for a complete example.

- `Config[T]` holds Name, Aliases, Parents, Fields, and API closures (List/Get/Create/Update/Delete). Only non-nil closures get a subcommand.
- `ParentScope` adds one flag per parent (e.g. `--organization`/`-o`, alias `--org`, hidden `--organization-id`). The value is used as-is if it is a UUID (or `RawID` is set), otherwise it goes through `Resolver`. Resolvers live in `cmd/resolvers.go`; new ones should use `resolveByName` with `nameFilter`.
- `FieldDef` maps a flag to a struct field (String, Bool, Int, StringSlice). On create, `Required` fields are enforced. On update, only explicitly set flags are applied.
- List adds `--filter` (RSQL). Get, update, and delete take `--id`.
- Field flags must not be named `output` (it is the global format flag); the job log field is `--job-output`.

Hand-written commands (`login`, `logout`, `team-token`, `operations`) call `newClient()` themselves.

### Config, env, and flags

- Config file: `~/.terrakube-cli.yaml` or `--config`. `login` writes only `api_url` and `token` (mode 0600, other keys kept); `logout` removes `token`.
- `login` takes `--api-url`/`-a` and `--pat`/`-t`, falling back to `TERRAKUBE_API_URL` / `TERRAKUBE_PAT`.
- Other commands read `api_url` from `TERRAKUBE_API_URL` or the config. The token comes from `TERRAKUBE_TOKEN`, then `TERRAKUBE_PAT`, then the config `token`. If either value is missing the command fails with "not logged in".
- `applyEnv` (`cmd/root.go`) fills a flag from `TERRAKUBE_<FLAG>` (uppercase, `-` to `_`) only when the flag was not set explicitly, and only for:
  - parent-scope flags, via `resource.EnvAnnotation` (e.g. `TERRAKUBE_ORGANIZATION` or `TERRAKUBE_ORGANIZATION_ID`, `TERRAKUBE_WORKSPACE` or `TERRAKUBE_WORKSPACE_ID`). This includes `Optional` parents: an exported `TERRAKUBE_WORKSPACE` scopes `notification-configuration` to that workspace;
  - `--output` (`TERRAKUBE_OUTPUT`) and `--hide-nulls` (`TERRAKUBE_HIDE_NULLS`).
- Field flags (`--name`, `--id`, `--description`, ...) are never read from env or the config file.
- `--output` is validated before any API call.

### Output rendering

Call `output.Render(w, data, format)` from `internal/output/renderer.go`. Supports json, yaml, table, tsv, none formats. Table rendering uses reflection to extract struct fields; fields tagged with `jsonapi "relation,..."` are skipped. Empty lists render as `[]` in JSON.

## Building and Testing

```bash
mise run build          # go build ./...
mise run test           # go test -race ./...
mise run lint           # golangci-lint run ./...
mise run vulncheck      # govulncheck ./...
mise run check          # all of the above
```

Dev environment: `mise install` sets up Go 1.27.4, golangci-lint 2.15.0, govulncheck, gofumpt. Lint config is `.golangci.yml` (gofumpt formatter).

## CI/CD

All GitHub Actions are pinned by commit SHA. Dependabot (`.github/dependabot.yml`) updates Go modules and actions weekly.

- **PR builds**: `.github/workflows/go.yml` - build, test (with `-race`), golangci-lint, govulncheck, `go mod tidy` check, coverage delta PR comment
- **CodeQL**: `.github/workflows/codeql.yml` - SAST analysis on PRs and weekly
- **Releases**: `.github/workflows/release.yml` runs GoReleaser (`.goreleaser.yaml`) when a GitHub release is created: archives for linux and darwin (amd64, arm64) and windows (amd64, 386), the multi-arch `azbuilder/terrakube-cli` image on Docker Hub, and the `terrakube-cli` snap on the Snap Store (channel from the tag: `-rc` candidate, `-beta` beta, other pre-release edge, else stable)
- **Pre-commit**: `.pre-commit-config.yaml` - golangci-lint, build, test, go mod tidy, govulncheck

## Conventions

- Test infrastructure in `testutil/` (`server.go`, `assertions.go`, `fixtures.go`). Unit tests live next to the code (`cmd/<resource>_test.go`, `internal/*/*_test.go`).
- Error handling: commands use `RunE` and return errors; do not exit from inside a command.
- Commit messages: conventional commits format `type(scope): description`.
