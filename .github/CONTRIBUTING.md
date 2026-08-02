# Contributing to Terrakube CLI

Thank you for your interest in contributing to `terrakube-cli`! This document outlines the guidelines for submitting bug reports, feature requests, code modifications, and adding new CLI subcommands.

---

## Code of Conduct

We expect all contributors to adhere to standard open-source community standards. Be respectful, inclusive, and collaborative.

---

## How to Contribute

### 1. Reporting Issues

Before opening a new issue, please search existing issues to see if it has already been reported. When filing an issue, include:
- A clear, descriptive title.
- Steps to reproduce the bug.
- Expected behavior vs. actual behavior.
- Terrakube CLI version (`terrakube --version`) and OS environment details.

### 2. Development Workflow

1. **Fork & Clone**: Fork the repository on GitHub and clone it locally.
2. **Create a Feature Branch**:
   ```bash
   git checkout -b feature/my-new-feature
   ```
3. **Make Code Changes**: Ensure code adheres to Go best practices and standard formatting.
4. **Run Formatters & Linters**:
   ```bash
   gofmt -w .
   ```
5. **Run Tests**:
   ```bash
   go test ./...
   ```
6. **Commit Changes**: Follow clear commit message conventions.
7. **Submit a Pull Request**: Push your branch to GitHub and submit a PR against the `main` branch.

---

## CLI Architecture & Resource Framework

All CLI commands in `terrakube-cli` follow a generic resource registration pattern. Avoid writing hand-written Cobra subcommands from scratch.

### Registering a New Resource

New resources are registered using the `resource.Register[T]()` framework:

```go
package myresource

import (
    "terrakube-cli/cmd/resource"
)

type MyResource struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
}

func init() {
    resource.Register[MyResource](resource.Config{
        Name:        "my-resource",
        Description: "Manage custom resources in Terrakube",
        Endpoint:    "/api/v1/my-resource",
    })
}
```

---

## Running End-to-End Tests

Integration tests use [BATS](https://github.com/bats-core/bats-core).

To run the end-to-end test suite locally:

1. Ensure a local test server instance is running (e.g. `https://terrakube-api.platform.local`).
2. Export your credentials:
   ```bash
   export TERRAKUBE_API_URL="https://terrakube-api.platform.local"
   export TERRAKUBE_PAT="XXXXXXXXXXXXX" # Terrakube Admin Personal Access Token
   ```
3. Execute the tests:
   ```bash
   bats tests/e2e_operations.bats
   ```
