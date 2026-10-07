"""Compare original and bounded tools with a forever-running fake Go on PATH.
Run after building matching binaries in OUTPUT/bin-before and OUTPUT/bin-after.
Raw logs stay in OUTPUT; only proof.json is suitable for committed evidence.
"""
import argparse
import json
import os
from pathlib import Path
import signal
import subprocess
import time
from python import run as bounded_run

parser = argparse.ArgumentParser()
parser.add_argument("--repository", required=True)
parser.add_argument("--output", required=True)
args = parser.parse_args()
repo, output = Path(args.repository).resolve(), Path(args.output).resolve()
probe = output / "probe-repo"
probe.mkdir(exist_ok=True)
(probe / "cmd/adamic-gate").mkdir(parents=True, exist_ok=True)
(probe / "cmd/adamic-gate/timings.json").write_text("{}\n")
(probe / "program.a").write_text("console.log(1);\n")
if not (probe/".git").exists():
    for command in (["git", "init", "-q"], ["git", "add", "."],
                    ["git", "-c", "user.name=deadline-proof", "-c", "user.email=proof@example.invalid", "commit", "-qm", "deadline proof"]):
        bounded_run(command, cwd=probe, check=True, deadline=30)

fake = output / "fake-bin"
fake.mkdir(exist_ok=True)
(fake / "go").write_text("#!/bin/sh\necho $$ > \"$ADAMIC_PROOF_PID\"\ntrap '' TERM\n(sh -c 'trap \"\" TERM; while :; do echo tick >> \"$ADAMIC_PROOF_HEARTBEAT\"; sleep .01; done') &\nwait\n")
(fake / "go").chmod(0o755)
identity = {}
for name, command in [("commit", ["git", "rev-parse", "HEAD"]), ("nproc", ["nproc"]),
                      ("go", ["go", "version"]), ("clang", ["clang", "--version"]), ("node", ["node", "--version"])]:
    identity[name] = bounded_run(command, cwd=repo, text=True, stdout=subprocess.PIPE, check=True, deadline=30).stdout.strip().splitlines()[0]
identity.update(cpu_max=Path("/sys/fs/cgroup/cpu.max").read_text().strip(),
                build_flags="go build -buildvcs=false; GOFLAGS="+os.environ.get("GOFLAGS", ""), cache="ADAMIC_GATE_UNCACHED=1; Go build cache warm")
records = []
for tool in ["adamic-gate", "adamic-test262", "adamic-reduce", "adamic-fuzz", "catalog", "setup"]:
    for round_number in range(3):
        for variant in ["before", "after"]:
            label = f"{tool}-{round_number}-{variant}"
            heartbeat, pidfile = output / (label+".heartbeat"), output / (label+".pid")
            heartbeat.unlink(missing_ok=True)
            pidfile.unlink(missing_ok=True)
            env = dict(os.environ, PATH=str(fake)+":"+os.environ["PATH"], ADAMIC_CHILD_DEADLINE="0.2",
                       ADAMIC_PROOF_PID=str(pidfile), ADAMIC_PROOF_HEARTBEAT=str(heartbeat), ADAMIC_GATE_UNCACHED="1")
            binary = output / ("bin-"+variant) / tool
            root = repo if variant == "after" else output / "before"
            cwd = repo
            if tool == "adamic-gate":
                command, cwd = [str(binary), "plan"], probe
            elif tool == "adamic-test262":
                command = [str(binary), "-compiler-subprocess", "-test262", str(repo/"cmd/adamic-test262/testdata/mini"), "pass"]
            elif tool == "adamic-reduce":
                command = [str(binary), "-work", str(output/(label+"-work")), str(probe/"program.a")]
            elif tool == "adamic-fuzz":
                command = [str(binary), "-count", "1", "-work", str(output/(label+"-work"))]
            elif tool == "catalog":
                command = ["bash", str(root/"verify/catalog/check.sh"), "ac92bf0", "--entry", "1"]
            else:
                command = ["bash", str(root/"cloud/setup.sh")]
                env["ADAMIC_TOOLS"] = str(output/"setup-tools")
            start = time.monotonic()
            load_before = Path("/proc/loadavg").read_text().strip()
            watchdog = False
            with (output/(label+".log")).open("w") as log:
                process = subprocess.Popen(command, cwd=cwd, env=env, stdout=log, stderr=log, start_new_session=True)
                try:
                    process.wait(timeout=3 if variant == "after" else 1.5)
                except subprocess.TimeoutExpired:
                    watchdog = True
                    os.killpg(process.pid, signal.SIGKILL)
                    process.wait(timeout=2)
            seconds = time.monotonic()-start
            before = heartbeat.read_bytes() if heartbeat.exists() else b""
            time.sleep(.1)
            after = heartbeat.read_bytes() if heartbeat.exists() else b""
            diagnostic = (output/(label+".log")).read_text()
            # Cleanup even a deliberately broken group-kill mutant, after observation.
            try:
                os.killpg(process.pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
            if pidfile.exists():
                try:
                    os.killpg(int(pidfile.read_text()), signal.SIGKILL)
                except ProcessLookupError:
                    pass
            if variant == "before":
                assert watchdog and before, (label, diagnostic)
            else:
                assert not watchdog and before and before==after, (label, "child/group survived", diagnostic)
                assert "deadline" in diagnostic and (str(fake/"go") in diagnostic or "go" in diagnostic), (label, diagnostic)
            records.append(dict(identity, tool=tool, child="go", variant=variant, round=round_number,
                                seconds=seconds, watchdog=watchdog, descendant_stopped=before==after,
                                proof_deadline_seconds=.2, instrument=command, load_before=load_before,
                                load_after=Path("/proc/loadavg").read_text().strip()))
            print(label, round(seconds, 3), "watchdog" if watchdog else "killed", flush=True)
            (output/"proof.json").write_text(json.dumps(records, indent=2)+"\n")
