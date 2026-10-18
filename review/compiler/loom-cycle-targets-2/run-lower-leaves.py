import json
from pathlib import Path
import subprocess
import time

root = Path(__file__).resolve().parents[3]
evidence = Path(__file__).resolve().parent
binary = "/tmp/adamic-gate/loom-cycle-lower.test"
with (evidence / "lower-build.log").open("w") as log:
    subprocess.run(["go", "test", "./internal/lower", "-c", "-o", binary], cwd=root, stdout=log, stderr=subprocess.STDOUT, timeout=180, check=True)
listing = subprocess.run([binary, "-test.list=^Test"], cwd=root / "internal/lower", capture_output=True, text=True, timeout=20, check=True)
(evidence / "lower-list.log").write_text(listing.stdout)
names = [name for name in listing.stdout.splitlines() if name.startswith("Test")]
results = []
for name in names:
    arguments = [binary, "-test.run=^" + name + "$", "-test.count=1", "-test.timeout=60s", "-test.parallel=4", "-test.v"]
    started = time.monotonic()
    with (evidence / (name + ".log")).open("w") as log:
        try:
            run = subprocess.run(arguments, cwd=root / "internal/lower", stdout=log, stderr=subprocess.STDOUT, timeout=60)
            code = run.returncode
        except subprocess.TimeoutExpired:
            code = 124
    seconds = time.monotonic() - started
    results.append(dict(test=name, args=arguments, exit=code, seconds=seconds))
    (evidence / "lower-results.json").write_text(json.dumps(results, indent=2) + "\n")
    print(f"{name}: exit={code} seconds={seconds:.3f}", flush=True)
    if seconds >= 60:
        raise SystemExit("leaf exceeded its budget; it will not be rerun")
failed = [row for row in results if row["exit"]]
print(f"{len(results)} top-level tests, {len(failed)} failures", flush=True)
raise SystemExit(bool(failed))
