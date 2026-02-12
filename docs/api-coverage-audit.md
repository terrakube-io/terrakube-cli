# Terrakube API Audit: Server vs CLI vs Terraform Provider

## Context

Three Go/Java projects interact with the Terrakube API, each with independent client implementations:

| Project | Tech | Resources | Client Location |
|---------|------|-----------|-----------------|
| **Server** | Spring Boot + Elide (JSON:API) | 24+ resource types, ~160 endpoints | Authoritative source |
| **CLI** | Go + Cobra | 6 resource types, ~22 operations | `client/client/`, `client/models/` |
| **TF Provider** | Go + TF Plugin Framework | 21 resources, 8 data sources | `internal/client/client.go` |

**Sources:**
- Server OpenAPI spec: `../terrakube/openapi-spec/v2_27_0.yml`
- CLI: `github.com/terrakube-io/terrakube-cli`
- TF Provider: `github.com/terrakube-io/terraform-provider-terrakube`

All three use `application/vnd.api+json` (JSON:API). The server uses Elide to auto-expose JPA entities. The CLI uses a hand-rolled JSON:API client. The provider uses `github.com/google/jsonapi` for marshaling and raw `net/http` for requests.

---

## 1. Three-Way Resource Comparison Matrix

Legend: **C**reate **R**ead **U**pdate **D**elete **L**ist **F**ilter **DS** = TF data source

| Resource | Server | CLI | TF Provider | Gap |
|----------|--------|-----|-------------|-----|
| **Organization** | CRUD+L | CUD+L | CRUD+L+F + DS | CLI missing Read (get by ID) |
| **Workspace** | CRUD+L (under org) | CUD+L | CRUD+L+F + DS (cli+vcs variants) | CLI missing Read |
| **Module** | CRUD+L (under org) | CUD+L | CRUD | CLI missing Read |
| **Module Version** | CRUD+L (under module) | - | - | Neither client supports |
| **Job** | CRUD+L (under org) | C+L (U/D broken) | - | CLI has bugs; provider skips jobs |
| **Job Step** | CRUD+L (under job) | - | - | No client support |
| **Job Address** | CRUD+L (under job) | - | - | No client support |
| **Team** | CRUD+L (under org) | CUD+L | CRUD + DS | CLI missing Read |
| **Team Token** | Custom REST (non-JSON:API) | - | CD | CLI missing entirely |
| **Variable** (workspace) | CRUD+L (under ws) | CUD+L | CRUD | CLI missing Read |
| **Global Variable** (org) | CRUD+L (under org) | - | CRUD | CLI missing entirely |
| **Template** | CRUD+L (under org) | - | CRUD+F + DS | CLI missing entirely |
| **VCS** | CRUD+L (under org) | - | CRUD + DS | CLI missing entirely |
| **SSH** | CRUD+L (under org) | - | CRUD+F + DS | CLI missing entirely |
| **Agent** | CRUD+L (under org) | - | CRUD | CLI missing entirely |
| **Schedule** (workspace) | CRUD+L (under ws) | - | CRUD | CLI missing entirely |
| **Workspace Access** | CRUD+L (under ws) | - | CRUD | CLI missing entirely |
| **Workspace Tag** | CRUD+L (under ws) | - | CRUD | CLI missing entirely |
| **Org Tag** | CRUD+L (under org) | - | CRUD+F + DS | CLI missing entirely |
| **Webhook** | CRUD+L | - | CRUD (v2 + events) | CLI missing entirely |
| **Collection** | CRUD+L (under org) | - | CRUD | CLI missing entirely |
| **Collection Item** | CRUD+L (under collection) | - | CRUD+F | CLI missing entirely |
| **Collection Reference** | CRUD+L | - | CRUD | CLI missing entirely |
| **Provider** | CRUD+L (under org) | - | - | Neither client supports |
| **Provider Version** | CRUD+L (under provider) | - | - | Neither client supports |
| **Provider Implementation** | CRUD+L (under version) | - | - | Neither client supports |
| **History** (workspace) | RUD+L (under ws) | - | - | Neither client supports |
| **Action** | CRUD+L | - | - | Neither client supports |
| **Address** | CRUD+L | - | - | Neither client supports |
| **Step** (top-level) | CRUD+L | - | - | Neither client supports |
| **GitHub App Token** | CRUD+L | - | - | Neither client supports |
| **TF Output** | Custom REST | - | DS only | CLI missing; provider read-only |
| **TF State** | Custom REST | - | - | Neither client supports |
| **Log Streaming** | WebSocket | - | - | Neither client supports |
| **PAT (Personal Token)** | Custom REST | - | - | Neither client supports |
| **TF Cloud Import** | Custom REST | - | - | Neither client supports |

