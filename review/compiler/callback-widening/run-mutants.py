"""Run callback adaptation omissions with Go overlays; sources stay in /tmp."""
import json
import os
from pathlib import Path
import subprocess
import tempfile

root = Path(__file__).resolve().parents[3]
os.chdir(root)
os.sched_setaffinity(0, sorted(os.sched_getaffinity(0))[:4])
environment = dict(os.environ, GOFLAGS="-buildvcs=false", GOMAXPROCS="4")
scratch = Path(tempfile.mkdtemp(prefix="callback-widening-mutants-", dir="/tmp/adamic-gate"))
refusal = "9836491bd"

def historical(path):
    return subprocess.check_output(["git", "show", refusal + ":" + path], text=True)

original = historical("internal/lower/callback_widening.go")
assert "if known && takes == ir.Union" in original
first = {
    "internal/lower/callback_widening.go": original.replace("if known && takes == ir.Union", "if false && known && takes == ir.Union"),
    "internal/lower/object.go": historical("internal/lower/object.go"),
    "internal/lower/library_array.go": historical("internal/lower/library_array.go"),
    "internal/ir/call_targets.go": historical("internal/ir/call_targets.go"),
    "internal/oracle/callback_widening_test.go": historical("internal/oracle/callback_widening_test.go").replace("../../review/compiler/callback-widening/fixtures/", "testdata/callback_widening/"),
}
adapter = Path("internal/lower/callback_widening.go").read_text()
assert "if !changed {" in adapter
targets = Path("internal/ir/call_targets.go").read_text()
assert "value = effects.Result" in targets
mutants = [
    ("refusal-revert", first, "./internal/oracle", "^TestCallbackWideningRefuse", 8),
    ("adaptation-omitted", {"internal/lower/callback_widening.go": adapter.replace("if !changed {", "if true || !changed {")}, "./internal/oracle", "^TestCallbackWideningNumber(Map|Filter|ForEach|Find|Some|Every|Reduce|Sort|GenericMap)$", 9),
    ("nullable-token-omitted", {"internal/lower/callback_widening.go": adapter.replace("if nullPointer {", "if false && nullPointer {")}, "./internal/oracle", "^TestCallbackWideningNullableObjectMap$", 1),
    ("unrepresented-parameter-admitted", {"internal/lower/callback_widening.go": adapter.replace("if fit(ir.Read{Of: given}, of).Type() != of {", "if false && fit(ir.Read{Of: given}, of).Type() != of {")}, "./internal/oracle", "^TestCallbackWideningUnrepresentedParameter$", 1),
    ("adapter-target-omitted", {"internal/ir/call_targets.go": targets.replace("value = effects.Result", "_ = effects; break")}, "./internal/ir", "^TestCallbackAdapterTargets$", 1),
]
results = []
for name, files, package, pattern, failures in mutants:
    directory = scratch / name
    directory.mkdir()
    replacements = {}
    for index, (path, source) in enumerate(files.items()):
        target = directory / (str(index) + ".go.txt")
        target.write_text(source)
        replacements[str((root / path).resolve())] = str(target)
    overlay = directory / "overlay.json"
    overlay.write_text(json.dumps({"Replace": replacements}))
    log = directory / "test.jsonl"
    command = ["go", "test", "-json", "-count=1", "-timeout=5m", "-overlay=" + str(overlay), package, "-run", pattern]
    with log.open("w") as output:
        ran = subprocess.run(command, env=environment, stdout=output, stderr=subprocess.STDOUT)
    events = [json.loads(line) for line in log.read_text().splitlines() if line.startswith("{")]
    failed = [event["Test"] for event in events if event["Action"] == "fail" and event.get("Test")]
    build_failed = any(event["Action"] == "build-fail" or event.get("FailedBuild") for event in events)
    caught = ran.returncode != 0 and len(failed) == failures and not build_failed
    row = {"mutant": name, "caught": caught, "exit": ran.returncode, "catchers": failed, "command": command}
    results.append(row)
    print(json.dumps(row), flush=True)
    (scratch / "results.json").write_text(json.dumps(results, indent=2) + "\n")
    assert caught, (name, str(log))
print("All five mutants caught; evidence:", scratch)
