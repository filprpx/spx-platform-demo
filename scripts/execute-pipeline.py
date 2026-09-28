#!/usr/bin/env python3
import json
import os
import shutil
import subprocess
import sys
from pathlib import Path
from urllib.request import Request, urlopen


ROOT = Path(__file__).resolve().parents[1]
PENDING = ROOT / ".local" / "spx" / "pending"
APPLIED = ROOT / ".local" / "spx" / "applied"


def load_env():
    env_file = ROOT / ".env"
    if not env_file.exists():
        raise RuntimeError("missing .env; run make setup first")
    for line in env_file.read_text(encoding="utf-8").splitlines():
        if line and not line.startswith("#") and "=" in line:
            key, value = line.split("=", 1)
            os.environ.setdefault(key, value)
    if not os.environ.get("ARM_SUBSCRIPTION_ID"):
        try:
            subscription = subprocess.check_output(
                ["az", "account", "show", "--query", "id", "--output", "tsv"],
                text=True,
            ).strip()
        except (OSError, subprocess.CalledProcessError) as exc:
            raise RuntimeError("unable to read the active Azure subscription; run az login first") from exc
        if not subscription:
            raise RuntimeError("the active Azure subscription is empty; run az account set")
        os.environ["ARM_SUBSCRIPTION_ID"] = subscription


def event(request_id, status, message):
    token = os.environ.get("WORKER_CALLBACK_TOKEN", "")
    if not token:
        raise RuntimeError("WORKER_CALLBACK_TOKEN is not configured")
    body = json.dumps({
        "schema_version": 2,
        "job_id": request_id,
        "provisioning_request_id": request_id,
        "status": status,
        "message": message,
    }).encode()
    request = Request(
        os.environ.get("PLATFORM_API_URL", "http://localhost:8000").rstrip("/") + "/internal/execution-events",
        data=body,
        headers={"Content-Type": "application/json", "X-Worker-Token": token},
        method="POST",
    )
    with urlopen(request, timeout=10):
        pass


def run_terraform(workload, state_dir, *args):
    command = ["terraform", *args]
    print("$", " ".join(command), flush=True)
    subprocess.run(command, cwd=workload, check=True)


def main():
    load_env()
    candidates = sorted(PENDING.glob("*.json"))
    if not candidates:
        print("No workload is waiting for pipeline execution.")
        return 0
    candidate = candidates[0]
    if len(candidates) > 1:
        print("Multiple workloads are waiting for pipeline execution:")
        for index, candidate in enumerate(candidates, 1):
            metadata = json.loads(candidate.read_text(encoding="utf-8"))
            print(f"  {index}. {metadata['owning_team']}/{metadata['application']}")
        choice = input("Select a workload number: ").strip()
        try:
            candidate = candidates[int(choice) - 1]
        except (ValueError, IndexError) as exc:
            raise RuntimeError("invalid workload selection") from exc

    metadata = json.loads(candidate.read_text(encoding="utf-8"))
    request_id = metadata["request_id"]
    workload = Path(metadata["workload_dir"]).resolve()
    state_dir = Path(metadata["state_dir"]).resolve()
    if ROOT not in workload.parents or not workload.is_dir():
        raise RuntimeError("pending workload directory is outside this repository or missing")

    print("Executing the local pipeline.")
    print("In the real platform, this represents an approved pull request and CI pipeline.")
    try:
        run_terraform(workload, state_dir, "fmt")
        run_terraform(workload, state_dir, "init", "-input=false", f"-backend-config=path={state_dir / 'terraform.tfstate'}")
        run_terraform(workload, state_dir, "validate")
        plan_file = state_dir / "plan.tfplan"
        run_terraform(workload, state_dir, "plan", "-input=false", f"-out={plan_file}")
    except subprocess.CalledProcessError as exc:
        event(request_id, "EXECUTION_FAILED", f"Terraform plan failed with exit code {exc.returncode}")
        return exc.returncode or 1

    if input("Apply this Terraform plan? [y/N] ").strip().lower() != "y":
        print("Plan retained. The request remains waiting for approval.")
        return 0

    event(request_id, "APPLYING", "Terraform apply started")
    try:
        run_terraform(workload, state_dir, "apply", "-input=false", str(plan_file))
    except subprocess.CalledProcessError as exc:
        event(request_id, "EXECUTION_FAILED", f"Terraform apply failed with exit code {exc.returncode}")
        return exc.returncode or 1
    event(request_id, "SUCCEEDED", "Terraform apply completed successfully")
    APPLIED.mkdir(parents=True, exist_ok=True)
    shutil.copy2(candidate, APPLIED / candidate.name)
    candidate.unlink()
    print("Pipeline completed successfully.")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except KeyboardInterrupt:
        print("\nPipeline cancelled.", file=sys.stderr)
        raise SystemExit(130)
    except Exception as exc:
        print(f"Pipeline failed: {exc}", file=sys.stderr)
        raise SystemExit(1)
