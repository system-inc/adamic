import concurrent.futures
import json
import os
from pathlib import Path
import signal
import subprocess
import tempfile

root = Path(__file__).resolve().parents[3]
evidence = root / "review/compiler/cycle-closure-targets"
base = "/tmp/adamic-cycle-base"
head = "/tmp/adamic-cycle-head"

def run(arguments, limit=45):
    process = subprocess.Popen(arguments, cwd=root, stdout=subprocess.PIPE, stderr=subprocess.PIPE, start_new_session=True)
    try:
        out, err = process.communicate(timeout=limit)
    except subprocess.TimeoutExpired:
        os.killpg(process.pid, signal.SIGKILL)
        out, err = process.communicate()
        return {"exit": -1, "stdout": out.decode(errors="replace"), "stderr": "timeout"}
    return {"exit": process.returncode, "stdout": out.decode(errors="replace"), "stderr": err.decode(errors="replace")}

paths = sorted(path for path in (root / "internal/oracle/testdata").rglob("*") if path.suffix in (".a", ".ts"))
def classify(path):
    old = run([base, "c", str(path)])
    new = run([head, "c", str(path)])
    return {"path": str(path.relative_to(root)), "base_exit": old["exit"], "head_exit": new["exit"], "base_error": old["stderr"].splitlines()[:1], "head_error": new["stderr"].splitlines()[:1]}
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:
    records = list(pool.map(classify, paths))
newly = [record for record in records if record["base_exit"] != 0 and record["head_exit"] == 0]
for record in newly:
    path = root / record["path"]
    with tempfile.TemporaryDirectory(prefix="cycle-admission-") as scratch:
        scratch = Path(scratch)
        source = run(["node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", str(path)])
        js = run([head, "js", str(path)])
        js_path = scratch / "program.mjs"
        js_path.write_text(js["stdout"])
        backend = run(["node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", str(js_path)]) if js["exit"] == 0 else js
        binary = scratch / "program"
        build = run([head, "build", str(path), "-o", str(binary)])
        native = run([str(binary)]) if build["exit"] == 0 else build
        record.update(node=source, javascript=backend, native=native)
        record["agree"] = source["exit"] >= 0 and source["exit"] == backend["exit"] == native["exit"] and source["stdout"] == backend["stdout"] == native["stdout"]
        print(record["path"] + ": " + ("agree" if record["agree"] else "DISAGREE"), flush=True)
report = {"programs": len(records), "newly_admitted": len(newly), "newly_refused": [r["path"] for r in records if r["base_exit"] == 0 and r["head_exit"] != 0], "classifications": records, "observations": newly}
evidence.joinpath("admission-delta.json").write_text(json.dumps(report, indent=2) + "\n")
if any(not record["agree"] for record in newly) or any(r["base_exit"] < 0 or r["head_exit"] < 0 for r in records):
    raise SystemExit(1)
print("admission delta clean: " + str(len(paths)) + " programs; " + str(len(newly)) + " newly admitted", flush=True)
