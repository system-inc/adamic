#!/usr/bin/env python3
import json
import os
from pathlib import Path
import signal
import subprocess
import tempfile
import time

ROOT = Path(__file__).resolve().parents[4]
OUT = ROOT / "review/compiler/admission-delta/yaml-evidence"
SCRATCH = Path(tempfile.mkdtemp(prefix="yaml-evidence-", dir="/workspace"))
ENV = os.environ.copy()
ENV["GOMAXPROCS"] = "4"
ENV["TMPDIR"] = str(SCRATCH / "tmp")
(SCRATCH / "tmp").mkdir()
ENV["GOCACHE"] = subprocess.check_output(["go", "env", "GOCACHE"], env=ENV, text=True, timeout=30).strip()
ENV.pop("ADAMIC_BUILD_CACHE", None)
RECORDS = []


def command(args, cwd, limit, name, env=ENV):
    start = time.monotonic()
    with (OUT / (name + ".jsonl")).open("w") as output, (OUT / (name + ".stderr.log")).open("w") as errors:
        process = subprocess.Popen(args, cwd=cwd, env=env, stdout=output, stderr=errors, start_new_session=True)
        try:
            status = process.wait(timeout=limit)
        except subprocess.TimeoutExpired:
            os.killpg(process.pid, signal.SIGKILL)
            status = process.wait()
    row = dict(name=name, command=args, cwd=str(cwd), limit_seconds=limit, exit=status, wall_seconds=time.monotonic()-start)
    RECORDS.append(row)
    (OUT / "status.json").write_text(json.dumps(RECORDS, indent=2)+"\n")
    print(name, "exit", status, "wall", round(row["wall_seconds"], 3), flush=True)
    return status


def git(repo, *args):
    return subprocess.check_output(["git", "-C", str(repo), *args], env=ENV, text=True, timeout=120).strip()


def prepare(name, sha):
    tree = SCRATCH / name
    git(ROOT, "worktree", "add", "--detach", str(tree), sha)
    cohere_sha = git(ROOT, "rev-parse", sha+":cohere")
    git(ROOT/"cohere", "worktree", "add", "--detach", str(tree/"cohere"), cohere_sha)
    ts_sha = git(tree/"cohere", "rev-parse", "HEAD:TypeScript")
    git(ROOT/"cohere/TypeScript", "worktree", "add", "--detach", str(tree/"cohere/TypeScript"), ts_sha)
    command(["npm", "ci", "--ignore-scripts"], tree/"stage3/api", 90, name+"-node-provision")
    env = ENV.copy()
    cache = SCRATCH / (name+"-cache")
    cache.mkdir()
    env["ADAMIC_BUILD_CACHE_DIR"] = str(cache/"products")
    env["XDG_CACHE_HOME"] = str(cache/"native")
    (cache/"native").mkdir()
    env["ADAMIC_BUILD_LOG"] = str(OUT/(name+"-build-cache.log"))
    command(["go", "test", "./stage1/cohere/yaml", "-run", "^$", "-count=1", "-timeout", "300s", "-json"], tree, 420, name+"-go-prime", env)
    return tree, env


def main():
    head = git(ROOT, "rev-parse", "84915995^{commit}")
    main_tip = git(ROOT, "rev-parse", "origin/main")
    base = git(ROOT, "merge-base", head, main_tip)
    (OUT/"inputs.json").write_text(json.dumps(dict(head=head, main_tip=main_tip, base=base, scratch=str(SCRATCH),
        gocache=ENV["GOCACHE"], gomaxprocs=4, cpu_max=Path('/sys/fs/cgroup/cpu.max').read_text().strip(),
        cold="Fresh per-revision build-product and native XDG caches; Go compilation primed separately.",
        yaml_library_configured=bool(ENV.get('ADAMIC_YAML_LIBRARY'))), indent=2)+"\n")
    for name, sha in [("refusals", head), ("main-base", base)]:
        print("prepare", name, sha, flush=True)
        tree, env = prepare(name, sha)
        args=["go", "test", "./stage1/cohere/yaml", "-count=1", "-timeout", "300s", "-json"]
        status=command(args, tree, 420, name+"-cold", env)
        text=(OUT/(name+"-cold.jsonl")).read_text()
        if status != 0 and ("file-driver setup" in text or "test timed out" in text):
            warm_env=env.copy()
            warm_env["ADAMIC_FILE_DRIVER_SETUP_CHILD"]="1"
            warm_env["ADAMIC_FILE_DRIVER_STATE"]=str(SCRATCH/(name+"-prime-state.json"))
            command(["go", "test", "./stage1/cohere/yaml", "-run", "^TestFileDriver_Setup$", "-count=1", "-timeout", "300s", "-json"], tree, 420, name+"-file-driver-prime", warm_env)
        command(args, tree, 420, name+"-warm", env)


if __name__ == "__main__":
    main()
