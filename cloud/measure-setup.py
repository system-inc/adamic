#!/usr/bin/env python3
"""Serial, alternating-order setup trials on one box and one checkout revision.

Warm trials reuse this checkout and cache. Cache-cold trials use empty Go caches.
All-cold trials also use fresh local parent checkouts, empty tools and submodules.
Logs preserve subprocess output, timestamps and build flags. Nothing is piped.
"""
import argparse
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import time

parser = argparse.ArgumentParser()
parser.add_argument("--before", default="e011f8f60899586d6373a5ccb07335ad82cfbf3c")
parser.add_argument("--cold", choices=["cache", "all"])
parser.add_argument("--loops", type=int, default=3)
parser.add_argument("--output", required=True)
arguments = parser.parse_args()
repository = Path(__file__).resolve().parent.parent
output = Path(arguments.output).resolve()
output.mkdir(parents=True, exist_ok=True)
scratch = Path(tempfile.mkdtemp(prefix="setup-trials-", dir="/workspace"))
revision = subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=repository).decode().strip()
baseline = subprocess.check_output(["git", "show", f"{arguments.before}:cloud/setup.sh"],
                                   cwd=repository).decode()
results_file = output / "timings.json"
results = json.loads(results_file.read_text()) if results_file.exists() else []
assert all(entry["before"]["commit"] == revision and entry["cold"] == arguments.cold
           for entry in results), "resumed trials must have the same commit and cold mode"


def captured(command, cwd, environment):
    try:
        return subprocess.check_output(command, cwd=cwd, env=environment,
                                       stderr=subprocess.DEVNULL).decode().strip()
    except (subprocess.CalledProcessError, FileNotFoundError):
        return "unavailable"


def flags(cwd, environment):
    return dict(commit=revision, nproc=captured(["nproc"], cwd, environment),
                cpu_max=Path("/sys/fs/cgroup/cpu.max").read_text().strip(),
                go=captured(["go", "version"], cwd, environment),
                clang=captured(["clang", "--version"], cwd, environment).splitlines()[0],
                node=captured(["node", "--version"], cwd, environment),
                load=Path("/proc/loadavg").read_text().strip(),
                cached="action-cache" if not arguments.cold else "empty-action-cache")


for loop in range(1, arguments.loops + 1):
    for mode in (["before", "after"] if loop % 2 else ["after", "before"]):
        if any(entry["loop"] == loop and entry["mode"] == mode and entry["exit"] == 0
               for entry in results):
            continue
        trial = scratch / f"{mode}-{loop}"
        trial.mkdir()
        cwd = repository
        environment = os.environ.copy()
        environment.pop("ADAMIC_GATE_UNCACHED", None)
        environment["PS4"] = '+ ${EPOCHREALTIME} ${BASH_SOURCE}:${LINENO}: '
        if arguments.cold:
            environment["GOCACHE"] = str(trial / "cache")
        if arguments.cold == "all":
            cwd = trial / "repository"
            with (output / f"{mode}-{loop}-checkout.log").open("wb") as log:
                subprocess.run(["git", "clone", "--shared", "--no-checkout", str(repository), str(cwd)],
                               stdout=log, stderr=log, check=True)
                subprocess.run(["git", "checkout", "--detach", revision], cwd=cwd,
                               stdout=log, stderr=log, check=True)
            environment["ADAMIC_TOOLS"] = str(trial / "tools")
            environment["PATH"] = "/usr/local/bin:/usr/bin:/bin"
            environment["GOPATH"] = str(trial / "gopath")
        script = cwd / "cloud/setup.sh"
        if mode == "before":
            script = trial / "before.sh"
            script.write_text(baseline.replace(
                'repository=$(cd "$(dirname "$0")/.." && pwd)',
                'repository=' + "'" + str(cwd).replace("'", "'\\''") + "'"))
        command = ["bash", "-x", str(script)]
        before = flags(cwd, environment)
        log_path = output / f"{mode}-{loop}.log"
        start = time.monotonic()
        with log_path.open("wb") as log:
            log.write(("build-flags-before: " + json.dumps(before) + "\n").encode())
            log.flush()
            result = subprocess.run(command, cwd=cwd, env=environment, stdout=log, stderr=log)
        elapsed = time.monotonic() - start
        if arguments.cold == "all":
            environment["PATH"] = str(trial / "tools/bin") + ":" + str(trial / "tools/go/bin") + ":" + environment["PATH"]
        after = flags(cwd, environment)
        entry = dict(loop=loop, mode=mode, cold=arguments.cold, seconds=elapsed,
                     instrument=command, before=before, after=after, exit=result.returncode,
                     log=str(log_path), scratch=str(trial))
        results.append(entry)
        with log_path.open("a") as log:
            log.write("build-flags-after: " + json.dumps(after) + "\n")
            log.write(f"wall={elapsed:.6f} exit={result.returncode}\n")
        (output / "timings.json").write_text(json.dumps(results, indent=2) + "\n")
        print(f"loop={loop} {mode} wall={elapsed:.3f}s exit={result.returncode} log={log_path}", flush=True)
        if arguments.cold == "all":
            # Remove only this runner's named trial after saving its evidence. A tools install
            # may add a bashrc source line; remove exactly that line before removing the tools.
            rc = Path.home() / ".bashrc"
            added = f"source {trial}/tools/env.sh\n"
            if rc.exists():
                text = rc.read_text()
                if added in text:
                    rc.write_text(text.replace(added, ""))
            assert trial.parent == scratch and trial.name == f"{mode}-{loop}"
            # Go deliberately makes downloaded module directories read-only. They are all
            # inside this runner's asserted scratch trial; restore directory write permission.
            for directory, _, _ in os.walk(trial):
                os.chmod(directory, 0o700)
            shutil.rmtree(trial)
        if result.returncode:
            raise SystemExit(result.returncode)
    if arguments.cold and any(entry["seconds"] > 300 for entry in results):
        print("A cold run exceeded five minutes; using the requested one-pair exception.")
        break