**Totals:**

| | Server | CLI | TF Provider |
|--|--------|-----|-------------|
| Resources with any support | 34 | 6 | 21 |
| Full CRUD | 28+ | 4 (org, workspace, module, team) | 20 |
| Data sources (read-only) | N/A | N/A | 8 |

---

## 2. Client Library Comparison

| Aspect | CLI | TF Provider |
|--------|-----|-------------|
| JSON:API library | Hand-rolled (`splitInterface`/`mergeInterface`) | `github.com/google/jsonapi` v1.0.0 |
| HTTP client | Hand-rolled (`do()` method) | Raw `net/http` per resource |
| Auth | Bearer token (from config file) | Bearer token (`TERRAKUBE_TOKEN` env) |
| Error handling | `do()` never checks HTTP status codes | Per-resource, checks status |
| Filtering | `newRequestWithFilter` (broken in 2 resources) | Query params `filter[type]=field==value` |
| Content type | `application/vnd.api+json` | `application/vnd.api+json` |
| Model definition | `client/models/` structs with json tags | `internal/client/client.go` with jsonapi tags |
| Path construction | `fmt.Sprintf` per method | `fmt.Sprintf` per resource file |
| Org soft delete | Hard DELETE | PATCH with `Disabled=true` |
| Bool handling | `omitempty` drops false (bug) | `omitempty` drops false (same bug) |

Both clients share the same `omitempty` bug on booleans -- false values get dropped from JSON serialization. Both use `fmt.Sprintf` for path construction with no shared abstraction.

---

## 3. Detailed Per-Resource Analysis

### Resources in CLI but NOT in TF Provider

| Resource | Why it matters |
|----------|---------------|
| **Job** | Jobs are imperative (run now). TF provider is declarative, so jobs don't fit the TF model. CLI is the right home for job management. |

### Resources in TF Provider but NOT in CLI

| Resource | CLI value | Notes |
|----------|-----------|-------|
| Template | High | Need to manage workflow templates from CLI |
| VCS | High | Need to configure VCS connections |
| SSH | Medium | Manage SSH keys for private repos |
| Agent | High | Manage self-hosted agents |
| Global Variable | Medium | Org-wide variable management |
| Schedule | High | Manage scheduled runs |
| Workspace Access | Medium | Access control management |
| Workspace Tag | Low | Tag workspaces |
| Org Tag | Low | Tag organizations |
| Webhook | Medium | Configure webhook integrations |
| Collection/Item/Reference | Low | Variable collection management |
| Team Token | Medium | Generate/revoke team API tokens |

### Resources in NEITHER client

| Resource | Server entity | Likely reason |
|----------|---------------|---------------|
| Provider registry | Provider, ProviderVersion, Implementation | Complex nested hierarchy, possibly internal |
| Job Steps/Addresses | Step, Address | Read-only execution details, internal |
| History | History | State version browsing, could be useful |
| Action | Action | Custom actions, possibly internal |
| GitHub App Token | GithubAppToken | Auth infrastructure |
| PAT | PersonalAccessToken | Custom controller, not Elide |
| TF State/Output | Custom controllers | Separate API pattern |
| Log Streaming | WebSocket | Different protocol entirely |

---

## 4. Shared Library Opportunity

### Current State: Duplicate Everything

```
terrakube-cli/client/          terraform-provider-terrakube/internal/client/
  client/                        client.go
    organization.go                - OrganizationEntity
    workspace.go                   - WorkspaceEntity
    module.go                      - ModuleEntity
    job.go                         - (no equivalent)
    team.go                        - TeamEntity
    variable.go                    - WorkspaceVariableEntity
  models/                          - ... 14 more entities
    organization.go                  (all in one file)
    workspace.go
    module.go
    job.go
    team.go
    variable.go
```

Both projects independently:
- Define Go structs for the same API resources
- Implement JSON:API serialization/deserialization
- Build URL paths with `fmt.Sprintf`
- Handle bearer token auth
- Make HTTP requests with `net/http`

### Proposed: Shared `terrakube-go` Client Library

