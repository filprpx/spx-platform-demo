# Keep provisioning execution outside the Django control plane

The Django application is the control plane. It records application intent, owning-team responsibility, creator attribution, provisioning-request records, and status visibility. Creating an application creates one pending provisioning request in the same transaction.

Django does not create Terraform files or execute Terraform. After the transaction commits, it publishes an immutable application snapshot to a job queue. A separate local worker consumes the job, validates the snapshot, generates the application Terraform files, and reports status events back to Django through an internal API. A separate `iac-pipeline` command runs Terraform using the developer's Azure CLI session.

The IaC repository is the source of truth for desired workload infrastructure. Each application has an explicit Terraform root under the workload tree, for example:

```text
infra/workloads/dev/teams/finance-engineering/payments-api/
├── main.tf
├── variables.tf
└── versions.tf
```

The first execution adapter runs locally. The pipeline uses the developer's existing Azure CLI session and local Terraform installation, without placing Azure credentials in the job payload or giving the worker direct database access. Terraform state remains separate from the Git-managed configuration and is never committed.

Future execution may use pull requests and independently running automation through Azure DevOps or GitHub. Those integrations replace the local execution adapter; they do not change the Django domain model or the IaC repository contract.

Application image deployment is a separate local `deploy-app` workflow. It builds the checked-in demo image, pushes it to ACR, and updates the already-provisioned Container App. It is deliberately not part of the Django provisioning job.

This boundary keeps domain state and user-facing orchestration in the API while leaving file generation, Git operations, Terraform, Azure execution, and application image deployment outside Django business logic.
