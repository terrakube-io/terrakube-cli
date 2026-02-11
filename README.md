# Terrakube CLI

`terrakube` is a command-line client for Terrakube Server. It brings organizations, workspaces and other Terrakube concepts to the terminal.

## Documentation

[See the manual](https://docs.terrakube.io/user-guide/terrakube-cli/install) for setup and usage instructions.

## Installation

### macOS

`terrakube` is available via [Homebrew](https://brew.sh) and as a downloadable binary from the [releases page][].

#### Homebrew

| Install:          | Upgrade:          |
| ----------------- | ----------------- |
| `brew install terrakube-io/cli/terrakube` | `brew upgrade terrakube` |

### Linux

Download packaged binaries from the [releases page][].

### Windows

`terrakube` is available via [Chocolatey](https://chocolatey.org) and as a downloadable binary from the [releases page][].

#### Chocolatey

| Install:           | Upgrade:           |
| ------------------ | ------------------ |
| `choco install terrakube` | `choco upgrade terrakube` |

### Build from source

Requires [Go](https://go.dev/) 1.25 or later.

```bash
git clone https://github.com/terrakube-io/terrakube-cli.git
cd terrakube-cli
go build -o terrakube .
```

## Contributing

If anything feels off, or if you feel that some functionality is missing, please check out the [contributing page](./.github/CONTRIBUTING.md). There you will find instructions for sharing your feedback, building the tool locally, and submitting pull requests to the project.

[releases page]: https://github.com/terrakube-io/terrakube-cli/releases/latest