```
github.com/terrakube-io/terrakube-go/
  models/           # JSON:API model structs (shared)
  client/           # HTTP client with CRUD methods (shared)
  jsonapi/          # JSON:API marshaling (use google/jsonapi)
```

**Benefits:**
- Single source of truth for models and API paths
- Fix bugs once (omitempty, status code checking, filter params)
- New resources added once, available to both CLI and provider
- Provider already uses `google/jsonapi` which is production-ready

**Risks:**
- Versioning: CLI and provider may need different client versions
- Breaking changes in shared lib affect both consumers
- Migration effort for both projects

**Migration path:**
1. Extract models from TF provider (it has the most coverage and uses `google/jsonapi`)
2. Add client methods with proper error handling
3. Update CLI to consume shared lib (replacing hand-rolled JSON:API)
4. Update TF provider to consume shared lib (replacing per-resource HTTP)
5. Add missing resources (Job, Provider registry, etc.) to shared lib

---

## 5. Prioritized Roadmap

### Phase 0: Fix CLI Bugs (prerequisite)
- [ ] Fix `do()` to check HTTP status codes (affects all operations)
- [ ] Fix Job.Delete path (`/module/` -> `/job/`)
- [ ] Fix Job.Update to accept `models.Job` and use `/job/` path
- [ ] Fix Job.List to use `newRequestWithFilter`
- [ ] Fix Variable.List to use `newRequestWithFilter`
- [ ] Fix Module update cmd registering wrong subcommand
- [ ] Fix Team boolean `omitempty` dropping false values
- [ ] Fix Variable boolean `omitempty` dropping false values
- [ ] Fix Workspace update not setting ID on model
- [ ] Add empty string ID validation
- [ ] Add "Get by ID" for all 6 existing resources

### Phase 1: Shared Library Foundation
- [ ] Create `terrakube-go` repo with models extracted from TF provider
- [ ] Adopt `google/jsonapi` for serialization (replacing hand-rolled code)
- [ ] Implement generic CRUD client with proper error handling
- [ ] Add filtering, pagination support
- [ ] Migrate CLI to use shared library
- [ ] Migrate TF provider to use shared library

### Phase 2: High-Priority CLI Resources (match TF provider)
- [ ] Agent (CRUD) - manage self-hosted execution agents
- [ ] Template (CRUD + filter) - manage workflow templates
- [ ] VCS (CRUD) - manage version control connections
- [ ] Schedule (CRUD under workspace) - manage recurring runs
- [ ] Team Token (Create/Delete) - generate/revoke tokens
- [ ] Job Update/Delete - fix and expose via CLI commands
- [ ] Job Steps (Read/List) - view execution step details

### Phase 3: Medium-Priority CLI Resources
- [ ] SSH key management (CRUD)
- [ ] Global Variables (CRUD under org)
- [ ] Workspace Access control (CRUD)
- [ ] Webhook management (CRUD + events)
- [ ] Module Versions (CRUD under module)
- [ ] Workspace History (Read/List)
- [ ] TF Output retrieval (data read)

### Phase 4: Full Parity Resources
- [ ] Collection / Collection Item / Collection Reference (CRUD)
- [ ] Org Tags + Workspace Tags (CRUD)
- [ ] Provider registry (CRUD + versions + implementations)
- [ ] TF State operations (push/pull)
- [ ] Log streaming (job output, WebSocket)
- [ ] PAT management (custom REST)

### Phase 5: Advanced / Niche
- [ ] TF Cloud importer
- [ ] State migration
- [ ] GitHub App Token management
- [ ] Action management
- [ ] Address/Step top-level management

---

## 6. Coverage Metrics

### Current State

| Metric | CLI | TF Provider |
|--------|-----|-------------|
| Resources supported | 6 of 34 (18%) | 21 of 34 (62%) |
| Operations working correctly | ~20 of 22 (91%) | ~84 of 84 (100%) |
| Shared code between them | 0% | 0% |

### After Phase 2 (target)

| Metric | CLI | TF Provider |
|--------|-----|-------------|
| Resources supported | 13 of 34 (38%) | 21 of 34 (62%) |
| Shared library coverage | 100% of CLI resources | 100% of provider resources |
| Known bugs | 0 | 0 (omitempty fixed in shared lib) |

### After Phase 4 (target)

| Metric | CLI | TF Provider |
|--------|-----|-------------|
| Resources supported | 26 of 34 (76%) | 26 of 34 (76%) |
| Feature parity | Full overlap except Job (CLI-only) | Full overlap |
