"""Run the whole lint package with all external input sets and record its results."""
import collections
import gzip
import json
import os
import subprocess
import tempfile
import time
from pathlib import Path

ROOT = Path(__file__).resolve().parents[6]
EVIDENCE = Path(__file__).resolve().parent / "evidence"
PIN = "050880ce59e30b356b686bd3144efe24f875ebc8"
source = os.environ["ADAMIC_TYPESCRIPT_SOURCE"]
assert subprocess.check_output(["git", "-C", source, "rev-parse", "HEAD"], text=True).strip() == PIN
assert not subprocess.check_output(["git", "-C", source, "status", "--porcelain"], text=True).strip()
assert Path(os.environ["WASI_SYSROOT"]).is_dir()
profile = tempfile.mkdtemp(prefix="no-cond-assign-profile-")
environment = dict(os.environ, ADAMIC_LINT_PROFILE_DIR=profile,
                   ADAMIC_LINT_PROFILE_SNAPSHOTS=profile, ADAMIC_LINT_BENCH="1",
                   ADAMIC_TEST_WASI="1")
command = ["go", "test", "./stage1/cohere/lint", "-count=1", "-timeout=60m", "-json"]
metadata = {"command": command, "compiler": source, "pin": PIN,
            "profile": profile, "WASI_SYSROOT": environment["WASI_SYSROOT"],
            "nproc": int(subprocess.check_output(["nproc"], text=True)),
            "load_before": Path("/proc/loadavg").read_text().strip()}
started = time.monotonic()
with (EVIDENCE / "full.jsonl").open("w") as output:
    result = subprocess.run(command, cwd=ROOT, env=environment, stdout=output, stderr=output)
metadata.update(exit=result.returncode, wall_seconds=time.monotonic() - started,
                load_after=Path("/proc/loadavg").read_text().strip())
counts = collections.Counter()
skips = []
with (EVIDENCE / "full.log").open("w") as output:
    for line in (EVIDENCE / "full.jsonl").read_text().splitlines():
        try:
            event = json.loads(line)
        except json.JSONDecodeError:
            output.write(line + "\n")
            continue
        output.write(event.get("Output", ""))
        if event.get("Test") and event["Action"] in ("pass", "fail", "skip"):
            counts[event["Action"]] += 1
            if event["Action"] == "skip": skips.append(event["Test"])
metadata.update(counts={action: counts[action] for action in ("pass", "fail", "skip")}, skips=skips)
(EVIDENCE / "full-summary.json").write_text(json.dumps(metadata, indent=2) + "\n")
with (EVIDENCE / "full.jsonl.gz").open("wb") as output:
    output.write(gzip.compress((EVIDENCE / "full.jsonl").read_bytes(), mtime=0))
(EVIDENCE / "full.jsonl").unlink()
print(json.dumps(metadata, indent=2))
raise SystemExit(result.returncode)
