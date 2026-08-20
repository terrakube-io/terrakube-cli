# Terrakube CLI

For full CLI reference documentation, visit our official guide at [https://docs.terrakube.io/user-guide/terrakube-cli](https://docs.terrakube.io/user-guide/terrakube-cli).

---

## Table of Contents

- [Download & Installation](#download--installation)
- [Adding Terrakube CLI to PATH](#adding-terrakube-cli-to-path)
  - [Linux & macOS](#linux--macos)
  - [Windows](#windows)
- [Building & Testing from Source](#building--testing-from-source)
  - [Compiling the Binary](#compiling-the-binary)
  - [Running Unit Tests](#running-unit-tests)
  - [Running End-to-End (BATS) Tests](#running-end-to-end-bats-tests)
- [Quick Start Guide](#quick-start-guide)
  - [1. Login to Terrakube](#1-login-to-terrakube)
  - [2. Create an Organization](#2-create-an-organization)
  - [3. Create an Admin Team & Assign Permissions](#3-create-an-admin-team--assign-permissions)
  - [4. Create a Workspace (OpenTofu & Docker Compose)](#4-create-a-workspace-opentofu--docker-compose)
  - [5. Add Workspace Environment Variables & Tags](#5-add-workspace-environment-variables--tags)
  - [6. Configure Notifications (Slack / Teams / Webhook)](#6-configure-notifications-slack--teams--webhook)
- [Contributing](#contributing)
- [Security](#security)

---

## Download & Installation

You can download pre-compiled binaries directly from our [GitHub Releases](https://github.com/terrakube-io/terrakube-cli/releases/latest) page.

### Linux & macOS (Bash / Zsh)

Download the binary using `curl` or `wget`:

```bash
# Example: Download latest Linux amd64 binary
curl -sSL -o terrakube https://github.com/terrakube-io/terrakube-cli/releases/latest/download/terrakube-linux-amd64
chmod +x terrakube
```

### Windows (PowerShell)

Download the binary using PowerShell:

```powershell
# Download latest Windows amd64 binary
Invoke-WebRequest -Uri "https://github.com/terrakube-io/terrakube-cli/releases/latest/download/terrakube-windows-amd64.exe" -OutFile "terrakube.exe"
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

- [Go](https://go.dev/) 1.22 or higher

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

## Contributing

We welcome contributions from the community! Please check out our [Contributing Guidelines](.github/CONTRIBUTING.md) for information on setting up your environment, code formatting conventions, submitting pull requests, and registering new CLI commands.

---

## Security

Security vulnerabilities or concerns should be reported responsibly. Please read our [Security Policy](.github/SECURITY.md) for details on supported versions and reporting procedures.
