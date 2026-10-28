#!/usr/bin/env python3
"""Run Stage 1 experiments in a Go overlay, or unskip the real CLI proof."""
import json
import pathlib
import subprocess
import sys
import tempfile

root = pathlib.Path(__file__).resolve().parents[4]
assets = pathlib.Path(__file__).resolve().parent
with tempfile.TemporaryDirectory(prefix="gd4csnt-overlay-") as scratch:
    scratch = pathlib.Path(scratch)
    if "--proof" in sys.argv:
        original = root / "internal/native/source_paths_test.go"
        lines = original.read_text().splitlines(keepends=True)
        changed = scratch / "source_paths_test.go"
        changed.write_text("".join(line for line in lines if 't.Skip("#gd4csnt:' not in line))
        replacements = {str(original): str(changed)}
        test = "^TestDebugSourcePaths$"
    else:
        replacements = {str(root / "internal/native/source_paths_scratch_test.go"):
                        str(assets / "experiment.go.txt")}
        test = "^TestDebugPathsScratch$"
    overlay = scratch / "overlay.json"
    overlay.write_text(json.dumps({"Replace": replacements}))
    result = subprocess.run(["go", "test", "-overlay", str(overlay), "./internal/native",
                             "-run", test, "-v", "-count=1", "-timeout", "600s"], cwd=root)
    sys.exit(result.returncode)
