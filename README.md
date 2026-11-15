# Terrakube CLI

For full CLI reference documentation, visit our official guide at [https://docs.terrakube.io/user-guide/terrakube-cli](https://docs.terrakube.io/user-guide/terrakube-cli).

---

## Table of Contents

- [Download & Installation](#download--installation)
  - [Homebrew (macOS / Linux)](#homebrew-macos--linux)
  - [packslip (verified install)](#packslip-verified-install)
  - [mise](#mise)
  - [Snap Package (Ubuntu / Linux)](#snap-package-ubuntu--linux)
  - [Linux & macOS](#linux--macos-bash--zsh)
  - [Windows](#windows-powershell)
  - [Docker](#docker)
- [Adding Terrakube CLI to PATH](#adding-terrakube-cli-to-path)
  - [Linux & macOS](#linux--macos)
  - [Windows](#windows)
- [Building & Testing from Source](#building--testing-from-source)
  - [Compiling the Binary](#compiling-the-binary)
  - [Building Snap Package Locally](#building-snap-package-locally)
  - [Running Unit Tests](#running-unit-tests)
  - [Running End-to-End (BATS) Tests](#running-end-to-end-bats-tests)
- [Quick Start Guide](#quick-start-guide)
  - [1. Login to Terrakube](#1-login-to-terrakube)
  - [2. Create an Organization](#2-create-an-organization)
  - [3. Create an Admin Team & Assign Permissions](#3-create-an-admin-team--assign-permissions)
  - [4. Create a Workspace (OpenTofu & Docker Compose)](#4-create-a-workspace-opentofu--docker-compose)
  - [5. Add Workspace Environment Variables & Tags](#5-add-workspace-environment-variables--tags)
  - [6. Configure Notifications (Slack / Teams / Webhook)](#6-configure-notifications-slack--teams--webhook)
  - [7. Run a Plan or Apply and Wait for It](#7-run-a-plan-or-apply-and-wait-for-it)
  - [8. Enforce Policies](#8-enforce-policies)
  - [9. Other Resource Fields](#9-other-resource-fields)
  - [10. Work with State](#10-work-with-state)
- [Breaking Changes](#breaking-changes)
- [Contributing](#contributing)
- [Security](#security)

---

## Download & Installation

Pre-compiled archives are published on [GitHub Releases](https://github.com/terrakube-io/terrakube-cli/releases/latest) for Linux and macOS (amd64, arm64) and Windows (amd64, 386). The CLI is also available from Homebrew, the Snap Store, packslip, mise and as a Docker image.

| Method | Platforms | Command |
|---|---|---|
| [Homebrew](#homebrew-macos--linux) | macOS, Linux | `brew install terrakube-io/cli/terrakube` |
| [packslip](#packslip-verified-install) | macOS, Linux | `packslip install github.com/terrakube-io/terrakube-cli` |
| [mise](#mise) | macOS, Linux | `mise use -g packslip:github.com/terrakube-io/terrakube-cli` |
| [Snap Store](#snap-package-ubuntu--linux) | Linux | `sudo snap install terrakube-cli` |
| [Release archive](#linux--macos-bash--zsh) | macOS, Linux, Windows | download from [GitHub Releases](https://github.com/terrakube-io/terrakube-cli/releases/latest) |
| [Docker](#docker) | any Docker host | `docker run --rm azbuilder/terrakube-cli --help` |
| [From source](#compiling-the-binary) | any Go platform | `go build -o terrakube main.go` |

Each method installs the `terrakube` command (the snap installs `terrakube-cli`, see below). Check it with `terrakube --version`.

### Homebrew (macOS / Linux)

```bash
brew install terrakube-io/cli/terrakube
terrakube --version
```

This one command adds the `terrakube-io/cli` tap and installs the formula; `brew upgrade terrakube` updates it.

### packslip (verified install)

Every release carries a signed [packslip](https://packslip.dev) bundle (`packslip.sigstore.json`) that lists the archives with their digests and links their build provenance. `packslip install` checks the signature and the digest before it installs `terrakube`:

```bash
packslip install github.com/terrakube-io/terrakube-cli
```

Get packslip itself with `mise use -g packslip` or `curl -fsSL https://packslip.sh | sh`. On Windows, use WSL.

### mise

[mise](https://mise.jdx.dev) installs the CLI through its `packslip` backend, which checks the release's signed packslip bundle before it installs the archive:

```bash
mise use -g packslip:github.com/terrakube-io/terrakube-cli         # latest release
mise use -g packslip:github.com/terrakube-io/terrakube-cli@x.y.z   # a specific version
terrakube --version
```

In a project, `mise use packslip:github.com/terrakube-io/terrakube-cli@x.y.z` pins the version in `mise.toml`, which also suits CI.

### Snap Package (Ubuntu / Linux)

Install from the Snap Store:

```bash
sudo snap install terrakube-cli
```

The snap exposes the command as `terrakube-cli`. To call it as `terrakube`, add an alias:

```bash
sudo snap alias terrakube-cli terrakube
terrakube --version
```

To install a locally built snap instead, see [Building Snap Package Locally](#building-snap-package-locally).

### Linux & macOS (Bash / Zsh)

Release archives are named `terrakube_<version>_<os>_<arch>.tar.gz` (version without the leading `v`). Download and extract one with `curl`:

```bash
# Example: Linux amd64; replace VERSION with a release from the Releases page
VERSION="x.y.z"
curl -sSLO "https://github.com/terrakube-io/terrakube-cli/releases/download/v${VERSION}/terrakube_${VERSION}_linux_amd64.tar.gz"
tar -xzf "terrakube_${VERSION}_linux_amd64.tar.gz" terrakube
```

### Windows (PowerShell)

Windows releases are `.zip` archives:

```powershell
# Example: Windows amd64; replace $Version with a release from the Releases page
$Version = "x.y.z"
Invoke-WebRequest -Uri "https://github.com/terrakube-io/terrakube-cli/releases/download/v$Version/terrakube_${Version}_windows_amd64.zip" -OutFile "terrakube.zip"
Expand-Archive terrakube.zip -DestinationPath .
```

### Docker

A multi-arch image (linux/amd64, linux/arm64) is published to Docker Hub as [`azbuilder/terrakube-cli`](https://hub.docker.com/r/azbuilder/terrakube-cli). The entrypoint is `terrakube`, so pass credentials as environment variables:

```bash
docker run --rm \
  -e TERRAKUBE_API_URL="https://terrakube-api.platform.local" \
  -e TERRAKUBE_TOKEN="XXXXXXXXXXXXX" \
  azbuilder/terrakube-cli organization list
```

---

## Adding Terrakube CLI to PATH

To run `terrakube` from any directory, add the binary location to your system's `PATH` environment variable.

### Linux & macOS

#### System-Wide Installation (Recommended)

Move the compiled or downloaded binary to `/usr/local/bin`:

```bash
sudo mv terrakube /usr/local/bin/
terrakube --version
```

#### User-Level Installation

If you prefer placing the binary in a custom folder (e.g., `~/.local/bin` or `~/bin`):

```bash
mkdir -p ~/.local/bin
mv terrakube ~/.local/bin/
```

Add the following line to your shell configuration file (`~/.bashrc` or `~/.zshrc`):

```bash
export PATH="$HOME/.local/bin:$PATH"
```

Reload shell settings:

```bash
source ~/.bashrc   # Or source ~/.zshrc
```

---

### Windows

#### Option A: PowerShell (Command Line)

To permanently append the folder containing `terrakube.exe` to your user `PATH`:

```powershell
# Assuming terrakube.exe is placed in C:\Tools\terrakube
$targetDir = "C:\Tools\terrakube"
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
[Environment]::SetEnvironmentVariable("Path", "$currentPath;$targetDir", "User")
```

Restart PowerShell and verify:

```powershell
terrakube --version
```

#### Option B: Windows GUI (System Properties)

1. Move `terrakube.exe` to a permanent folder (e.g., `C:\Tools\terrakube`).
2. Press `Win + R`, type `sysdm.cpl`, and press **Enter**.
3. Open the **Advanced** tab and click **Environment Variables**.
4. Under **User variables**, select **Path** and click **Edit**.
5. Click **New** and add the folder path (e.g., `C:\Tools\terrakube`).
6. Click **OK** to save and restart your command prompt or terminal.

---

## Building & Testing from Source

### Prerequisites

- [Go](https://go.dev/) 1.27 or higher

### Compiling the Binary

Clone the repository and compile using `go build`:

```bash
git clone https://github.com/terrakube-io/terrakube-cli.git
cd terrakube-cli
go build -o terrakube main.go
```

Verify compilation:

```bash
./terrakube --help
```

### Building Snap Package Locally

To build the Snap package on Ubuntu/Linux:

1. Install `snapcraft`:
   ```bash
   sudo snap install snapcraft --classic
   ```

2. Build the snap:
   ```bash
   # Build directly in container (LXD/Multipass):
   snapcraft

   # Or build locally in destructive mode:
   snapcraft --destructive-mode
   ```

3. Install the locally generated `.snap` package:
   ```bash
   sudo snap install --dangerous terrakube-cli_*.snap
   ```

### Running Unit Tests

Run unit tests across all Golang packages:

```bash
go test ./...
```

### Running End-to-End (BATS) Tests

Terrakube CLI includes comprehensive end-to-end integration tests written using [BATS](https://github.com/bats-core/bats-core).

> [!IMPORTANT]
> To execute the E2E test suite, you **must** have a running Terrakube Server test instance.
> Set the API endpoint URL (e.g., `https://terrakube-api.platform.local`) and an Admin Personal Access Token (PAT) in your environment variables:
>
> ```bash
> export TERRAKUBE_API_URL="https://terrakube-api.platform.local"
> export TERRAKUBE_PAT="XXXXXXXXXXXXX"  # Replace with a valid Terrakube Admin Personal Access Token
> ```
>
> The suite builds `./terrakube` and runs `terrakube login` with these values. It sets `HOME` to a temporary directory first, so your own `~/.terrakube-cli.yaml` is not modified.

Execute the E2E operations test suite:

```bash
bats tests/e2e_operations.bats
```

---

## Quick Start Guide

This step-by-step example demonstrates standard Terrakube CLI operations based on end-to-end workflows.

### 1. Login to Terrakube

Authenticate against your Terrakube server using your API URL and Personal Access Token (PAT):

```bash
terrakube login -a "https://terrakube-api.platform.local" -t "XXXXXXXXXXXXX"
```

`login` checks the credentials against the server and saves `api_url` and `token` to `~/.terrakube-cli.yaml` (or the file given with `--config`) with `0600` permissions. It also reads `TERRAKUBE_API_URL` and `TERRAKUBE_PAT` when the flags are omitted:

```bash
export TERRAKUBE_API_URL="https://terrakube-api.platform.local"
export TERRAKUBE_PAT="XXXXXXXXXXXXX"
terrakube login
```

To log in through the browser instead of with a PAT, use `--web`:

```bash
terrakube login --web --api-url "https://terrakube-api.platform.local"
```

`login --web` reads `login.v1` from the server's `/.well-known/terraform.json`, prints the login URL and opens it in the browser (`open` on macOS, `xdg-open` on Linux, `rundll32` on Windows; on a headless machine, copy the URL instead). It runs the OAuth authorization code flow with PKCE (S256) and waits up to `--timeout` (default `5m`) for the redirect to `http://localhost:<port>/login`, on the first free port in the range the server advertises. `--api-url` falls back to `TERRAKUBE_API_URL`, then to `api_url` in the config. If no advertised port is free, the callback's `state` does not match, the login is denied or the timeout passes, it exits with code 1 and saves nothing.

What it saves depends on the server:

- with the Terrakube login broker (`/oauth/authorize`), a Terrakube token; you choose its lifetime on the consent page, and it is used like a PAT;
- with Dex (`/auth`), the Dex access token, plus the `refresh_token`, the token's expiry and the token endpoint and client ID needed to renew it.

The Dex access token is short-lived. Before a command uses a saved Dex token that has expired or expires within a minute, the CLI exchanges the refresh token for a new one and saves it (mode `0600`). If that fails, the command exits with code 1 and `session expired, run terrakube login --web`. A token from `TERRAKUBE_TOKEN` or `TERRAKUBE_PAT`, and a PAT saved by `terrakube login`, are never refreshed. `terrakube login` with a PAT and `terrakube logout` remove the saved refresh data.

#### Workload identity in CI/CD

Inside GitHub Actions, GitLab CI, Bitbucket Pipelines or Azure Pipelines, `login --oidc` detects the platform, obtains its OIDC token and logs in with it. Terrakube accepts the token when a `federated` credential matches its issuer, audience and claims. `--terraform-credentials` also writes the token to `~/.terraform.d/credentials.tfrc.json` (`%APPDATA%\terraform.d` on Windows) for `terraform init` / `tofu init`; `--terraform-host` picks other hosts than the API URL's.

```yaml
# GitHub Actions
permissions:
  id-token: write
steps:
  - run: terrakube login -a https://terrakube-api.example.com --oidc --oidc-audience terrakube --terraform-credentials

# GitLab CI
plan:
  id_tokens:
    TERRAKUBE_OIDC_TOKEN:
      aud: terrakube
  script:
    - terrakube login -a https://terrakube-api.example.com --oidc --terraform-credentials
```

Bitbucket needs `oidc: true` on the step. Azure Pipelines needs a workload identity service connection, with `SYSTEM_ACCESSTOKEN: $(System.AccessToken)` and `AZURESUBSCRIPTION_SERVICE_CONNECTION_ID` mapped into the step `env`. Elsewhere, pass `--oidc-token-file` or set `TERRAKUBE_OIDC_TOKEN`. OIDC tokens are short-lived and never refreshed, so log in right before use.

`--oidc` requires an https API URL, and GitHub Actions requires `--oidc-audience` because the default audience is trusted by other services too. Azure Pipelines tokens always carry the Azure AD audience, so give the service connection no Azure role assignments. The token is stored in HOME, so clean it up on persistent self-hosted runners.

To run commands without a config file (for example in CI), set `TERRAKUBE_API_URL` and a token. The token is taken from the first of these that is set: `TERRAKUBE_TOKEN`, `TERRAKUBE_PAT`, the `token` saved in the config file. `terrakube logout` removes the stored token (a token in the environment still applies).

Parent-scope flags can also come from the environment when not passed on the command line, for example `TERRAKUBE_ORGANIZATION` (or `TERRAKUBE_ORGANIZATION_ID`) for `-o` and `TERRAKUBE_WORKSPACE` (or `TERRAKUBE_WORKSPACE_ID`) for `-w`, as do `TERRAKUBE_OUTPUT` and `TERRAKUBE_HIDE_NULLS`. Other flags such as `--name` or `--id` must be passed explicitly.

> [!NOTE]
> An exported `TERRAKUBE_WORKSPACE` / `TERRAKUBE_WORKSPACE_ID` also applies where the workspace is optional. For example, `notification-configuration list` and `create` then work on that workspace instead of the organization. Unset it to work at organization level.

Manage your personal access tokens and check who you are logged in as:

```bash
terrakube pat create --days 30 --description "ci pipeline"   # prints the token once, with its id
terrakube pat list
terrakube pat delete --id <PAT_ID>
terrakube whoami                                  # your teams (identity provider groups)
terrakube whoami --org my-org --workspace prod    # plus your effective permissions there
```

`--days 0` asks for a token that does not expire; the server refuses it when it caps token lifetimes. The server lets only a user logged in through the browser (`terrakube login --web`) create tokens; with a PAT login `pat create` gets 403.

### 2. Create an Organization

Create a new organization configured for remote execution:

```bash
terrakube organization create \
  --name "demo-org" \
  --description "Production Engineering Organization" \
  --execution-mode "remote" \
  --output json
```

The settings from the workspace's Settings page are flags of `workspace create|update`: `--template` (default template), `--require-approve-to-apply`, `--restrict-apply-to-configured-branch`, `--allow-remote-apply`, `--global-remote-state`, `--shared-ids` (comma-separated workspace IDs that may read the state) and `--module-ssh-key` (name or ID; `""` clears it).

*Save the returned organization ID (e.g., `09e86337-b6bb-49e0-82aa-c114ad0f41ab`) for subsequent operations.*

### 3. Create an Admin Team & Assign Permissions

Create a `TERRAKUBE_ADMIN` team inside your organization with admin permissions:

```bash
terrakube team create \
  -o "09e86337-b6bb-49e0-82aa-c114ad0f41ab" \
  --name "TERRAKUBE_ADMIN" \
  --role "Admin" \
  --manage-workspace \
  --manage-module \
  --manage-provider \
  --manage-state \
  --manage-collection \
  --manage-vcs \
  --manage-template \
  --manage-job \
  --plan-job \
  --approve-job \
  --destroy-job \
  --manage-policies \
  --output table
```

`workspace-access` grants a team rights on one workspace (`--plan-job`, `--approve-job`, `--destroy-job`, `--role`, and the `--manage-*` flags); `project-access` does the same for a project.

### 4. Create a Workspace (OpenTofu & Docker Compose)

Create a new workspace using OpenTofu (`tofu`) linked to the repository [terrakube-docker-compose](https://github.com/terrakube-io/terrakube-docker-compose):

```bash
terrakube workspace create \
  -o "09e86337-b6bb-49e0-82aa-c114ad0f41ab" \
  --name "docker-compose-infra" \
  --description "Docker Compose OpenTofu Workspace" \
  --source "https://github.com/terrakube-io/terrakube-docker-compose" \
  --branch "main" \
  --folder "/" \
  --iac-type "tofu" \
  --iac-version "1.12.5" \
  --execution-mode "remote" \
  --output json
```

### 5. Add Workspace Environment Variables & Tags

#### Add Environment Variables

Create environment variables for the workspace:

```bash
terrakube variable create \
  -o "09e86337-b6bb-49e0-82aa-c114ad0f41ab" \
  -w "<WORKSPACE_ID>" \
  --key "ENVIRONMENT" \
  --value "production" \
  --category "ENV" \
  --output table
```

#### Create and Associate Workspace Tags

1. Create an organization-level tag:
   ```bash
   terrakube tag create \
     -o "09e86337-b6bb-49e0-82aa-c114ad0f41ab" \
     --name "production" \
     --output json
   ```

2. Associate the tag with your workspace:
   ```bash
   terrakube workspace-tag create \
     -o "09e86337-b6bb-49e0-82aa-c114ad0f41ab" \
     -w "<WORKSPACE_ID>" \
     --tag-id "<TAG_NAME_OR_ID>" \
     --output table
   ```

### 6. Configure Notifications (Slack / Teams / Webhook)

1. Create an organization or workspace notification configuration:
   ```bash
   terrakube notification-configuration create \
     -o "09e86337-b6bb-49e0-82aa-c114ad0f41ab" \
     --name "slack-alerts" \
     --channel-type "SLACK" \
     --destination-url "https://hooks.slack.com/services/T00/B00/X00" \
     --message-style "DETAILED" \
     --active \
     --output json
   ```

   `--templates Plan,Apply` limits a configuration to jobs of those templates (names or IDs; the list replaces the current set, `--templates ""` empties it).

2. Add a trigger on job completion or failure:
   ```bash
   terrakube notification-trigger create \
     -o "09e86337-b6bb-49e0-82aa-c114ad0f41ab" \
     --notification-configuration "<NOTIFICATION_CONFIGURATION_ID>" \
     --job-status "completed" \
     --output table
   ```

---

### 7. Run a Plan or Apply and Wait for It

`run plan|apply|destroy` starts a job on a workspace. The template decides what the job does: pass `-t` with a template name or ID, or set `TERRAKUBE_TEMPLATE_PLAN`, `TERRAKUBE_TEMPLATE_APPLY` or `TERRAKUBE_TEMPLATE_DESTROY`. Without either, `run apply` uses the workspace's default template.

```bash
terrakube run plan -o "my-org" -w "prod" -t "Plan" --target module.network --wait --timeout 30m
echo $?   # 0 completed or no changes, 1 error, 2 failed, rejected or cancelled, 3 timed out
```

`job create --wait` and `job wait <job-id>` wait the same way, so a CI step can branch on the exit code. `job logs <job-id>` prints every step's log in order (`--step 200` for one step); `--follow` streams the running step until the job ends and exits with the same codes.

A job waiting for approval is listed with `job list -o my-org -w prod --status waitingApproval` and handled with `job approve <job-id>`, `job reject <job-id>` or `job cancel <job-id>`. Each checks the job's current status first and changes nothing if it does not allow the action. `workspace lock prod -o my-org --reason "release freeze"` stops new runs on a workspace until `workspace unlock prod -o my-org`.

A run trigger starts a run in one workspace after an apply in another. `-w` is the workspace that gets the run, `--source` the one whose apply fires it; without `--template` the run uses the destination's default template:

```bash
terrakube run-trigger create -o my-org -w app --source network --template Apply
terrakube run-trigger update -o my-org -w app --id <TRIGGER_ID> --enabled=false
```

### 8. Enforce Policies

A policy set points at a repository of OPA policies. Attach it to exactly one workspace, project or tag per attachment (names or IDs); a `policy-attachment` with no target or more than one fails before any request.

```bash
terrakube policy-set create -o my-org --name baseline --repository https://github.com/acme/policies --enforcement-level SOFT_MANDATORY
terrakube policy-set-parameter create -o my-org --policy-set baseline --key max_instances --value 3
terrakube policy-attachment create -o my-org --policy-set baseline --workspace prod
```

An exemption skips one rule of a policy set until it expires, for one workspace (`--workspace`), one project (`--project`) or, with neither, the whole organization. When a job stops on a soft-mandatory failure, read its evaluations and override one with a justification, then approve the job:

```bash
terrakube policy-exemption create -o my-org --policy-set baseline --rule deny_public_s3 \
  --justification "legacy bucket, migration in SEC-12" --ticket SEC-12 --expires-at 2027-01-31T00:00:00Z --workspace prod
terrakube policy-evaluation list -o my-org --job 42
terrakube policy-override create -o my-org --job 42 --policy-evaluation <EVALUATION_ID> --justification "accepted risk"
terrakube job approve 42 -o my-org
```

`policy-override create` records the user named in your token (email, else user name) as the one who overrode.

### 9. Other Resource Fields

- Templates: `template create --name Plan --content-file plan.yaml` reads the flow from a file instead of `--content`.
- Registry: `module-version` and `provider-version` take `--status DEPRECATED --deprecation-message "..."`; `provider` takes `--registry-namespace` (needed with `--imported`); `module` takes `--vcs` and `--ssh` (names or IDs) for private repositories.
- Jobs: `job create --variables region=eu-west-1 --variables size=3` passes run variables to that job only.
- `workspace-schedule` takes `--enabled` and `--description`, `workspace-tag` takes `--value`, and `variable` takes `--incomplete`.

### 10. Work with State

`state` reads a workspace's Terraform state through the API with your token, so no `curl` is needed. Workspaces are given by name or ID.

```bash
terrakube state pull prod -o my-org > prod.tfstate             # current state, byte for byte
terrakube state pull prod -o my-org --version <HISTORY_ID> -f old.tfstate
```

`--version` takes an ID from `history list` and writes that version's raw state (the same format as without it). The CLI only sends the token to the configured API: if a history entry points at another host, `state pull` fails with exit code 1 and sends nothing there.

`state rollback prod <HISTORY_ID> -o my-org` makes that version the current state, as the UI's rollback does. It asks you to type the workspace name; in scripts pass `--yes`. Without a terminal and without `--yes` it fails before any request.

`state diff prod <HISTORY_ID_A> <HISTORY_ID_B> -o my-org` lists what changed from version A to version B: resource instances by address and outputs by name, never their values.

```text
Resources: 1 added, 1 removed, 1 changed
  + aws_s3_bucket.logs[0]
  - aws_instance.old
  ~ module.net.aws_subnet.private["b"]
Outputs: 0 added, 0 removed, 1 changed
  ~ vpc_id
```

With an explicit `--output json` (or `yaml`) it prints `{"resources": {"added": [...], "removed": [...], "changed": [...]}, "outputs": {...}}`; `tsv` prints one `change kind address` line per difference.

`state outputs prod -o my-org` prints the outputs of the newest state. With `--output json` (the default) it prints the same map as `terraform output -json`, so CI can read it with `jq -r .vpc_id.value`; `table` and `tsv` print one line per output. Sensitive values print as `(sensitive)` unless `--show-sensitive` is set.

## Breaking Changes

Changes in behaviour that can affect existing scripts:

- `workspace delete` marks the workspace deleted (`deleted=true`), so the server also stops its schedules and jobs; `--hard` sends `DELETE` as before. `organization delete` disables the organization (`disabled=true`) after you type its name, or with `--yes`; without a terminal and without `--yes` it fails before any request.
- `create` sends only the flags you set; the server applies its defaults to the rest. Before, every attribute was sent, so an unset boolean reached the server as `false` (for example `job create` turned `refresh` off).
- `update` with no field flag fails before any request (`nothing to update: set at least one field flag, see --help`).
- Flags that take a template, VCS connection, SSH key, agent, project or tag accept a name as well as an ID. Lookup errors now read `template "x" not found in organization <id>` and `multiple agents match "x": <id>, <id>; pass the ID instead`.
- `job create` / `job update`: the job log field flag is now `--job-output`. `--output` is always the global output format and is validated before any API call.
- `job create`: `--command`/`-c` is replaced by `--template-reference`/`-t` (the template ID to run). The server never accepted `command`.
- Flags for attributes the API does not expose are removed: `webhook --path/--branch/--template-id/--event` (use `webhook-event`), `vcs --refresh-token/--token-expiration`, `address --type`, `agent --type`.
- `github-app-token --expires-at` takes an RFC 3339 timestamp (`2026-01-02T15:04:05Z`). Dates print as UTC RFC 3339 timestamps.
- A path in `api_url` is now kept as a prefix: `https://host/terrakube` sends requests to `https://host/terrakube/api/v1/...`. Before, the path was dropped, so an `api_url` ending in `/api/v1` must lose that suffix.
- Empty lists print `[]` in JSON instead of `null`.
- The config file only provides `api_url` and `token`. Other keys in it no longer fill flags.
- Environment variables no longer fill field flags (`--name`, `--id`, `--description`, ...). Only parent-scope flags, `--output`, and `--hide-nulls` are read from `TERRAKUBE_*` variables.
- Name lookups (for example `-o my-org`) require an exact name match.

---

## Contributing

We welcome contributions from the community! Please check out our [Contributing Guidelines](.github/CONTRIBUTING.md) for information on setting up your environment, code formatting conventions, submitting pull requests, and registering new CLI commands.

---

## Security

Security vulnerabilities or concerns should be reported responsibly. Please read our [Security Policy](.github/SECURITY.md) for details on supported versions and reporting procedures.
