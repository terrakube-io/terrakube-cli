# Changelog

All notable changes to the Terrakube CLI are listed here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow [Semantic Versioning](https://semver.org/). From 1.0.0 on, [Compatibility](README.md#compatibility) in the README defines what counts as a breaking change.

## [Unreleased]

### Removed

- Chocolatey publishing (#91). Every push to push.chocolatey.org returned 504 Gateway Timeout, from both the Linux and the Windows client, and no version reached the community repository. Windows users install from the release zips.

## [1.0.0-rc.3] - 2026-11-15

The same CLI as rc.1; only the release pipeline changed.

### Changed

- The Chocolatey package is pushed from a Windows runner (#90). The push still returned 504, and #91 removes it.

## [1.0.0-rc.2] - 2026-11-15

The same CLI as rc.1; only the release pipeline changed.

### Fixed

- A failed Chocolatey push no longer fails the release job. For rc.1 it had skipped the packslip bundle and the Snap metadata upload (#89).

## [1.0.0-rc.1] - 2026-11-15

### Added

- `run plan|apply|destroy`, `job wait`, `job logs [--follow]`, `job approve|reject|cancel`, `workspace lock|unlock`. `job wait`, `job create --wait` and `job logs --follow` exit with a code that reflects how the job ended (#84).
- Policy sets, policy overrides and exemptions, run triggers, state versions and outputs (`state`), and resource fields that were missing (#85).
- `login --web` (browser login through the server's login broker or Dex, with automatic Dex token refresh), `pat` and `whoami` (#86).
- `login --oidc`: log in with the workload identity token of GitHub Actions, GitLab CI, Bitbucket Pipelines or Azure Pipelines. `--terraform-credentials` writes the token to `credentials.tfrc.json` (#65).
- Flags that take a template, VCS connection, SSH key, agent, project or tag accept a name as well as an ID (#83).
- Releases publish a Homebrew formula (`brew install terrakube-io/cli/terrakube`) and a signed [packslip](https://packslip.dev) bundle for `packslip install` and mise (#88).

### Changed

These changes can break existing scripts:

- `workspace delete` marks the workspace deleted (`deleted=true`), so the server also stops its schedules and jobs. `--hard` sends `DELETE` as before. `organization delete` disables the organization (`disabled=true`) once you type its name, or with `--yes`. Without a terminal and without `--yes` it fails before any request (#83).
- `create` sends only the flags you set, and the server applies its defaults to the rest. Before, every attribute was sent, so an unset boolean reached the server as `false`: for example, `job create` turned `refresh` off (#83).
- `update` sends only changed attributes. With no field flag it fails before any request (`nothing to update: set at least one field flag, see --help`) (#68, #83).
- Lookup errors now read `template "x" not found in organization <id>` and `multiple agents match "x": <id>, <id>; pass the ID instead`. Name lookups (for example `-o my-org`) require an exact name match (#83).
- `job create` / `job update`: the job log field flag is now `--job-output`. `--output` is always the global output format and is checked before any API call (#69).
- `job create`: `--command`/`-c` is replaced by `--template-reference`/`-t` (the template ID to run). The server never accepted `command`.
- Flags for attributes the API does not expose are removed: `webhook --path/--branch/--template-id/--event` (use `webhook-event`), `vcs --refresh-token/--token-expiration`, `address --type`, `agent --type`.
- `github-app-token --expires-at` takes an RFC 3339 timestamp (`2026-01-02T15:04:05Z`). Dates print as UTC RFC 3339 timestamps.
- A path in `api_url` is kept as a prefix: `https://host/terrakube` sends requests to `https://host/terrakube/api/v1/...`. Before, the path was dropped, so an `api_url` that ends in `/api/v1` must lose that suffix.
- Empty lists print `[]` in JSON instead of `null` (#69).
- `login` exits non-zero when it fails, and writes only `api_url` and `token`, with mode `0600`. The config file only provides `api_url` and `token`; other keys in it no longer fill flags (#67).
- Environment variables no longer fill field flags (`--name`, `--id`, `--description`, ...). Only parent-scope flags, `--output` and `--hide-nulls` are read from `TERRAKUBE_*` variables (#67).

### Fixed

- `module list` failed on current servers with `Trying to Unmarshal an object of type "module_version", but "version" does not match` (#64). The fix came with the upgrade to terrakube-go v1 (#77).
- Table output prints lists and maps instead of Go type names (#87).
- Names in RSQL lookups are quoted, so names with spaces or special characters resolve (#69).
- `login` and `logout` no longer leave a half-written or world-readable config file (#67).

## Earlier releases

`1.0.0-beta.1` to `1.0.0-beta.11` are described on the [GitHub releases page](https://github.com/terrakube-io/terrakube-cli/releases).

[Unreleased]: https://github.com/terrakube-io/terrakube-cli/compare/v1.0.0-rc.3...HEAD
[1.0.0-rc.3]: https://github.com/terrakube-io/terrakube-cli/compare/v1.0.0-rc.2...v1.0.0-rc.3
[1.0.0-rc.2]: https://github.com/terrakube-io/terrakube-cli/compare/v1.0.0-rc.1...v1.0.0-rc.2
[1.0.0-rc.1]: https://github.com/terrakube-io/terrakube-cli/compare/v1.0.0-beta.11...v1.0.0-rc.1
