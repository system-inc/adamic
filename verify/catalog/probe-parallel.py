"""Run the actual entry-07 fault while suppressing restoration; require rejection."""
import importlib.util
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import threading

catalog = Path(__file__).resolve().parent
repository = catalog.parent.parent
spec = importlib.util.spec_from_file_location("catalog_check", catalog / "check.py")
checker = importlib.util.module_from_spec(spec)
spec.loader.exec_module(checker)
entry = next(row for row in json.loads((catalog / "catalog.json").read_text()) if row["number"] == 7)
commit = subprocess.check_output(["git", "-C", str(repository), "rev-parse", sys.argv[1]], text=True).strip()
metadata = json.loads((catalog / "environment-parallel.json").read_text())
print("build flags: commit=" + commit + "; nproc=" + metadata["nproc"]
      + "; cpu.max=" + metadata["cpu_max"] + "; " + metadata["go_version"]
      + "; " + metadata["clang_version"] + "; Node " + metadata["node_version"]
      + "; ADAMIC_GATE_UNCACHED=1; go test -trimpath -count=1; native release -O2, sanitizer -O1 -g -fsanitize=address,undefined")
original_git = checker.git


def omit_restoration(worktree, *arguments, **options):
    if arguments[:2] == ("apply", "-R"):
        return subprocess.CompletedProcess(arguments, 0)
    return original_git(worktree, *arguments, **options)


checker.git = omit_restoration
with tempfile.TemporaryDirectory(prefix="adamic-restoration-probe-") as temporary:
    parent = Path(temporary)
    logs = catalog / "logs" / "parallel-restoration-probe"
    logs.mkdir(exist_ok=True)
    # Keep prior probe logs if explicitly rerun.
    run_logs = Path(tempfile.mkdtemp(prefix="run-", dir=logs))
    result = checker.check_one(repository, catalog, entry, commit, parent, run_logs, threading.Lock())
    assert not result["ok"], result
    assert any("worktree was not restored" in line for line in result["messages"]), result
    assert any("applies-and-fails-as-recorded" in line for line in result["messages"]), result
    assert not (parent / "entry-07").exists()
    for line in result["messages"]:
        print(line)
    print("Omitted restoration mutant: rejected after the actual fault was detected; own worktree removed")

# Audit measured runs, then make the two concurrency reporting invariants fail.
measurements = json.loads((catalog / "parallel-measurements.json").read_text())
for measurement in measurements:
    entries = measurement["after"]["manifest"]["entries"]
    active = [row for row in entries if "worktree" in row]
    def ordered(rows):
        return [row["number"] for row in rows] == list(range(1, 17))
    def isolated(rows):
        paths = [row["worktree"] for row in rows]
        return len(paths) == len(set(paths))
    assert ordered(entries) and isolated(active)
    assert not ordered(list(reversed(entries)))
    reused = [dict(row) for row in active]
    reused[1]["worktree"] = reused[0]["worktree"]
    assert not isolated(reused)
print("Measured output-order and per-entry-worktree audits pass; reversed order and shared-path mutants rejected")
