SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

ROOT_DIR := $(abspath .)
BOOTSTRAP_DIR := $(ROOT_DIR)/infra/bootstrap
CLI_DIR := $(ROOT_DIR)/cli
BUILD_DIR := $(ROOT_DIR)/build
CLI_BUILD := $(BUILD_DIR)/spx
CLI_INSTALL_DIR := $(HOME)/.local/bin
CLI_INSTALL := $(CLI_INSTALL_DIR)/spx
CLI_CONFIG_DIR := $(if $(XDG_CONFIG_HOME),$(XDG_CONFIG_HOME),$(HOME)/.config)/spx
CLI_CONFIG_FILE := $(CLI_CONFIG_DIR)/config.env

.PHONY: help doctor api-venv bootstrap-init bootstrap-plan bootstrap-apply bootstrap-destroy env up down logs admin worker worker-start worker-stop worker-status worker-logs iac-pipeline iac-destroy deploy-app clean-local clean-bootstrap-state teardown cli-build install-cli uninstall-cli test setup

help:
	@printf '%s\n' 'SPX setup targets:'
	@printf '%s\n' '  make setup             Verify, bootstrap, configure, start, and install the CLI'
	@printf '%s\n' '  make doctor            Check az, Terraform, Docker, Docker Compose, Go, and login state'
	@printf '%s\n' '  make bootstrap-plan    Preview Entra bootstrap changes'
	@printf '%s\n' '  make bootstrap-apply   Apply Entra bootstrap changes interactively'
	@printf '%s\n' '  make env               Generate .env and installed CLI config from Terraform outputs'
	@printf '%s\n' '  make up                Start the Django API'
	@printf '%s\n' '  make worker            Start the host-side worker in the background'
	@printf '%s\n' '  make worker-status     Show the background worker status'
	@printf '%s\n' '  make worker-logs       Follow the background worker log'
	@printf '%s\n' '  make iac-pipeline      Plan and apply generated workload Terraform'
	@printf '%s\n' '  make deploy-app        Build and deploy the bundled simple API to Azure'
	@printf '%s\n' '  make install-cli       Install spx under ~/.local/bin'
	@printf '%s\n' '  make clean-local       Remove local database, env, CLI config, Docker volumes, and CLI binary'
	@printf '%s\n' '  make teardown          Destroy Entra resources, then remove all local setup state'

doctor:
	@$(ROOT_DIR)/scripts/check-dependencies.sh

bootstrap-init:
	@terraform -chdir=$(BOOTSTRAP_DIR) init

bootstrap-plan: bootstrap-init
	@ARM_SUBSCRIPTION_ID="$$(az account show --query id --output tsv)" terraform -chdir=$(BOOTSTRAP_DIR) plan

bootstrap-apply: bootstrap-init
	@ARM_SUBSCRIPTION_ID="$$(az account show --query id --output tsv)" terraform -chdir=$(BOOTSTRAP_DIR) apply

bootstrap-destroy:
	@ARM_SUBSCRIPTION_ID="$$(az account show --query id --output tsv)" terraform -chdir=$(BOOTSTRAP_DIR) destroy

env:
	@$(ROOT_DIR)/scripts/generate-env.sh

api-venv:
	@command -v python3 >/dev/null 2>&1 || (printf 'python3 is required to run the host worker.\n' >&2; exit 1)
	@if [[ ! -x $(ROOT_DIR)/api/.venv/bin/python ]]; then python3 -m venv $(ROOT_DIR)/api/.venv; fi
	@$(ROOT_DIR)/api/.venv/bin/pip install -r $(ROOT_DIR)/api/requirements-dev.txt

up: env
	@docker compose up --build -d

down:
	@docker compose down

clean-local:
	@$(MAKE) worker-stop
	@if docker info >/dev/null 2>&1; then \
		docker compose down --volumes --remove-orphans; \
	else \
		printf 'Warning: Docker daemon unavailable; local Docker resources were not changed.\n'; \
	fi
	@rm -rf $(ROOT_DIR)/.local
	@rm -f $(ROOT_DIR)/.env $(ROOT_DIR)/api/db.sqlite3 $(CLI_INSTALL) $(CLI_CONFIG_FILE)
	@printf 'Removed local environment, SQLite database, CLI config, and installed CLI (if present).\n'

