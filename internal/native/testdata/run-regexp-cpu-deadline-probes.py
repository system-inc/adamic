#!/usr/bin/env python3
"""Prove hang rejection, CPU-budget rejection, and a green loaded-box run."""
import json
import os
from pathlib import Path
import subprocess
import tempfile

root = Path(__file__).resolve().parents[3]
logs = Path(tempfile.mkdtemp(prefix="regexp-cpu-deadlines-"))
environment = dict(os.environ, GOFLAGS="-buildvcs=false")

def run(name, arguments, expected, witness, package="./internal/native"):
    with (logs / (name + ".log")).open("w") as output:
        result = subprocess.run(["go", "test", *arguments, package, "-count=1", "-v", "-timeout=10m"], cwd=root, env=environment, stdout=output, stderr=subprocess.STDOUT)
    text = (logs / (name + ".log")).read_text()
    if (result.returncode == 0) != expected or witness not in text:
        raise RuntimeError(f"{name}: unexpected outcome, see {logs / (name + '.log')}")
    print(f"{name}: {'passed' if expected else 'caught'}; {logs / (name + '.log')}", flush=True)

def overlay(name, path, before, after):
    original = root / path
    text = original.read_text()
    if text.count(before) != 1:
        raise RuntimeError(f"{name}: mutation anchor is not unique")
    changed = logs / (name + ".go")
    changed.write_text(text.replace(before, after))
    mapping = logs / (name + ".json")
    mapping.write_text(json.dumps({"Replace": {str(original): str(changed)}}))
    return ["-overlay=" + str(mapping)]

# Use the same wall-cap implementation with a short demonstration cap, avoiding
# a five-minute hang in every proof run. The production cap remains five minutes.
original = root / "internal/native/regexp_test.go"
changed = original.read_text().replace(
    "adamic_start(argc,argv);adamic_regex_set_step_limit(1000);",
    "adamic_start(argc,argv);for (;;) {} adamic_regex_set_step_limit(1000);")
changed = changed.replace("5*time.Minute, 10*time.Second", "250*time.Millisecond, 10*time.Second")
if changed == original.read_text():
    raise RuntimeError("hang mutation did not change the fixture")
(logs / "hang.go").write_text(changed)
(logs / "hang.json").write_text(json.dumps({"Replace": {str(original): str(logs / "hang.go")}}))
run("hang", ["-overlay=" + str(logs / "hang.json"), "-run=^TestRegExpNativeStepLimit$"], False, "wall cap 250ms exceeded")
run("cpu-check-disabled", overlay("cpu-check-disabled", "internal/native/regexp_cpu_test.go", "if cpu > cpuBudget {", "if false {" ) + ["-run=^TestRegExpChildCPUBudget$"], False, "CPU overrun escaped budget")

# Prove the area timing gate rejects a genuinely nonreturning compiled fixture.
original = root / "internal/oracle/regexp_native_fixes_test.go"
text = original.read_text()
anchor = "native.Build(native.C(program), binary, native.Options{})"
if text.count(anchor) != 1:
    raise RuntimeError("area hang mutation anchor is not unique")
hang_source = '#include "adamic.h"\nint main(int argc, char **argv) { adamic_start(argc, argv); for (;;) {} }\n'
text = text.replace(anchor, "native.Build(native.C(program)[:0] + " + json.dumps(hang_source) + ", binary, native.Options{Sanitize: true})")
text = text.replace("nil, 5*time.Minute, 3*time.Second", "nil, 250*time.Millisecond, 3*time.Second")
(logs / "area-hang.go").write_text(text)
(logs / "area-hang.json").write_text(json.dumps({"Replace": {str(original): str(logs / "area-hang.go")}}))
run("area-hang", ["-overlay=" + str(logs / "area-hang.json"), "-run=^TestRegExpNativeTiming$/class_string_duplicates"], False, "wall cap 250ms exceeded", "./internal/oracle")
run("area-cpu-check-disabled", overlay("area-cpu-check-disabled", "internal/oracle/regexp_cpu_test.go", "if cpuBudget > 0 && cpu > cpuBudget {", "if false {") + ["-run=^TestRegExpOracleCPUBudget$"], False, "CPU overrun escaped budget", "./internal/oracle")

# Hogs have no subprocesses; finally kills and reaps exactly those we spawned.
hogs = [subprocess.Popen(["python3", "-c", "while True: pass"], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL) for _ in range(8)]
try:
    print("loaded-box: eight busy loops running", flush=True)
    run("loaded-box", ["-run=^TestRegExpNativeStepLimit$"], True, "CPU=")
    run("area-loaded-box", ["-run=^TestRegExp(NativeTiming|LongBacktrackNode)$"], True, "CPU", "./internal/oracle")
finally:
    for hog in hogs:
        hog.kill()
    for hog in hogs:
        hog.wait()
print("proof logs: " + str(logs), flush=True)
