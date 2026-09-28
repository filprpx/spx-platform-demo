#!/usr/bin/env python3
import json
import os
import subprocess
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
APPLIED = ROOT / ".local" / "spx" / "applied"
APP_SOURCE = ROOT / "examples" / "simple-api"


def load_env():
    env_file = ROOT / ".env"
    if not env_file.exists():
        raise RuntimeError("missing .env; run make setup first")
    for line in env_file.read_text(encoding="utf-8").splitlines():
        if line and not line.startswith("#") and "=" in line:
            key, value = line.split("=", 1)
            os.environ.setdefault(key, value)


def run(command, **kwargs):
    print("$", " ".join(command), flush=True)
    return subprocess.run(command, check=True, text=True, **kwargs)


def select_metadata():
    candidates = sorted(APPLIED.glob("*.json"))
    if not candidates:
        raise RuntimeError("no applied workload found; run make iac-pipeline first")
    if len(candidates) == 1:
        return candidates[0]
    print("Multiple applied workloads found:")
    for index, candidate in enumerate(candidates, 1):
        metadata = json.loads(candidate.read_text(encoding="utf-8"))
        print(f"  {index}. {metadata['owning_team']}/{metadata['application']}")
    try:
        return candidates[int(input("Select a workload number: ").strip()) - 1]
    except (ValueError, IndexError) as exc:
        raise RuntimeError("invalid workload selection") from exc


def main():
    load_env()
    if not APP_SOURCE.is_dir():
        raise RuntimeError(f"missing demo application: {APP_SOURCE}")
    if not shutil_which("docker"):
        raise RuntimeError("docker is required to deploy the demo application")
    if not shutil_which("az"):
        raise RuntimeError("az is required to deploy the demo application")
    run(["docker", "info"], stdout=subprocess.DEVNULL)
    run(["az", "account", "show"], stdout=subprocess.DEVNULL)

    metadata = json.loads(select_metadata().read_text(encoding="utf-8"))
    acr_name = metadata["acr_name"]
    acr_login_server = metadata["acr_login_server"]
    image = f"{acr_login_server}/{metadata['owning_team']}/{metadata['application']}:latest"

    run(["az", "acr", "login", "--name", acr_name])
    run(["docker", "build", "-t", image, str(APP_SOURCE)])
    run(["docker", "push", image])
    run([
        "az", "containerapp", "update",
        "--name", metadata["container_app_name"],
        "--resource-group", metadata["resource_group_name"],
        "--image", image,
    ])
    result = subprocess.run(
        [
            "az", "containerapp", "show",
            "--name", metadata["container_app_name"],
            "--resource-group", metadata["resource_group_name"],
            "--query", "properties.configuration.ingress.fqdn",
            "--output", "tsv",
        ],
        check=True,
        text=True,
        capture_output=True,
    )
    fqdn = result.stdout.strip()
    print(f"Application deployed: {image}")
    if fqdn:
        print("\nTo test the API, run:\n")
        print(f"  curl https://{fqdn}")
    return 0


def shutil_which(command):
    for directory in os.environ.get("PATH", "").split(os.pathsep):
        candidate = Path(directory) / command
        if candidate.is_file() and os.access(candidate, os.X_OK):
            return str(candidate)
    return None


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except subprocess.CalledProcessError as exc:
        print(f"Command failed with exit code {exc.returncode}.", file=sys.stderr)
        raise SystemExit(exc.returncode or 1)
    except Exception as exc:
        print(f"Application deployment failed: {exc}", file=sys.stderr)
        raise SystemExit(1)
