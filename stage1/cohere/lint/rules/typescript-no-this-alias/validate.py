#!/usr/bin/env python3
"""Test the .a candidates through a reviewable scratch compatibility overlay.

No shared repository source is modified. Source env.sh first. The compatibility
patch is a pending proposal, not part of the installed registration foundation.
"""
import argparse
import json
import os
from pathlib import Path
import shutil
import subprocess

owned = Path(__file__).resolve().parent
repository = owned.parents[4]
parser = argparse.ArgumentParser()
parser.add_argument("--scratch", type=Path, required=True)
parser.add_argument("--typescript", type=Path, required=True)
parser.add_argument("--prepare-only", action="store_true")
args = parser.parse_args()
scratch = args.scratch.resolve()
scratch.mkdir(parents=True, exist_ok=True)
paths = ["stage1/cohere/lint/registry/registry.go",
         "stage1/cohere/lint/lint_test.go",
         "stage1/cohere/lint/profile_test.go"]
for relative in paths:
    target = scratch / relative
    target.parent.mkdir(parents=True, exist_ok=True)
    shutil.copyfile(repository / relative, target)
with (scratch / "patch.log").open("w") as log:
    subprocess.run(["patch", "--batch", "-p1", "-d", str(scratch),
                    "-i", str(owned.parent / "nexus-consistency-no-boolean-outcome" / "compatibility.patch")],
                   stdout=log, stderr=subprocess.STDOUT, check=True)
replacements = {str(repository / relative): str(scratch / relative) for relative in paths}
replacements[str(repository / "stage1/cohere/lint/wave12_overlay_test.go")] = str(owned / "validation_test.go.txt")
overlay = scratch / "overlay.json"
overlay.write_text(json.dumps({"Replace": replacements}))
if args.prepare_only:
    print(f"overlay={overlay}")
    raise SystemExit(0)
environment = os.environ.copy()
environment["ADAMIC_GATE_UNCACHED"] = "1"
environment["ADAMIC_TYPESCRIPT_SOURCE"] = str(args.typescript.resolve())
commands = [
    ("parity.log", ["go", "test", "-overlay=" + str(overlay), "./stage1/cohere/lint", "-count=1", "-v", "-timeout=20m",
                    "-run", "^(TestOwnedWitnesses|TestCompilerAndStage1Agree|TestAliasUpstream|TestAliasOptions|TestAliasThroughput)$"]),
    ("mutants.log", ["go", "test", "-overlay=" + str(overlay), "./stage1/cohere/lint", "-count=1", "-v", "-timeout=10m",
                     "-run", "^TestMutants$/(this_alias_allowed_name_ignored)$"]),
    ("registry.log", ["go", "test", "-overlay=" + str(overlay), "./stage1/cohere/lint/registry", "-count=1", "-v"]),
    ("vet.log", ["go", "vet", "-overlay=" + str(overlay), "./..."]),
    ("oracle.log", ["go", "test", "./internal/oracle", "-run", "^TestTheOracleCatchesOneByte$", "-count=1", "-v", "-timeout=10m"]),
]
for name, command in commands:
    print(" ".join(command), flush=True)
    with (scratch / name).open("w") as log:
        result = subprocess.run(command, cwd=repository, env=environment,
                                stdout=log, stderr=subprocess.STDOUT)
    print(f"exit={result.returncode} log={scratch / name}", flush=True)
    if result.returncode:
        raise SystemExit(result.returncode)
