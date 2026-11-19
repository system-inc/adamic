"""Run the two requested entry passes serially and publish each result."""

import gzip
import hashlib
import json
import os
from pathlib import Path
import subprocess
import time


root = Path("/tmp/adamic-g02")
repo = root / "evidence"
evidence = Path(__file__).resolve().parent
binary = root / "entry-census"
branch = "refs/heads/typescript/family5-c02-sites-evidence"
job_env = {key: value for key, value in os.environ.items() if not key.startswith("LATENT_")}
job_env.update(
    LATENT_FULL="0",
    LATENT_ASSERT_NO_OUTPUT="1",
    LATENT_TAIL="1",
    LATENT_TIMES="/dev/stderr",
    CENSUS_MODULE_FORMAT="CommonJSBundle",
)


def git(*args):
    subprocess.run(["git", "-C", repo, *args], check=True)


def compress(path, destination):
    with path.open("rb") as source, destination.open("wb") as output:
        with gzip.GzipFile(filename="", mode="wb", fileobj=output, mtime=0) as compressed:
            while chunk := source.read(1024 * 1024):
                compressed.write(chunk)


def run(side):
    work = root / ("entry-" + side)
    work.mkdir()
    destination = evidence / "entry" / side
    destination.mkdir(parents=True, exist_ok=True)
    stream = work / "latent.jsonl"
    cmd = [binary, root / side / "src/tsc/tsc.ts", stream]
    started = time.time()
    print(f"START {side} at {started}", flush=True)
    with (work / "stdout.log").open("wb") as stdout, (work / "stderr.log").open("wb") as stderr:
        child = subprocess.run(cmd, cwd=root / "main", env=job_env, stdout=stdout, stderr=stderr)
    status = {
        "command": list(map(str, cmd)),
        "cwd": str(root / "main"),
        "environment": {key: value for key, value in job_env.items() if key.startswith(("LATENT_", "CENSUS_"))},
        "adamic": "491e020de2aa138d7b608611c40a641571a5a1eb",
        "binary_sha256": hashlib.file_digest(binary.open("rb"), "sha256").hexdigest(),
        "started_unix": started,
        "elapsed_seconds": round(time.time() - started, 3),
        "exit_code": child.returncode,
    }
    print(f"EXIT {side} {child.returncode}", flush=True)
    for filename in ("latent.jsonl", "stdout.log", "stderr.log"):
        path = work / filename
        if path.is_file():
            compress(path, destination / (filename + ".gz"))
    (destination / "status.json").write_text(json.dumps(status, indent=2) + "\n")
    git("add", str(destination))
    git("commit", "-m", f"stage3: publish G02 {side} first-error tsc entry stream")
    git("push", "origin", "HEAD:" + branch)
    return child.returncode


exits = [run(side) for side in ("base", "edit")]
raise SystemExit(int(any(exits)))
