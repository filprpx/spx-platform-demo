# Real platform vs. this demo

This document explains which parts of the SPX architecture are represented faithfully by this repository and which parts are intentionally simplified so the demo can be reproduced from a developer workstation.

## Real platform architecture

The intended platform has a remote execution plane. The control plane records developer intent, while a worker changes the IaC repository and opens a pull request. CI/CD then builds the application image, publishes it to ACR, applies the infrastructure, and deploys the application.

The revised diagram should be stored at:

```text
docs/images/real-platform.png
```

It should show:

```text
CLI
  → Microsoft Entra ID
  → Django control plane
  → Redis/Celery
  → remote worker
  → IaC repository and pull request
  → CI/CD pipeline
  → ACR and Azure Container Apps
```

The application source repository should feed the CI/CD image build. The worker should report status back to Django, and Django should not appear as the component that runs Terraform.

## This demo

The local demo keeps the same control-plane boundary but replaces remote integrations with local commands:

```text
CLI
  → Django API
  → Redis/Celery
  → host-side Python worker
  → generated local Terraform
  → make iac-pipeline
  → Azure infrastructure
  → make deploy-app
  → ACR and Container App
```

The revised demo diagram should be stored at:

```text
docs/images/demo-platform.png
```

It should make these boundaries visible:

- Django and Redis run in Docker Compose;
- the worker runs on the developer's host;
- the worker generates local Terraform but does not execute it;
- `make iac-pipeline` runs Terraform with the developer's Azure CLI session;
- `make deploy-app` builds the bundled `examples/simple-api`, pushes it to ACR, and updates the Container App;
- no pull request or remote CI/CD pipeline is created;
- `curl` verifies the real HTTP endpoint in Azure.

## Comparison

| Concern | Real platform | This demo |
|---|---|---|
| CLI/API | Go CLI and Django API | Go CLI and Django API |
| Authentication | Microsoft Entra | Microsoft Entra |
| Queue | Redis/Celery or managed equivalent | Local Redis/Celery |
| Worker | Remote execution service | Host-side Python worker |
| IaC source of truth | Git repository | Local generated IaC files |
| Change review | Pull request | Interactive Terraform approval |
| Terraform execution | CI/CD pipeline | `make iac-pipeline` |
| Image build | CI/CD pipeline | `make deploy-app` |
| Image registry | Shared/private ACR | Shared Basic ACR |
| Workload target | Azure Container Apps | Azure Container Apps |
| Terraform state | Remote managed state | Local non-production state |
| Credentials | Managed identity or service connection | Developer's Azure CLI session |
| Application source | Developer repository | Bundled `examples/simple-api` |
| Deployment status | Platform-visible deployment lifecycle | Local command output only |

## What remains faithful

The demo preserves the important control-plane and execution-plane ideas:

- authenticated CLI-to-Django communication;
- authenticated identity mapped to a Platform User;
- immutable provisioning job payloads;
- asynchronous queue processing;
- a separate worker boundary;
- generated Terraform as workload desired state;
- explicit approval before infrastructure apply;
- ACR-backed image deployment;
- managed identity for Container App image pulls;
- Azure Container Apps as the workload target;
- cleanup of workload and bootstrap resources through Terraform.

## What is simplified

The demo deliberately omits:

- remote Git providers;
- pull requests;
- CI/CD service connections;
- remote workers;
- remote Terraform state;
- production approval policy;
- team-level authorization;
- deployment status callbacks;
- arbitrary application source repositories;
- production networking and observability;
- application destruction as a separate user-facing workflow.

The demo's `deploy-app` command is therefore not pretending to be the final deployment architecture. It is a local stand-in for the build-and-deploy stage that a future CI/CD pipeline would perform.

## Command mapping

```text
Real platform request
  → CLI/API
  → queue
  → worker
  → pull request
  → CI/CD

Demo request
  → CLI/API
  → queue
  → worker
  → generated local Terraform
  → make iac-pipeline
  → make deploy-app
```

The durable architectural decisions are recorded in [ADR 0003](adr/0003-keep-provisioning-outside-the-django-control-plane.md), [ADR 0004](adr/0004-use-container-apps-for-demo-workloads.md), and [ADR 0005](adr/0005-use-shared-acr-and-separate-app-deployment.md).