clean-bootstrap-state:
	@rm -f $(BOOTSTRAP_DIR)/terraform.tfstate $(BOOTSTRAP_DIR)/terraform.tfstate.backup
	@rm -rf $(BOOTSTRAP_DIR)/.terraform
	@printf 'Removed local Terraform state and working directory.\n'

teardown:
	@$(MAKE) worker-stop
	@$(MAKE) iac-destroy
	@$(MAKE) bootstrap-destroy
	@if [[ -x $(CLI_INSTALL) && -f $(CLI_CONFIG_FILE) ]]; then \
		$(CLI_INSTALL) logout; \
	elif [[ -f $(CLI_CONFIG_FILE) ]] && command -v go >/dev/null 2>&1; then \
		(cd $(CLI_DIR) && go run ./cmd/spx logout); \
	else \
		printf 'No CLI configuration found; no cached authentication token was present to clear.\n'; \
	fi
	@$(MAKE) clean-local
	@$(MAKE) clean-bootstrap-state
	@printf '\nTeardown complete. Source code and Terraform lock file were preserved.\n'

logs:
	@docker compose logs -f api

admin:
	@docker compose exec api python manage.py createsuperuser

worker: worker-start

worker-start: api-venv env
	@$(ROOT_DIR)/scripts/worker-start.sh

worker-stop:
	@$(ROOT_DIR)/scripts/worker-stop.sh

worker-status:
	@$(ROOT_DIR)/scripts/worker-status.sh

worker-logs:
	@test -f $(ROOT_DIR)/.local/spx/worker.log || (printf 'No worker log exists; run make worker first.\n' >&2; exit 1)
	@tail -f $(ROOT_DIR)/.local/spx/worker.log

iac-pipeline:
	@$(ROOT_DIR)/scripts/execute-pipeline.py

iac-destroy:
	@python3 $(ROOT_DIR)/scripts/destroy-workloads.py

deploy-app:
	@python3 $(ROOT_DIR)/scripts/deploy-app.py

cli-build:
	@mkdir -p $(BUILD_DIR)
	@cd $(CLI_DIR) && go build -o $(CLI_BUILD) ./cmd/spx
	@printf 'Built %s\n' '$(CLI_BUILD)'

install-cli: cli-build
	@mkdir -p $(CLI_INSTALL_DIR)
	@install -m 0755 $(CLI_BUILD) $(CLI_INSTALL)
	@printf 'Installed %s\n' '$(CLI_INSTALL)'
	@if [[ ":$$PATH:" != *":$(CLI_INSTALL_DIR):"* ]]; then \
		printf 'Warning: %s is not on PATH. Add:\n  export PATH="$$HOME/.local/bin:$$PATH"\n' '$(CLI_INSTALL_DIR)'; \
	fi

uninstall-cli:
	@if [[ -e $(CLI_INSTALL) ]]; then rm -f $(CLI_INSTALL); printf 'Removed %s\n' '$(CLI_INSTALL)'; else printf '%s\n' 'spx is not installed'; fi

test:
	@$(ROOT_DIR)/scripts/test-generate-env.sh
	@bash $(ROOT_DIR)/scripts/test-cleanup.sh
	@bash $(ROOT_DIR)/scripts/test-worker.sh
	@bash $(ROOT_DIR)/scripts/test-iac.sh
	@cd $(CLI_DIR) && go test ./...
	@if $(MAKE) --no-print-directory help | grep -Eq 'cli-login|cli-whoami|cli-list'; then \
		printf 'Removed CLI execution targets must not appear in help output.\n' >&2; \
		exit 1; \
	fi
	@if [[ -x $(ROOT_DIR)/api/.venv/bin/pytest ]]; then \
		cd $(ROOT_DIR)/api && .venv/bin/pytest -q; \
	else \
		docker compose run --rm api pytest -q; \
	fi

setup:
	@$(MAKE) doctor
	@$(MAKE) bootstrap-init
	@$(MAKE) bootstrap-apply
	@$(MAKE) env
	@$(MAKE) up
	@$(MAKE) worker-start
	@$(MAKE) cli-build
	@$(MAKE) install-cli
	@printf '\nSetup complete.\n\nNext:\n  make admin\n  spx login\n  spx whoami\n  spx app list\n\nWorker:\n  make worker-status\n  make worker-logs\n'
