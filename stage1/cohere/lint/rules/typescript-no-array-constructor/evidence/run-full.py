import json
import os
from pathlib import Path
import statistics
import subprocess
import time

repository = Path(__file__).resolve().parents[6]
out = Path(os.environ.get("WAVE19_LOG_DIR", "/tmp/wave19-facts-full"))
out.mkdir(parents=True, exist_ok=True)
environment = os.environ.copy()
environment.update({
    "ADAMIC_TYPESCRIPT_SOURCE": "/workspace/wave19-typescript",
    "ADAMIC_LINT_BENCH": "1",
    "ADAMIC_LINT_PROFILE_DIR": str(out / "profiles"),
    "ADAMIC_LINT_PROFILE_SNAPSHOTS": str(out / "profiles"),
    "GOCACHE": "/tmp/wave19-unpark-cache/go-build",
    "XDG_CACHE_HOME": "/tmp/wave19-unpark-cache",
    "TMPDIR": "/tmp/adamic-gate",
    "GOMAXPROCS": "4",
})
command = ["go", "test", "-json", "-count=1", "-timeout", "3h", "./stage1/cohere/lint"]
(out / "inputs.json").write_text(json.dumps({"command": command, "environment": {key: environment[key] for key in ["ADAMIC_TYPESCRIPT_SOURCE", "ADAMIC_LINT_BENCH", "ADAMIC_LINT_PROFILE_DIR", "ADAMIC_LINT_PROFILE_SNAPSHOTS", "GOCACHE", "XDG_CACHE_HOME", "TMPDIR", "GOMAXPROCS"]}}, indent=2) + "\n")
loads = []
started = time.monotonic()
(out / "started.json").write_text(json.dumps({"epoch": time.time(), "nproc": int(subprocess.check_output(["nproc"]))}) + "\n")
with (out / "lint.jsonl").open("w") as log:
    process = subprocess.Popen(command, cwd=repository, env=environment, stdout=log, stderr=subprocess.STDOUT)
    while process.poll() is None:
        sample = os.getloadavg()[0]
        loads.append(sample)
        with (out / "load.jsonl").open("a") as samples:
            samples.write(json.dumps({"epoch": time.time(), "load1m": sample}) + "\n")
        time.sleep(1)
summary = {"exit": process.returncode, "wallSeconds": time.monotonic() - started, "nproc": int(subprocess.check_output(["nproc"])), "load1m": {"min": min(loads), "median": statistics.median(loads), "max": max(loads)}}
counts = {"pass": 0, "fail": 0, "skip": 0}
package = dict(counts)
skips = []
for line in (out / "lint.jsonl").read_text().split('\n'):
    try:
        row = json.loads(line)
    except ValueError:
        continue
    action = row.get("Action")
    if action in counts:
        (counts if "Test" in row else package)[action] += 1
        if action == "skip":
            skips.append(row.get("Test", row.get("Package")))
summary.update({"testCounts": counts, "packageCounts": package, "skips": skips})
(out / "summary.json").write_text(json.dumps(summary, indent=2) + "\n")
print(json.dumps(summary))
raise SystemExit(process.returncode)
