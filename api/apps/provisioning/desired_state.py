import json
import hashlib
import os
import re
from pathlib import Path

from jinja2 import Environment, FileSystemLoader, StrictUndefined

from .contracts import ProvisioningJob


class DesiredStateError(ValueError):
    pass


def _slug(value: str, field: str) -> str:
    normalized = value.strip().lower()
    if not re.fullmatch(r"[a-z0-9]+(?:-[a-z0-9]+)*", normalized):
        raise DesiredStateError(f"{field} must contain only lowercase letters, numbers, and hyphens")
    return normalized


def _azure_name(prefix: str, team: str, application: str, limit: int) -> str:
    base = f"{prefix}-{team}-{application}"
    if len(base) <= limit:
        return base
    suffix = hashlib.sha256(base.encode("utf-8")).hexdigest()[:8]
    return f"{base[: limit - len(suffix) - 1]}-{suffix}"


def write_desired_state(job: ProvisioningJob, repository_root: Path) -> Path:
    team = _slug(job.application.owning_team, "owning_team")
    application = _slug(job.application.name, "application name")
    workload = repository_root / "infra" / "workloads" / "dev" / "teams" / team / application
    if workload.exists() and any(workload.iterdir()):
        raise DesiredStateError(f"workload already exists: {workload}")
    workload.mkdir(parents=True, exist_ok=True)

    state_dir = repository_root / ".local" / "spx" / "terraform-state" / str(job.provisioning_request_id)
    state_dir.mkdir(parents=True, exist_ok=True)
    acr_name = os.environ.get("PLATFORM_ACR_NAME", "")
    acr_login_server = os.environ.get("PLATFORM_ACR_LOGIN_SERVER", "")
    bootstrap_resource_group_name = os.environ.get("PLATFORM_BOOTSTRAP_RESOURCE_GROUP", "spx-demo-bootstrap")
    if not acr_name or not acr_login_server:
        raise DesiredStateError("PLATFORM_ACR_NAME and PLATFORM_ACR_LOGIN_SERVER are required")
    size = {
        "small": {"cpu": 0.25, "memory": "0.5Gi"},
        "medium": {"cpu": 0.5, "memory": "1Gi"},
    }[job.application.compute_size]
    context = {
        "application_name": application,
        "owning_team": team,
        "location": "eastus",
        "compute_cpu": size["cpu"],
        "compute_memory": size["memory"],
        "container_port": job.application.container_port,
        "ingress_external": job.application.ingress == "external",
        "min_replicas": job.application.min_replicas,
        "max_replicas": job.application.max_replicas,
        "acr_name": acr_name,
        "acr_login_server": acr_login_server,
        "bootstrap_resource_group_name": bootstrap_resource_group_name,
        "resource_group_name": _azure_name("spx", team, application, 90),
        "environment_name": _azure_name("spx-env", team, application, 32),
        "workspace_name": _azure_name("spx-logs", team, application, 63),
        "identity_name": _azure_name("spx-id", team, application, 128),
        "container_app_name": _azure_name("spx", team, application, 32),
    }
    template_dir = Path(__file__).parent / "templates" / "terraform"
    environment = Environment(
        loader=FileSystemLoader(template_dir),
        undefined=StrictUndefined,
        autoescape=False,
        keep_trailing_newline=True,
    )
    template_files = {
        "versions.tf": "versions.tf.j2",
        "backend.tf": "backend.tf.j2",
        "variables.tf": "variables.tf.j2",
        "main.tf": "main.tf.j2",
        "terraform.tfvars": "terraform.tfvars.j2",
    }
    for output_name, template_name in template_files.items():
        rendered = environment.get_template(template_name).render(**context)
        (workload / output_name).write_text(rendered, encoding="utf-8")

    pending = repository_root / ".local" / "spx" / "pending"
    pending.mkdir(parents=True, exist_ok=True)
    metadata = {
        "request_id": str(job.provisioning_request_id),
        "workload_dir": str(workload),
        "state_dir": str(state_dir),
        "application": application,
        "owning_team": team,
        "resource_group_name": context["resource_group_name"],
        "container_app_name": context["container_app_name"],
        "acr_name": acr_name,
        "acr_login_server": acr_login_server,
        "bootstrap_resource_group_name": bootstrap_resource_group_name,
    }
    (pending / f"{job.provisioning_request_id}.json").write_text(
        json.dumps(metadata, indent=2) + "\n", encoding="utf-8"
    )
    return workload
