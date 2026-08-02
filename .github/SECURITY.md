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
- Use environment variables (`TERRAKUBE_PAT`, `TERRAKUBE_API_URL`) or secure secret management systems when running `terrakube-cli` in CI/CD pipelines.
