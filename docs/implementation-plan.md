# Internal Developer Platform — Sliced Implementation Plan

This plan deliberately separates the control plane from the execution plane and postpones workload provisioning until the authenticated local behavior is proven.

The prototype uses one monorepo with independent boundaries for the CLI, Django API, and IaC bootstrap. Those components can move into separate repositories after their contracts and workflows stabilize.

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
platform login
platform whoami
```

The CLI uses browser-based authorization-code flow with PKCE and stores the OAuth token in the operating system credential store. Its non-secret configuration is generated from Terraform outputs and works from the repository or outside it.

## Slice 3 — API identity mapping

### Goal

Validate Entra access tokens and map the authenticated identity to a local `PlatformUser`.

The API validates token signature, issuer, tenant, audience, expiration, and delegated scope. Invalid or unauthenticated requests return `401`.

Complex authorization remains deferred; any valid authenticated user can use the demo.

## Slice 4 — Application intent capture

### Goal

Record an application request without provisioning infrastructure.

Example commands:

```bash
platform app create payments-api --type api --runtime go --owning-team finance-engineering
platform app list
platform app describe payments-api
```

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

## Slice 6 — Celery orchestration with a fake PR provider

### Goal

Introduce asynchronous execution without Azure dependencies.

### Build

- Celery.
- Redis.
- Docker Compose.
- Background provisioning task.
- Retry handling.
- Provider interface for Git and pull-request operations.

Define an internal provider interface with operations equivalent to:

```text
create_branch()
write_manifest()
create_pull_request()
get_pull_request_status()
get_pipeline_status()
```

Initially implement a fake provider that:

- creates a local simulated pull-request record;
- returns a fake pull-request URL;
- transitions through predictable states.

Flow:

```text
API request
  → Celery task
  → fake PR provider
  → ProvisioningRequest updated
```

This slice validates task dispatch, retries, status transitions, polling, failure handling, and idempotency.

### Deferred

- Real Azure DevOps credentials.
- Real repositories.
- Real pipeline execution.

## Slice 7 — Real Azure DevOps pull-request integration

### Goal

Replace only the fake PR provider.

Configure Azure DevOps for:

- organization;
- project;
- repository;
- branch permissions;
- pull-request permissions;
- development credentials;
- pipeline repository access.

The API should use an Azure DevOps adapter. Azure DevOps calls must not be scattered through business logic.

Flow:

```text
Django
  → Celery
  → Azure DevOps adapter
  → branch + manifest commit
  → pull request
```

The API remains private and polls Azure DevOps from the worker.

At this point, the pull request may contain only a simple manifest file. Terraform does not need to run yet.

### Deferred

- Azure subscription provisioning.
- Terraform apply.
- Azure resource creation.

## Slice 8 — Azure DevOps pipeline with Terraform validation only

### Goal

Prove that a pull request can trigger infrastructure automation without provisioning Azure resources.

### Build

- Azure DevOps pipeline YAML.
- Terraform installation.
- `terraform fmt`.
- `terraform init`.
- `terraform validate`.
- Basic plan validation.
- Pull-request status reporting.

Initially use a harmless local/demo Terraform module or configuration that validates structure without creating Azure resources.

Flow:

```text
PR created
  → pipeline triggered
  → Terraform validated
  → pipeline result visible
  → backend polling observes result
```

### Deferred

- AzureRM resource creation.
- Production state storage.
- Managed identity/service connection.
- Terraform apply.

## Slice 9 — Terraform provisions a real Azure resource

### Goal

Connect the execution plane to Azure only after the control plane and pipeline are stable.

### Required Azure setup

- Azure subscription.
- Resource group for platform infrastructure.
- Terraform state storage.
- State locking.
- Azure DevOps service connection.
- Workload identity or service principal.
- Least-privilege role assignment.
- Deployment region.
- Naming and tagging policy.
- Approval policy for production.

The prototype will start with one simple resource, such as a resource group, and then move to the first application Golden Path:

```text
Application
  → Azure Container App
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
7. Celery + Redis with a fake pull-request provider
8. Real Azure DevOps pull-request integration
9. Terraform validation pipeline
10. Terraform Azure provisioning
```

Every slice should leave behind a working demo. Azure configuration is intentionally isolated to later slices rather than being a prerequisite for proving the CLI, API, state model, or orchestration design.
