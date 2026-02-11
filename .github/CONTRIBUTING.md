# How to contribute to Terrakube CLI

First, thanks for taking the time to contribute to our project! There are many ways you can help out.

### Questions

If you have a question that needs an answer, [create an issue](https://github.com/terrakube-io/terrakube-cli/issues/new) and label it as a question.

### Issues for bugs or feature requests

If you encounter any bugs in the code, or want to request a new feature or enhancement, please [create an issue](https://github.com/terrakube-io/terrakube-cli/issues/new) to report it. Kindly add a label to indicate what type of issue it is.

### Contribute Code

We welcome your pull requests for bug fixes. To implement something new, please create an issue first so we can discuss it together.

#### Development Setup

1. Install [mise](https://mise.jdx.dev/) for tool management
2. Clone the repo and run `mise install` to set up Go, linters, and other tools
3. Install [pre-commit](https://pre-commit.com/) and run `pre-commit install`

#### Building and Testing

```bash
mise run build       # go build ./...
mise run test        # go test -race ./...
mise run lint        # golangci-lint run ./...
mise run vulncheck   # govulncheck ./...
mise run check       # all of the above
```

#### Creating a Pull Request

Please follow [best practices](https://github.com/trein/dev-best-practices/wiki/Git-Commit-Best-Practices) for creating git commits. Use conventional commit format: `type(scope): description`.
