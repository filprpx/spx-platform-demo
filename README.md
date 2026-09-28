# SPX Internal Developer Platform Demo

This repository demonstrates an authenticated internal developer platform that captures infrastructure intent, provisions an Azure Container App with Terraform, and deploys a small demo API.

It is a reproducible teaching demo, not a production platform. See [Real platform vs. this demo](docs/real-platform-vs-demo.md) for the architectural differences.

## Time, resources, and cost expectations

The first provisioning run may take around 15 minutes, and teardown may take around 25 minutes. These are observed timings, not guarantees. The demo creates real Azure resources and may incur charges; see [Prerequisites](docs/prerequisites.md) and [Real platform vs. this demo](docs/real-platform-vs-demo.md).

Run the complete cleanup when finished:

```bash
make teardown
```

## Requirements

You need:

- an Azure subscription and an authenticated Azure CLI;
- Docker and Docker Compose;
- Terraform;
- Go;
- Python with virtual-environment support;
- OpenSSL.

See [Prerequisites](docs/prerequisites.md) for installation, Azure permissions, ports, costs, and troubleshooting.

## Run the demo

### 1. Clone the repository

```bash
git clone https://github.com/filprpx/spx-platform-demo.git
cd spx-platform-demo
```

### 2. Authenticate Azure

```bash
az login
```

Select the subscription that should receive the demo resources if your account has access to more than one:

```bash
az account set --subscription "<subscription-id>"
```

### 3. Set up the local control plane

```bash
make setup
```

This verifies dependencies, interactively applies the bootstrap Terraform, creates the Entra registrations and shared Basic ACR, generates local configuration, starts Django/Redis, starts the host-side worker, and installs `spx` under `~/.local/bin/spx`.

Create an administrator for Django Admin:

```bash
make admin
```

### 4. Authenticate the CLI

```bash
spx login
spx whoami
```

The CLI uses browser-based Microsoft Entra authentication. Tokens are stored in the operating-system credential store.

### 5. Declare infrastructure intent

Run the interactive wizard:

```bash
spx app create
```

The wizard asks for:

- application name;
- owning team;
- compute size;
- container port;
- network visibility;
- minimum replicas;
- maximum replicas.

These choices describe how the platform should run the application. The wizard does not ask for a source repository, Git commit, Dockerfile, image, or Azure credential.

The request can be inspected with:

```bash
spx app list
spx app describe <application-name>
```

### 6. Provision infrastructure

```bash
make iac-pipeline
```

The worker has already generated Terraform under `infra/workloads/`. This command formats, initializes, validates, and plans the workload, then asks for confirmation before applying it with the developer's Azure CLI session.

It creates the resource group, Container Apps Environment, Log Analytics workspace, managed identity, ACR pull permission, and a dormant Container App using a public placeholder image.

### 7. Build and deploy the demo API

```bash
make deploy-app
```

This builds the checked-in application under `examples/simple-api`, pushes it to the shared ACR, and updates the Container App. It ends by printing a copyable validation command:

```text
To test the API, run:

  curl https://<container-app-url>
```

The expected response is:

```json
{"application":"SPX demo API","status":"ok"}
```

The two commands have separate responsibilities:

```text
make iac-pipeline
  Infrastructure Terraform and Azure resources.

make deploy-app
  Docker image build, ACR push, and Container App update.
```

### 8. Clean up

When finished, remove the workload resources, bootstrap resources, local services, credentials, and generated state:

```bash
make teardown
```

Do not use `make clean-local` as a replacement after applying Azure infrastructure. It removes local workload state without contacting Azure. Use it only for local cleanup when no Azure resources need to be destroyed.

## Useful commands

```bash
make doctor          # Check local tools, Docker, and Azure login
make worker-status   # Check the background worker
make worker-logs     # Follow worker output
make logs            # Follow Django API logs
make down            # Stop Docker Compose services
make up              # Start Docker Compose services again
make test            # Run project tests
spx logout           # Clear the CLI token without removing configuration
```

## Project boundaries

```text
cli/              Go CLI
api/              Django API and host-side Celery worker
infra/bootstrap/  Entra and shared ACR bootstrap Terraform
infra/workloads/  Generated per-application Terraform
examples/         Bundled application used by make deploy-app
docs/             Architecture decisions and implementation guidance
```

The production architecture would add a Git provider, pull requests, remote workers, remote Terraform state, and CI/CD. Those pieces are intentionally deferred. See [Real platform vs. this demo](docs/real-platform-vs-demo.md).
