# SPX Internal Developer Platform — Milestone 1

This milestone implements the authenticated control plane only:

```text
Go CLI → Microsoft Entra ID → Django REST API → SQLite → Django Admin
```

No Azure workload, Terraform execution, Celery, Redis, PostgreSQL, or Azure DevOps integration is included yet.

## Local setup

Required tools:

- Azure CLI (`az`), already authenticated with `az login`;
- Terraform;
- Docker with the Compose plugin;
- Go.

The complete setup is:

```bash
az login
make setup
```

`make setup` checks dependencies, initializes and interactively applies the Entra bootstrap Terraform, generates the Docker/Django `.env` and the installed CLI configuration from Terraform outputs, starts the API, builds the Go CLI, and installs it at `~/.local/bin/platform`.

Create the Django admin user:

```bash
make admin
```

Use the installed CLI:

```bash
platform login
platform whoami
platform app list
```

If the API rejects a cached token, clear the local token and authenticate again:

```bash
platform logout
platform login
```

Logout removes only the cached authentication token. Tokens are stored in the operating system credential store through the Go keyring integration: macOS Keychain, Windows Credential Manager, or Linux Secret Service. It preserves the generated CLI configuration.

On Linux, a Secret Service provider such as GNOME Keyring or another compatible desktop credential store must be running. The CLI reports an actionable error if no OS credential store is available; it does not silently fall back to plaintext token storage. Existing plaintext token files are migrated to the OS credential store on first use and then removed.

If `platform` is not found, add the user-local bin directory to the current shell:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

The setup can also be run in smaller steps:

```bash
make doctor
make bootstrap-plan
make bootstrap-apply
make env
make up
make install-cli
```

`make env` is idempotent: it regenerates Terraform-derived values, preserves the existing Django secret, writes both configuration files atomically, and sets permissions to `0600`. The root `.env` is the Docker/Django configuration source. The installed CLI configuration is written to `~/.config/spx-platform/config.env` (or the directory selected by `XDG_CONFIG_HOME`) and contains only `PLATFORM_*` values; the Django secret is never copied there. Both files are local-only and ignored by Git.

The CLI loads configuration in this order, with later entries overriding earlier ones:

1. the user-level CLI configuration;
2. the nearest project `.env` in the current directory or a parent directory;
3. explicit `PLATFORM_*` process environment variables.

This means the installed binary works from the repository, from `cli/`, or from another directory after `make setup`. No Makefile CLI command or manual `source .env` step is required.

The CLI can be rebuilt or removed with:

```bash
make cli-build
make install-cli
make uninstall-cli
```

## Cleanup

Remove only disposable local state:

```bash
make clean-local
```

This removes the local Docker Compose resources, SQLite database, generated `.env`, generated user CLI configuration, and installed CLI binary. It does not destroy the Entra app registrations or Terraform state. `make clean-local` does not log out. `make uninstall-cli` removes only the installed binary and leaves the CLI configuration and cached token untouched.

To completely reset the demo:

```bash
make teardown
```

`make teardown` first runs the interactive Terraform destroy, then logs out locally by clearing the cached token, removes local Docker resources, `.env`, SQLite, the installed CLI and generated CLI configuration, and finally removes local Terraform state. If Terraform destroy fails, the cleanup stops before logging out or deleting Terraform state.

The Terraform provider lock file is preserved so the next setup remains reproducible.

Open Django Admin at `http://localhost:8000/admin/` to inspect Platform Users, Applications, and Pending Provisioning Requests.

## Verification

```bash
make test
```

Troubleshooting targets:

```bash
make logs
make down
make up
```

Workload provisioning, Celery, Redis, Azure DevOps, and Terraform execution against Azure remain deferred from this milestone.
