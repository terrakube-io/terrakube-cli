# Terrakube CLI - Agent Instructions

## Project Overview

Terrakube CLI (`terrakube`) is a command-line tool for managing Terrakube Server resources: organizations, workspaces, modules, variables, teams, and jobs. It handles the full Terraform lifecycle (plan, apply, destroy) through the Terrakube API.

- **Language**: Go 1.24
- **Module path**: `terrakube`
- **CLI framework**: [cobra](https://github.com/spf13/cobra) + [viper](https://github.com/spf13/viper)
- **API client**: JSON:API via [google/jsonapi](https://github.com/google/jsonapi)
- **Config**: `~/.terrakube-cli.yaml` or `--config` flag
- **Auth**: Bearer token via config or `TERRAKUBE_TOKEN` env var
- **Output formats**: json (default), table, tsv, none

## Architecture

```
main.go              # Entry point, calls cmd.Execute()
cmd/                 # Cobra command definitions
  root.go            # Root command, config init, output rendering
  <resource>.go      # Parent command for each resource (organization, workspace, etc.)
  <resource>_<verb>.go  # Subcommands: create, list, update, delete
client/
  client/            # HTTP client layer
    client.go        # Base client: auth, request building, response handling
    <resource>.go    # Resource-specific API methods
  models/            # Data structs for API request/response
    <resource>.go    # JSON:API model definitions
```

### Patterns

- **Adding a new resource**: Create model in `client/models/`, client methods in `client/client/`, register client in `client/client/client.go` NewClient(), add cobra commands in `cmd/`.
- **Adding a new subcommand**: Follow existing `cmd/<resource>_<verb>.go` pattern. Register as subcommand of the resource parent. Use `viper` for flag binding with env var support (`TERRAKUBE_` prefix).
- **Flag handling**: Flags are bound to viper. Environment variables use the `TERRAKUBE_` prefix. Config file values are lowest priority, then env vars, then flags.
- **Output rendering**: Call `renderOutput(result, output)` with the result struct. Supports json, table, tsv, none formats. Table rendering uses reflection to extract struct fields.

## Building and Testing

```bash
go build -v ./...
go test -v ./...
```

## CI/CD

- **PR builds**: `.github/workflows/go.yml` - builds, tests, lints with golangci-lint
- **Releases**: `.github/workflows/release.yml` - cross-compiles (linux/windows/darwin, 386/amd64), publishes to GitHub Releases, bumps Homebrew formula

## Conventions

- No test files exist yet. When adding tests, use standard Go testing (`_test.go` files, `testing` package).
- Error handling: use `cobra.CheckErr()` for fatal errors in commands, return errors from client methods.
- Environment variable naming: `TERRAKUBE_<FLAG_NAME>` (uppercase, underscores).
- Commit messages: conventional commits format `type(scope): description`.
