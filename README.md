# Terrakube CLI

For full CLI reference documentation, visit our official guide at [https://docs.terrakube.io/user-guide/terrakube-cli](https://docs.terrakube.io/user-guide/terrakube-cli).

---

## Table of Contents

- [Download & Installation](#download--installation)
  - [Snap Package (Ubuntu / Linux)](#snap-package-ubuntu--linux)
  - [Linux & macOS](#linux--macos)
  - [Windows](#windows)
  - [Docker](#docker)
- [Adding Terrakube CLI to PATH](#adding-terrakube-cli-to-path)
  - [Linux & macOS](#linux--macos-1)
  - [Windows](#windows-1)
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
- [Breaking Changes](#breaking-changes)
- [Contributing](#contributing)
- [Security](#security)

---

## Download & Installation

Pre-compiled archives are published on [GitHub Releases](https://github.com/terrakube-io/terrakube-cli/releases/latest) for Linux and macOS (amd64, arm64) and Windows (amd64, 386). The CLI is also available from the Snap Store and as a Docker image.

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

To run commands without a config file (for example in CI), set `TERRAKUBE_API_URL` and a token. The token is taken from the first of these that is set: `TERRAKUBE_TOKEN`, `TERRAKUBE_PAT`, the `token` saved in the config file. `terrakube logout` removes the stored token (a token in the environment still applies).

Parent-scope flags can also come from the environment when not passed on the command line, for example `TERRAKUBE_ORGANIZATION` (or `TERRAKUBE_ORGANIZATION_ID`) for `-o` and `TERRAKUBE_WORKSPACE` (or `TERRAKUBE_WORKSPACE_ID`) for `-w`, as do `TERRAKUBE_OUTPUT` and `TERRAKUBE_HIDE_NULLS`. Other flags such as `--name` or `--id` must be passed explicitly.

> [!NOTE]
> An exported `TERRAKUBE_WORKSPACE` / `TERRAKUBE_WORKSPACE_ID` also applies where the workspace is optional. For example, `notification-configuration list` and `create` then work on that workspace instead of the organization. Unset it to work at organization level.

### 2. Create an Organization

Create a new organization configured for remote execution:

```bash
terrakube organization create \
  --name "demo-org" \
  --description "Production Engineering Organization" \
  --execution-mode "remote" \
  --output json
```

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
  --output table
```

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
     --tag-id "<TAG_ID>" \
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

2. Add a trigger on job completion or failure:
   ```bash
   terrakube notification-trigger create \
     -o "09e86337-b6bb-49e0-82aa-c114ad0f41ab" \
     --notification-configuration "<NOTIFICATION_CONFIGURATION_ID>" \
     --job-status "completed" \
     --output table
   ```

---

## Breaking Changes

Changes in behaviour that can affect existing scripts:

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
