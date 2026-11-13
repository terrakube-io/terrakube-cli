# Security Policy

## Supported Versions

Only the latest release of `terrakube-cli` receives security updates.

| Version | Supported          |
| ------- | ------------------ |
| Latest  | :white_check_mark: |
| < 1.0   | :x:                |

---

## Reporting a Vulnerability

The Terrakube team takes security seriously. If you discover a security vulnerability within `terrakube-cli`, please report it responsibly:

- **Email**: Send vulnerability details to [security@terrakube.io](mailto:security@terrakube.io) (or submit via GitHub Private Vulnerability Reporting).
- **Include**:
  - Description of the vulnerability.
  - Proof of Concept (PoC) or steps to reproduce.
  - Affected versions.

**Please do NOT open public GitHub issues for security vulnerabilities.**

---

## Credentials & Token Safety

- Never hardcode Personal Access Tokens (PATs) or sensitive credentials in scripts or repositories.
- In CI/CD pipelines, pass credentials from a secret store as environment variables: `TERRAKUBE_API_URL` plus `TERRAKUBE_TOKEN` (or `TERRAKUBE_PAT`) let commands run without a config file.
- `terrakube login` stores the token in `~/.terrakube-cli.yaml` (or the `--config` file) with `0600` permissions. Run `terrakube logout` to remove it.
