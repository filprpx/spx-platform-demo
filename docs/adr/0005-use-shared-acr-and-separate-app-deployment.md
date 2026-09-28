# Use a shared bootstrap registry and separate application deployment

The bootstrap Terraform creates one shared Basic-tier Azure Container Registry. It is created empty and reused by demo workloads rather than creating a registry per application.

Workload Terraform gives each Container App a managed identity and grants it `AcrPull` access to the shared registry. Registry administrator credentials are not stored in the application configuration or job payload.

Infrastructure and application deployment are intentionally separate:

```text
make iac-pipeline
  → resource group, Container Apps Environment, Container App, identity, registry access

make deploy-app
  → build examples/simple-api
  → push image to ACR
  → update the Container App
```

The Django API and worker own infrastructure intent and desired-state generation. `deploy-app` is a local demonstration command that uses Docker and the developer's Azure CLI session. It does not publish a Django job or update provisioning status.

The shared ACR is removed by the bootstrap destroy step. Workload resources are destroyed before local Terraform state is removed so the demo does not orphan Azure resources.
