# Internal Developer Platform — Sliced Implementation Plan

This plan deliberately separates the control plane from the execution plane. Milestone 1 proves the authenticated Django control plane. Milestone 2 adds local infrastructure generation and Terraform execution without requiring Azure DevOps or GitHub CI/CD configuration. The worker generates Terraform; `iac-pipeline` executes it locally.

The prototype uses one monorepo with independent boundaries for the CLI, Django API, and IaC bootstrap. Those components can move into separate repositories after their contracts and workflows stabilize.

For the runnable tutorial, see the [README](../README.md). Installation and Azure access details are in [Prerequisites](prerequisites.md), and the production-versus-demo boundary is explained in [Real platform vs. this demo](real-platform-vs-demo.md).

## Milestone 2 — Local Container App infrastructure

Milestone 2 extends the completed authenticated control plane with a real local execution path:

```text
Go CLI
  → Django API
  → Redis job
  → Python worker
  → Generated application Terraform root
  → `make iac-pipeline`
  → Terraform using the local Azure CLI session
  → Azure
  → Django status visibility
```

The milestone does not require Azure DevOps, GitHub Actions, remote Terraform state, or a pipeline service connection. The local worker and `iac-pipeline` represent the future execution pipeline while keeping the IaC repository, Terraform files, and status boundaries real.

Milestone 2 is complete when a developer can create an application, observe a queued job, inspect generated Terraform, run `iac-pipeline`, deploy the bundled demo image with `deploy-app`, verify the Container App with `curl`, and see the infrastructure execution status in the CLI and Django Admin.

## Slice 0 — Define the contract and local boundaries

### Goal

Establish the interface between the CLI and Django before implementing provisioning.

### Build

- Django project with Django REST Framework.
- Go CLI project with Cobra.
- Initial API contract.
- Local API URL configuration.
- Request and response examples.
- Independent CLI, API, and IaC boundaries.
- Domain vocabulary for developers, platform users, applications, owning teams, and provisioning requests.

### Initial API

```text
POST /api/v1/applications
GET  /api/v1/applications
GET  /api/v1/applications/{id}
GET  /api/v1/provisioning/{id}
```

### Completed boundary

The Django API records platform intent and provisioning state. It does not execute Terraform or provision Azure resources. Use SQLite initially to minimize infrastructure friction.

## Slice 1 — Entra bootstrap

### Goal

Configure the real identity dependency required by the authenticated control plane.

### Build

- Entra API and CLI application registrations;
- API Application ID URI and delegated `access_as_user` scope;
- public-client configuration and fixed localhost callback;
- Terraform outputs consumed by the root environment and installed CLI configuration.

### Deferred

- application workload resources;
- pipeline execution and provisioning credentials.

## Slice 2 — CLI authentication

### Goal

Authenticate the developer through Microsoft Entra and establish the CLI-to-API flow.

Example commands:

```bash
spx login
spx whoami
```

The CLI uses browser-based authorization-code flow with PKCE and stores the OAuth token in the operating system credential store. Its non-secret configuration is generated from Terraform outputs and works from the repository or outside it.

## Slice 3 — API identity mapping

### Goal

Validate Entra access tokens and map the authenticated identity to a local `PlatformUser`.

The API validates token signature, issuer, tenant, audience, expiration, and delegated scope. Invalid or unauthenticated requests return `401`.

Complex authorization remains deferred; any valid authenticated user can use the demo.

## Slice 4 — Application intent capture

### Goal

Record an application infrastructure request without collecting application source code.

Example commands:

```bash
spx app create
spx app create --name payments-api --owning-team finance-engineering --compute-size small --container-port 8080 --ingress external --min-replicas 0 --max-replicas 1
spx app list
spx app describe payments-api
```

`spx app create` opens an interactive wizard grouped into application identity, compute, network, scaling, and review steps. The wizard submits the deployment configuration directly; the old `type` and `runtime` fields are not part of the CLI workflow.

The Django API should:

- derive the creator from the bearer token;
- validate and normalize application input;
- assign exactly one owning team and creator;
- create exactly one pending `ProvisioningRequest` in the same transaction;
- return the application and provisioning-request identifier.

The CLI should:

- send HTTP requests;
- display validation errors;
- display the created application;
- handle API connection failures;
- read the backend URL from configuration;
- render validation and API connection errors.

No Terraform or provisioning adapter is called.

## Slice 5 — Admin visibility

### Goal

Make recorded platform intent and pending provisioning requests visible to operators.

Register `PlatformUser`, `Application`, and `ProvisioningRequest` in Django Admin. The Admin surface observes and manages recorded intent; it is not a second provisioning mechanism.

## Slice 6 — Local asynchronous IaC execution

### Goal

Execute the real workload Terraform locally while preserving the boundary that a future remote pipeline will use.

### Build

- Celery and Redis;
- Docker Compose Redis service;
- a Python worker running on the developer's workstation;
- immutable job payloads containing the application and desired-state snapshot;
- an internal Django endpoint for execution status events;
- direct writes to the checked-out IaC repository for the single-user demo;
- explicit application Terraform roots under the IaC repository;
- local Terraform state stored outside the Git-managed IaC tree;
- a background host-side worker managed by `make setup`;
- Terraform execution through `make iac-pipeline` using the developer's existing Azure CLI session;
- a shared Basic-tier ACR created by bootstrap Terraform;
- Container Apps workload Terraform with scale-to-zero defaults;
- a separate `make deploy-app` command for the bundled simple API;
- retry, timeout, idempotency, and failure handling.

