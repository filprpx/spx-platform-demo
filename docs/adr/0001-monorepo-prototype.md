# Use a monorepo for the prototype

The prototype keeps the CLI, Django API, and infrastructure bootstrap in one repository so the authenticated control-plane milestone can evolve together.

The repository has three independently bounded components:

- `cli/` contains the Go/Cobra developer CLI;
- `api/` contains the Django control-plane API and Admin surface;
- `infra/bootstrap/` contains Terraform for the Entra application bootstrap.

The components communicate through explicit contracts and remain independently replaceable. Moving them into separate repositories is deferred until those contracts and workflows stabilize.
