#!/usr/bin/env python3
"""Prove both new checker operations with compiling Go overlays."""
import json
import subprocess
import sys
from pathlib import Path

repository = Path(__file__).resolve().parents[4]
scratch = Path(sys.argv[1]).resolve()
scratch.mkdir(parents=True, exist_ok=True)
for name, operation in [("awaited", "Checker_getAwaitedType"),
                        ("constraint", "Checker_getBaseConstraintOfType")]:
    source = repository / "bridge/tsgo/checker" / (name + "_shape.go")
    text = source.read_text()
    before = "checker." + operation + "(c, p.typesByID[id-1])"
    assert text.count(before) == 1
    mutant = scratch / (name + "-mutant.go")
    mutant.write_text(text.replace(before, "p.typesByID[id-1]"))
    overlay = scratch / (name + "-mutant.json")
    overlay.write_text(json.dumps({"Replace": {str(source): str(mutant)}}))
    log = scratch / (name + "-mutant.log")
    with log.open("wb") as output:
        result = subprocess.run(["go", "test", "-overlay=" + str(overlay),
            "./bridge/tsgo/checker", "-run", "^TestWave02CheckerQuestions$",
            "-count=1", "-v"], cwd=repository, stdout=output, stderr=subprocess.STDOUT)
    assert result.returncode == 1
    assert name + "-shape differs from direct checker operation" in log.read_text()
    print(name + ": compiling mutant killed by direct checker identity comparison")
