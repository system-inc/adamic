#!/usr/bin/env python3
"""Run owned landing probes using the unchanged shared oracle and comparison."""
import json, subprocess, sys, tempfile
from pathlib import Path
owned = Path(__file__).resolve().parent
root = owned.parents[4]
scratch = Path(tempfile.mkdtemp(prefix="wave14-integrated-probe-"))
overlay = scratch / "overlay.json"
overlay.write_text(json.dumps({"Replace": {str(root / "stage1/cohere/lint/slot14_landing_test.go"): str(owned / "landing_suite.go.txt")}}))
log = owned / "evidence/area-per-rule-upstream.log"
with log.open("wb") as output:
    result = subprocess.run(["go", "test", "-overlay=" + str(overlay), "./stage1/cohere/lint", "-run", "^TestWave14(IntegratedUpstream|LonelyIfMutant)$", "-count=1", "-timeout", "30m", "-v"], cwd=root, stdout=output, stderr=subprocess.STDOUT)
print("exit=" + str(result.returncode) + " log=" + str(log))
sys.exit(result.returncode)
