# Terrakube CLI - Agent Instructions

## Project Overview

Terrakube CLI (`terrakube`) is a command-line tool for managing Terrakube Server resources: organizations, workspaces, modules, variables, teams, and jobs. It handles the full Terraform lifecycle (plan, apply, destroy) through the Terrakube API.

- **Language**: Go 1.25
- **Module path**: `terrakube`
- **CLI framework**: [cobra](https://github.com/spf13/cobra) + [viper](https://github.com/spf13/viper)
- **API client**: [github.com/terrakube-io/terrakube-go](https://github.com/terrakube-io/terrakube-go) (JSON:API via [google/jsonapi](https://github.com/google/jsonapi))
- **Config**: `~/.terrakube-cli.yaml` or `--config` flag
- **Auth**: Bearer token via config or `TERRAKUBE_TOKEN` env var
- **Output formats**: json (default), yaml, table, tsv, none

## Architecture

```
main.go              # Entry point, calls cmd.Execute()
cmd/                 # Cobra command definitions
  root.go            # Root command, config init, output rendering
  <resource>.go      # Parent command for each resource (organization, workspace, etc.)
  <resource>_<verb>.go  # Hand-written subcommands: create, list, update, delete
  template.go        # Framework-registered resource (via resource.Register)
  vcs.go             # Framework-registered resource
  workspace_tag.go   # Framework-registered resource
internal/
  output/            # Output rendering (json, yaml, table, tsv, none)
    renderer.go      # Render(w, data, format) function
  resource/          # Generic resource framework
    resource.go      # Config[T], Register[T](), field population
    resolve.go       # Parent scope name resolution
testutil/            # Test infrastructure
  server.go          # HTTP test server for API mocking
  assertions.go      # Common test assertion helpers
  fixtures.go        # Shared test data
```

### Two patterns for resources

**Hand-written (legacy)** -- used by organization, workspace, module, variable, team, job. Each verb is a separate file (`cmd/<resource>_<verb>.go`) with its own Cobra command, flag setup, and API call.

**Framework (preferred for new resources)** -- used by template, vcs, workspace-tag. A single `cmd/<resource>.go` file calls `resource.Register[T]()` with a `resource.Config[T]` struct. The framework generates list, get, create, update, and delete subcommands automatically.

### Framework details

- `resource.Config[T]` holds Name, Aliases, Parents, Fields, and API closures (List/Get/Create/Update/Delete).
- `resource.ParentScope` defines parent resources (e.g., organization) with `--<parent>-id` and `--<parent>-name` flags. Name flags use a `Resolver` function for automatic name-to-ID resolution.
- `resource.FieldDef` maps CLI flags to struct fields with type safety (String, Bool, Int).
- `resource.Register[T]()` creates a parent command and up to 5 subcommands. Only non-nil API closures get subcommands.
- On create, all `Required` fields are enforced. On update, only flags that were explicitly set are applied.
- Output is handled by `internal/output.Render(w, data, format)`.

### Adding a new resource (framework pattern)

Create `cmd/<resource>.go` with an `init()` that calls `resource.Register[T]()`. Typical registration is ~30 lines. See `cmd/template.go` for a complete example.

### Adding a new subcommand (hand-written pattern)

Follow existing `cmd/<resource>_<verb>.go` pattern. Register as subcommand of the resource parent. Use `viper` for flag binding with env var support (`TERRAKUBE_` prefix).

### Flag handling

Flags are bound to viper. Environment variables use the `TERRAKUBE_` prefix. Config file values are lowest priority, then env vars, then flags.

### Output rendering

Call `output.Render(w, data, format)` from `internal/output/renderer.go`. Supports json, yaml, table, tsv, none formats. Table rendering uses reflection to extract struct fields; fields tagged with `jsonapi "relation,..."` are skipped.

## Building and Testing

```bash
mise run build          # go build ./...
mise run test           # go test -race ./...
mise run lint           # golangci-lint run ./...
mise run vulncheck      # govulncheck ./...
mise run check          # all of the above
```

Dev environment: `mise install` sets up Go 1.25.7, golangci-lint, govulncheck, gofumpt.

## CI/CD

- **PR builds**: `.github/workflows/go.yml` - build, test (with `-race`), golangci-lint, govulncheck, `go mod tidy` check, coverage delta PR comment
- **CodeQL**: `.github/workflows/codeql.yml` - SAST analysis on PRs and weekly
- **Releases**: `.github/workflows/release.yml` - cross-compiles (linux/windows/darwin, 386/amd64), publishes to GitHub Releases, bumps Homebrew formula
- **Pre-commit**: `.pre-commit-config.yaml` - golangci-lint, build, test, go mod tidy, govulncheck

## Conventions

- Test infrastructure in `testutil/` (`server.go`, `assertions.go`, `fixtures.go`). Tests in `cmd/cmd_test.go`, `cmd/render_test.go`.
- Error handling: use `cobra.CheckErr()` for fatal errors in commands, return errors from client methods.
- Environment variable naming: `TERRAKUBE_<FLAG_NAME>` (uppercase, underscores).
- Commit messages: conventional commits format `type(scope): description`.