The worker must not access the Django database directly. It receives all execution inputs through the queue and reports status through the internal API.

The worker generates desired state and stops at `WAITING_FOR_APPROVAL`. `make iac-pipeline` discovers the pending workload without user-supplied identifiers, runs Terraform formatting, initialization, validation, and plan, then asks for confirmation before apply. This command represents the approved pull-request pipeline in the real platform.

The IaC repository should be organized like:

```text
infra/
├── bootstrap/
└── workloads/
    └── dev/
        └── teams/
            └── finance-engineering/
                └── payments-api/
                    ├── main.tf
                    ├── variables.tf
                    └── versions.tf

examples/
└── simple-api/
```

The worker creates only the application-specific Terraform files. The local demo deliberately writes to the checked-out repository instead of using worktrees; worktrees and remote branches remain future concurrency and pull-request concerns.

Flow:

```text
CLI request
  → Django transaction
  → immutable job snapshot
  → Redis
  → background local Python worker
  → Terraform files in the checked-out IaC repository
  → `iac-pipeline` representing approved PR execution
  → Terraform plan/apply using az login
  → internal status events
  → ProvisioningRequest updated
```

The worker creates the application directory. The local demo does not create branches or commits. `iac-pipeline` represents the future approved pipeline and runs Terraform from the generated directory. After infrastructure succeeds, `make deploy-app` builds and pushes the checked-in `examples/simple-api` image and updates the Container App. Application image deployment is deliberately separate from Django provisioning. Application deletion is handled by workload Terraform during teardown.

### Local execution boundary

The API and Redis run in Docker Compose. The worker runs in the background on the host so it can use the host's `az` and `terraform` installations and the developer's Azure CLI login. `make teardown` stops the project worker before removing local state. The worker does not receive or mount the Django database.

### Deferred

- Azure DevOps or GitHub pull-request integration;
- remote pipeline execution;
- pipeline service connections and federated workload identities;
- remote Terraform state.

## Slice 7 — Remote pull-request integration (optional)

### Goal

Replace local Git branch handling with a real pull-request provider after local execution is reliable.

The first supported provider may be Azure DevOps or GitHub. Configuration includes:

- organization or account;
- project or repository;
- branch and pull-request permissions;
- application identity or integration credentials;
- repository-side workflow configuration.

The API should use a provider adapter. Provider calls must not be scattered through business logic.

Flow:

```text
Django
  → Celery
  → remote Git provider adapter
  → branch + Terraform files commit
  → pull request
```

The local worker's generated Terraform roots and commit contract remain unchanged.

### Deferred

- Remote pipeline execution.
- Azure workload provisioning credentials.

## Slice 8 — Remote pipeline execution (optional)

### Goal

Run the same workload Terraform through Azure DevOps or GitHub after a pull request is approved.

### Build

- Pipeline YAML for the selected provider.
- Terraform installation.
- `terraform fmt`.
- `terraform init`.
- `terraform validate`.
- `terraform plan` and `terraform apply`.
- Pull-request and pipeline status reporting.

Flow:

```text
PR created
  → pipeline triggered
  → approval
  → Terraform executed
  → pipeline result visible
  → worker observes result
```

### Deferred

- Production state storage.
- Managed identity or service connection hardening.
- Production approval policy.

## Slice 9 — Production-grade Terraform execution

### Goal

Harden the execution plane after local and remote execution are stable.

### Required Azure setup

- Azure subscription.
- Resource group for platform infrastructure.
- Terraform state storage.
- State locking.
- Azure DevOps or GitHub service connection.
- Workload identity or service principal.
- Least-privilege role assignment.
- Deployment region.
- Naming and tagging policy.
- Approval policy for production.

The prototype will start with one simple resource, such as a resource group, and then move to the first application Golden Path:

```text
Application
  → Azure Container App
  → Shared Azure Container Registry
  → Managed Identity
  → Log Analytics
  → Application Insights
```

The first real provisioning flow should target `dev` only.

Production remains deferred until state management, approval behavior, rollback/failure behavior, ownership, and audit records are proven.

## Terraform state note

For the prototype, a local Terraform backend may be used to avoid early Azure configuration. Terraform state must not be committed to Git and this mode must be clearly marked as non-production.

The production migration target is an Azure Blob backend with locking and pipeline-only access.

## Implementation order

```text
1. Context and independent boundaries
2. Entra bootstrap
3. CLI authentication
4. API identity mapping
5. Application intent capture
6. Django Admin visibility
7. Celery + Redis with a local Terraform worker
8. Remote pull-request integration, if needed
9. Remote pipeline execution, if needed
10. Production-grade Terraform execution
```

Every slice should leave behind a working demo. Remote Azure DevOps and GitHub configuration is isolated to later slices; the local execution milestone uses only the developer's existing Azure CLI login and the Terraform tooling already required by the bootstrap.
