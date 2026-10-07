#!/usr/bin/env python3
"""Drop each warming-key component independently and require its regression to fail."""
import os
import argparse
from pathlib import Path
import subprocess
import tempfile

parser = argparse.ArgumentParser()
parser.add_argument("--integration", action="store_true")
arguments = parser.parse_args()
source = Path(__file__).resolve().parent
scratch = Path(tempfile.mkdtemp(prefix="setup-mutants-"))
print("mutant logs:", scratch)
original = (source / "setup-key.py").read_text()
for component in ["head", "sums", "version", "environment", "packages", "cache", "mode"]:
    lines = original.splitlines(keepends=True)
    mutant = "".join(line for line in lines if not line.lstrip().startswith(f'"{component}":'))
    assert mutant != original
    module = scratch / f"drop-{component}.py"
    module.write_text(mutant)
    environment = dict(os.environ, ADAMIC_SETUP_KEY_MODULE=str(module))
    environment.pop("ADAMIC_SETUP_INTEGRATION", None)
    log = scratch / f"drop-{component}.log"
    with log.open("wb") as output:
        result = subprocess.run(["python3", str(source / "test_setup.py"), "WarmingKey"],
                                env=environment, stdout=output, stderr=subprocess.STDOUT)
    text = log.read_text()
    if result.returncode != 1 or f"component='{component}'" not in text:
        raise SystemExit(f"mutant {component} escaped or failed for the wrong reason: {log}")
    print(f"drop {component}: caught by test_every_component_invalidates (exit 1)")

for name, old, new in [
    ("root-sum", 'manifests = {"go.mod", "go.sum"}', 'manifests = {"go.mod"}'),
    ("workspace-sum", 'manifests.update([workspace, workspace + ".sum"])', 'manifests.add(workspace)'),
    ("dependency-sum", 'manifests.add(str(path.with_name("go.sum")))', 'pass'),
]:
    module = scratch / f"omit-{name}.py"
    assert old in original
    module.write_text(original.replace(old, new))
    environment = dict(os.environ, ADAMIC_SETUP_KEY_MODULE=str(module))
    environment.pop("ADAMIC_SETUP_INTEGRATION", None)
    log = scratch / f"omit-{name}.log"
    with log.open("wb") as output:
        result = subprocess.run(["python3", str(source / "test_setup.py"), "WarmingKey"],
                                env=environment, stdout=output, stderr=subprocess.STDOUT)
    if result.returncode != 1 or "test_collected_manifests_invalidate" not in log.read_text():
        raise SystemExit(f"mutant {name} escaped or failed for the wrong reason: {log}")
    print(f"omit {name}: caught by test_collected_manifests_invalidate (exit 1)")

if arguments.integration:
    for component in ["head", "sums"]:
        log = scratch / f"real-go-drop-{component}.log"
        environment = dict(os.environ, ADAMIC_SETUP_KEY_MODULE=str(scratch / f"drop-{component}.py"),
                           ADAMIC_SETUP_INTEGRATION="1")
        with log.open("wb") as output:
            result = subprocess.run(["python3", str(source / "test_setup.py"), "SetupIntegration"],
                                    env=environment, stdout=output, stderr=subprocess.STDOUT)
        if result.returncode != 1 or "go build ready" not in log.read_text():
            raise SystemExit(f"real Go mutant {component} escaped or failed elsewhere: {log}")
        print(f"real Go drop {component}: stamp wrongly skipped rebuild, integration caught it (exit 1)")
