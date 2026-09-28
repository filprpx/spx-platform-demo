#!/usr/bin/env python3
import json
import os
import subprocess
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
WORKLOADS = ROOT / ".local" / "spx" / "applied"


def load_env():
    env_file = ROOT / ".env"
    if not env_file.exists():
        return
    for line in env_file.read_text(encoding="utf-8").splitlines():
        if line and not line.startswith("#") and "=" in line:
            key, value = line.split("=", 1)
            os.environ.setdefault(key, value)


def main():
    load_env()
    candidates = sorted(WORKLOADS.glob("*.json"))
    if not candidates:
        print("No applied workload infrastructure found.")
        return 0

    for candidate in candidates:
        metadata = json.loads(candidate.read_text(encoding="utf-8"))
        workload = Path(metadata["workload_dir"]).resolve()
        state_dir = Path(metadata["state_dir"]).resolve()
        if ROOT not in workload.parents or not workload.is_dir():
            raise RuntimeError(f"workload directory is outside this repository or missing: {workload}")
        answer = input(f"Destroy {metadata['owning_team']}/{metadata['application']}? [y/N] ").strip().lower()
        if answer != "y":
            print(f"Retained {metadata['application']}.")
            continue
        plan_file = state_dir / "destroy.tfplan"
        env = os.environ.copy()
        if not env.get("ARM_SUBSCRIPTION_ID"):
            env["ARM_SUBSCRIPTION_ID"] = subprocess.check_output(
                ["az", "account", "show", "--query", "id", "--output", "tsv"], text=True
            ).strip()
        subprocess.run(["terraform", "-chdir=" + str(workload), "init", "-input=false", f"-backend-config=path={state_dir / 'terraform.tfstate'}"], check=True, env=env)
        subprocess.run(["terraform", "-chdir=" + str(workload), "plan", "-destroy", "-input=false", f"-out={plan_file}"], check=True, env=env)
        subprocess.run(["terraform", "-chdir=" + str(workload), "apply", "-input=false", str(plan_file)], check=True, env=env)
        candidate.unlink()
        print(f"Destroyed {metadata['application']} infrastructure.")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except KeyboardInterrupt:
        print("\nWorkload destruction cancelled.", file=sys.stderr)
        raise SystemExit(130)
    except Exception as exc:
        print(f"Workload destruction failed: {exc}", file=sys.stderr)
        raise SystemExit(1)
