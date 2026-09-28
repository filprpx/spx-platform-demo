# Demo prerequisites

This page contains the setup details behind the short requirements list in the [README](../README.md). The project checks for dependencies, but it does not install them automatically.

## Local tools

Install these tools using their official documentation:

- [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli) — authenticates the developer and provides the Azure commands used by bootstrap and deployment.
- [Terraform](https://developer.hashicorp.com/terraform/install) — initializes, plans, and applies the bootstrap and workload infrastructure.
- [Docker Engine or Docker Desktop](https://docs.docker.com/engine/install/) — runs Django and Redis and builds the demo API image.
- [Docker Compose](https://docs.docker.com/compose/install/) — starts the API and Redis services together.
- [Go](https://go.dev/doc/install) — builds and tests the `spx` CLI.
- [Python 3](https://www.python.org/downloads/) — runs the host-side Celery worker.
- [Python virtual environments](https://docs.python.org/3/library/venv.html) — creates the worker environment under `api/.venv`.
- OpenSSL — generates local Django and worker secrets. On Linux and macOS it is normally available through the operating system package manager.

Verify the project environment with:

```bash
make doctor
```

The check also verifies that the Docker daemon is reachable, Docker Compose is available, Azure CLI is authenticated, and `~/.local/bin` is visible in `PATH`.

## Azure access

You need:

- an Azure subscription;
- access to an Entra tenant;
- permission to create Entra app registrations and service principals;
- permission to create resource groups;
- permission to create an Azure Container Registry;
- permission to assign the `AcrPull` role;
- permission to create Container Apps, Container Apps Environments, and Log Analytics resources.

Authenticate and select the subscription before running setup:

```bash
az login
az account show
az account set --subscription "<subscription-id>"
```

The bootstrap uses the active Azure CLI session. Tenant policies may prevent app registration, service-principal creation, or administrator consent. If that happens, ask a tenant administrator to provide the required permission or record the prerequisite before continuing.

The demo creates an empty shared Basic-tier ACR during bootstrap. Workload Terraform creates a Container App with a managed identity and grants that identity permission to pull from the registry. See [Container Apps image pull with managed identity](https://learn.microsoft.com/en-us/azure/container-apps/managed-identity-image-pull).

## Local services and ports

| Service | Address | Purpose |
|---|---|---|
| Django API | `http://localhost:8000` | Authenticated control-plane API and Admin |
| Django Admin | `http://localhost:8000/admin/` | Inspect users, applications, and provisioning requests |
| Redis | `localhost:6379` | Celery broker |
| OAuth callback | `http://localhost:8765/callback` | Browser login callback for the CLI |

Django and Redis run in Docker Compose. The Celery worker runs on the host so it can write the local IaC files and use the host's Terraform and Azure CLI installations.

## Authentication storage

The `spx` CLI stores authentication tokens in the operating-system credential store rather than plaintext files:

- macOS Keychain;
- Windows Credential Manager;
- Linux Secret Service.

Linux users need a running Secret Service provider such as GNOME Keyring or another compatible desktop credential store. `spx login` reports an actionable error if no credential store is available. It does not fall back to plaintext token storage.

The generated CLI configuration is separate from the token and contains only non-secret `PLATFORM_*` settings.

## Cost and cleanup

The demo is designed for short-lived experiments, but Azure resources are not automatically free:

- the bootstrap creates a shared Basic ACR;
- Container Apps use the Consumption plan and default to scale to zero;
- ACR, Log Analytics, storage, requests, and active Container App replicas can still incur charges;
- Azure resources continue to exist until Terraform destroys them.

After testing, run:

```bash
make teardown
```

`make teardown` destroys applied workload Terraform before destroying bootstrap resources and removing local state. `make clean-local` only removes local files and Docker resources; it does not destroy Azure resources.

See the [Azure Container Apps pricing](https://azure.microsoft.com/en-us/pricing/details/container-apps/) and [Azure Container Registry pricing](https://azure.microsoft.com/en-us/pricing/details/container-registry/) pages for current pricing.

## Common setup failures

### Azure CLI is not authenticated

Run:

```bash
az login
az account show
```

If the wrong subscription is selected, run `az account set --subscription "<subscription-id>"` and retry.

### Docker daemon is unavailable

Start Docker Desktop or the Docker daemon. On Linux, make sure the current user can access the Docker socket, then verify:

```bash
docker info
docker compose version
```

### Python virtual environment support is missing

Install the operating-system package that provides `python3 -m venv`, then retry `make doctor`.

### Terraform outputs are missing

Run the bootstrap apply before generating configuration:

```bash
make bootstrap-init
make bootstrap-apply
make env
```

### Entra permissions are blocked

The bootstrap creates app registrations and service principals. A tenant policy or insufficient directory role can block those resources. Resolve the permission issue with the tenant administrator rather than replacing real Entra authentication with a mock.

### Container Apps Environment creation is slow

The first Container Apps Environment can take several minutes while Azure provisions the managed environment and Log Analytics integration. Keep the Terraform process running unless it reports an error. Check the resource with:

```bash
az containerapp env show \
  --name "<environment-name>" \
  --resource-group "<resource-group-name>" \
  --query properties.provisioningState \
  --output tsv
```

### ACR role assignment fails

The workload requires permission to assign `AcrPull` to the Container App identity. Confirm that the Azure account can create role assignments at the registry scope.

### The worker is not running

Check and restart it with:

```bash
make worker-status
make worker-logs
make worker-start
```

### No workload is waiting or applied

The worker must consume the provisioning job before the IaC pipeline can run. Check the worker log, then inspect the generated local state under `.local/spx/`.

Run `spx app describe <application-name>` to inspect the provisioning status returned by Django.
