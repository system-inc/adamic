#!/usr/bin/env python3
"""Reproduce parking comparisons without applying the shared harness to Git."""
import argparse
import json
import os
from pathlib import Path
import subprocess
import tarfile

parser = argparse.ArgumentParser()
parser.add_argument("--scratch", type=Path, required=True)
parser.add_argument("--typescript", type=Path, required=True)
args = parser.parse_args()
repository = Path(__file__).resolve().parents[5]
root = args.scratch.resolve()
root.mkdir(parents=True, exist_ok=False)
archive = root / "source.tar"
with archive.open("wb") as output:
    subprocess.run(["git", "archive", "HEAD"], cwd=repository, stdout=output, check=True)
with tarfile.open(archive) as source:
    source.extractall(root, filter="data")
owned = {p.name for p in (root / "stage1/cohere/lint/rules").iterdir() if p.is_dir()}
harness = "6a66e2d3e42854d4385801d032f17cff314f55b7"
paths = subprocess.check_output(["git", "ls-tree", "-r", "--name-only", harness,
                                 "stage1/cohere/lint"], cwd=repository, text=True).splitlines()
for name in paths:
    parts = Path(name).parts
    if len(parts) > 4 and parts[3] == "rules" and parts[4] in owned:
        continue
    target = root / name
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_bytes(subprocess.check_output(["git", "show", harness + ":" + name], cwd=repository))
cohere = root / "cohere"
if cohere.is_dir() and not list(cohere.iterdir()):
    cohere.rmdir()
cohere.symlink_to(repository / "cohere", target_is_directory=True)
overlay_dir = root / "parking-overlay"
with (root / "prepare.log").open("w") as output:
    subprocess.run(["python3", str(root / "stage1/cohere/lint/rules/no-multi-str/validate.py"),
                    "--scratch", str(overlay_dir), "--typescript", str(args.typescript),
                    "--prepare-only"], cwd=root, stdout=output, stderr=subprocess.STDOUT, check=True)
overlay = overlay_dir / "overlay.json"
data = json.loads(overlay.read_text())
for batch, directory in [("original", "no-useless-computed-key"),
                         ("second", "structure-tailwind-no-physical-direction"),
                         ("third", "typescript-no-unnecessary-type-constraint"),
                         ("fourth", "default-case-last"), ("fifth", "no-constructor-return")]:
    data["Replace"][str(root / ("stage1/cohere/lint/wave08_" + batch + "_test.go"))] = str(root / "stage1/cohere/lint/rules" / directory / "validation_test.go.txt")
overlay.write_text(json.dumps(data))
environment = os.environ.copy()
environment["ADAMIC_GATE_UNCACHED"] = "1"
environment["ADAMIC_TYPESCRIPT_SOURCE"] = str(args.typescript.resolve())
commands = [
    ("parity.log", ["go", "test", "-overlay=" + str(overlay), "./stage1/cohere/lint", "-count=1", "-v", "-timeout=20m", "-run", "^(TestOwnedWitnesses|TestOriginalUpstream|TestWave08UpstreamSupported|TestThirdUpstream|TestFourthUpstream|TestFifthUpstream|TestSixthUpstream|TestThirdShapes|TestFifthShapes|TestCompilerAndStage1Agree)$"]),
    ("shapes.log", ["go", "test", "-overlay=" + str(overlay), "./stage1/cohere/lint", "-count=1", "-v", "-timeout=10m", "-run", "^TestWave08Shapes$/^NonTSX$"]),
    ("mutants.log", ["go", "test", "-overlay=" + str(overlay), "./stage1/cohere/lint", "-count=1", "-v", "-timeout=20m", "-run", "^TestMutants$/(" + "|".join(json.loads(p.read_text())["name"] for p in sorted((root / "stage1/cohere/lint/rules").glob("*/mutant.json")) if p.parent.name in owned) + ")$"]),
    ("registry.log", ["go", "test", "-overlay=" + str(overlay), "./stage1/cohere/lint/registry", "-count=1", "-v"]),
    ("vet.log", ["go", "vet", "-overlay=" + str(overlay), "./..."]),
    ("oracle.log", ["go", "test", "./internal/oracle", "-run", "^TestTheOracleCatchesOneByte$", "-count=1", "-v", "-timeout=10m"]),
]
for name, command in commands:
    print(" ".join(command), flush=True)
    with (root / name).open("w") as output:
        result = subprocess.run(command, cwd=root, env=environment, stdout=output, stderr=subprocess.STDOUT)
    print("exit=" + str(result.returncode) + " log=" + str(root / name), flush=True)
    if result.returncode:
        raise SystemExit(result.returncode)
