"""Run the complete lint package with its required inputs and durable metrics."""
import json
import os
import subprocess
import tempfile
import time
from pathlib import Path
ROOT = Path(__file__).resolve().parents[6]
CORPUS = Path(os.environ.get("ADAMIC_TYPESCRIPT_SOURCE", "/workspace/scratch/typescript-6.0.3"))
assert subprocess.check_output(["git","rev-parse","HEAD"],cwd=CORPUS,text=True).strip() == "050880ce59e30b356b686bd3144efe24f875ebc8"
assert not subprocess.check_output(["git","status","--porcelain","--ignored"],cwd=CORPUS)
assert Path(os.environ["WASI_SYSROOT"]).is_dir()
profile = tempfile.mkdtemp(prefix="structure-react-component-no-display-name-profiles-")
env = dict(os.environ, ADAMIC_TYPESCRIPT_SOURCE=str(CORPUS), ADAMIC_LINT_PROFILE_DIR=profile, ADAMIC_LINT_PROFILE_SNAPSHOTS=profile, ADAMIC_LINT_BENCH="1")
evidence = Path(__file__).parent.parent / "evidence"
evidence.mkdir(exist_ok=True)
log = evidence / "whole-package.jsonl"
metrics = {"profile":profile,"corpus":str(CORPUS),"wasiSysroot":env["WASI_SYSROOT"],"nproc":subprocess.check_output(["nproc"],text=True).strip(),"loadBefore":Path("/proc/loadavg").read_text().strip()}
start = time.monotonic()
with log.open("w") as output:
    result = subprocess.run(["go","test","./stage1/cohere/lint","-count=1","-json","-timeout=120m"],cwd=ROOT,env=env,stdout=output,stderr=subprocess.STDOUT)
metrics.update(exitCode=result.returncode,wallSeconds=time.monotonic()-start,loadAfter=Path("/proc/loadavg").read_text().strip())
counts = {"pass":0,"fail":0,"skip":0}
skips = []
unfinished = set()
for line in log.read_text().splitlines():
    try: event=json.loads(line)
    except ValueError: continue
    if "Test" in event and event.get("Action") == "run": unfinished.add(event["Test"])
    if "Test" in event and event.get("Action") in counts:
        unfinished.discard(event["Test"])
        counts[event["Action"]] += 1
        if event["Action"] == "skip": skips.append(event["Test"])
metrics.update(counts=counts,skips=skips,unfinished=sorted(unfinished))
(evidence / "whole-package-metrics.json").write_text(json.dumps(metrics,indent=2)+"\n")
print(json.dumps(metrics,indent=2))
raise SystemExit(result.returncode)
