#!/usr/bin/env python3
"""Build a scratch Go overlay. Never edit shared registry, dispatch, or test files."""
import json
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile

owned = Path(__file__).resolve().parent
repository = owned.parents[4]
scratch = Path(sys.argv[1]).resolve() if len(sys.argv) > 1 else Path(tempfile.mkdtemp(prefix="lint-wave1-02-third-"))
scratch.mkdir(parents=True, exist_ok=True)
replacements = {}
for relative in ("stage1/cohere/lint/registry/registry.go", "stage1/cohere/lint/lint_test.go"):
    original = repository / relative
    target = scratch / relative
    target.parent.mkdir(parents=True, exist_ok=True)
    shutil.copyfile(original, target)
    replacements[str(original)] = str(target)
with (scratch / "patch.log").open("w") as log:
    subprocess.run(["patch", "-p1", "-i", str(owned / "compatibility.patch")], cwd=scratch, stdout=log, stderr=subprocess.STDOUT, check=True)
test = scratch / "stage1/cohere/lint/lint_test.go"
with test.open("a") as output:
    output.write((owned / "validation.go.txt").read_text())
overlay = scratch / "overlay.json"
overlay.write_text(json.dumps({"Replace": replacements}))
print(overlay)
