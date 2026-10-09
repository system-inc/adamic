#!/usr/bin/env python3
"""Bounded, sequential measurements on four workers; no compiler observations cached."""
import json
import os
from pathlib import Path
import subprocess
import tempfile
import time

ROOT = Path(__file__).resolve().parents[4]
OUT = ROOT / "review/compiler/admission-delta/measurements"
BASE = "2b1be38362046455e0e5454d8b0e674a8630e95d"
HEAD = "ced32bf9b1f410b8be07f078adb78365a20a9f5b"
GENERATOR = "0fd9b945e15525b1dc25a840a40324bdbf0c31d0"
SMALL_PATHS = ["internal/lower/testdata/array_narrowing/" + name + ".a" for name in ["cat", "every", "find"]]


def git(*args, env=None):
    return subprocess.check_output(["git", *args], cwd=ROOT, env=env, timeout=60).decode().strip()


def small_revision():
    # Keep the fixes compiler exactly; restore all other .a inputs to base.
    # Only Git's temporary index changes, never the delivery checkout.
    with tempfile.TemporaryDirectory(prefix="admission-index-", dir="/workspace") as directory:
        env = os.environ.copy()
        env["GIT_INDEX_FILE"] = directory + "/index"
        git("read-tree", HEAD, env=env)
        changed = git("diff", "--name-only", "--no-renames", BASE, HEAD, "--", "*.a").splitlines()
        for path in changed:
            if path in SMALL_PATHS:
                continue
            listing = git("ls-tree", BASE, "--", path)
            if listing:
                metadata, _ = listing.split("\t", 1)
                mode, _, blob = metadata.split()
                git("update-index", "--add", "--cacheinfo", mode, blob, path, env=env)
            else:
                git("update-index", "--force-remove", path, env=env)
        tree = git("write-tree", env=env)
        sha = git("commit-tree", tree, "-p", BASE, "-m", "Measure the real fixes compiler with three changed Adamic inputs.")
    assert git("diff", "--name-only", "--diff-filter=AM", "--no-renames", BASE, sha, "--", "*.a").splitlines() == SMALL_PATHS
    (OUT / "small.patch").write_text(git("diff", "--binary", BASE, sha) + "\n")
    (OUT / "small.commit.txt").write_text(git("cat-file", "commit", sha) + "\n")
    return sha


def manifest(sha, name):
    path = OUT / (name + ".manifest.json")
    with path.open("w") as output, (OUT / (name + ".manifest.log")).open("w") as errors:
        subprocess.run(["python3", "cloud/admission-corpus/manifest.py", "--sha", sha], cwd=ROOT, stdout=output, stderr=errors, timeout=60, check=True)
    return path


def measure(name, head, pinned_manifest, cache, extra):
    command = ["/workspace/admission-measure-tool", "--base", BASE, "--head", head,
               "--manifest", str(pinned_manifest), "--manifest-generator-revision", GENERATOR,
               "--workers", "4", "--budget", "60", "--json", *extra]
    env = os.environ.copy()
    env.pop("ADAMIC_BUILD_CACHE", None)
    env["ADAMIC_BUILD_CACHE_DIR"] = cache
    env["ADAMIC_BUILD_LOG"] = str(OUT / (name + ".build-cache.log"))
    env["TMPDIR"] = "/workspace/admission-speed-tmp"
    before = len(list(Path(cache).iterdir()))
    print("starting", name, "cache entries", before, flush=True)
    started = time.monotonic()
    with (OUT / (name + ".json")).open("w") as output, (OUT / (name + ".log")).open("w") as errors:
        result = subprocess.run(command, cwd=ROOT, env=env, stdout=output, stderr=errors, timeout=600)
    wall = time.monotonic() - started
    record = dict(command=command, cache_directory=cache, cache_entries_before=before, exit=result.returncode, wall_seconds=wall)
    (OUT / (name + ".status.json")).write_text(json.dumps(record, indent=2) + "\n")
    observed = json.loads((OUT / (name + ".json")).read_text())
    assert result.returncode == 0 and observed["verdict"] == "partial"
    assert not any(p["class"].startswith("compiler-") for p in observed["programs"])
    print("completed", name, "wall", round(wall, 3), "phases", observed["phase_seconds"], flush=True)


def main():
    small = small_revision()
    fx_cache = tempfile.mkdtemp(prefix="admission-measure-fx-cache-", dir="/workspace")
    small_cache = tempfile.mkdtemp(prefix="admission-measure-small-cache-", dir="/workspace")
    (OUT / "inputs.json").write_text(json.dumps(dict(base=BASE, head=HEAD, small_head=small, small_paths=SMALL_PATHS,
        generator=GENERATOR, workers=4, shards=3, fx_cache=fx_cache, small_cache=small_cache,
        cold_definition="Empty compiler-product cache; preexisting Go package cache and pinned Git objects retained."), indent=2) + "\n")
    fx_manifest = manifest(HEAD, "fx")
    small_manifest = manifest(small, "small")
    measure("fx-diff-cold", HEAD, fx_manifest, fx_cache, ["--corpus-filter", "diff"])
    measure("fx-diff-warm", HEAD, fx_manifest, fx_cache, ["--corpus-filter", "diff"])
    measure("small-diff-cold", small, small_manifest, small_cache, ["--corpus-filter", "diff"])
    measure("small-diff-warm", small, small_manifest, small_cache, ["--corpus-filter", "diff"])
    for index in range(1, 4):
        measure("shard-" + str(index) + "-warm", HEAD, fx_manifest, fx_cache, ["--shard", str(index) + "/3"])


if __name__ == "__main__":
    main()
