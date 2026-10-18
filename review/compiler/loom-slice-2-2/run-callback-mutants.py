"""Run callback adaptation omissions with Go overlays; sources stay in /tmp."""
import json
import os
from pathlib import Path
import subprocess
import tempfile
import sys

root = Path(__file__).resolve().parents[3]
os.chdir(root)
os.sched_setaffinity(0, sorted(os.sched_getaffinity(0))[:4])
environment = dict(os.environ, GOFLAGS=os.environ.get("GOFLAGS", "-buildvcs=false -trimpath"), GOMAXPROCS="4")
scratch = root / "review/compiler/loom-slice-2-2/callback-mutants"
scratch.mkdir(exist_ok=True)
recorded = root / 'review/compiler/loom-slice-2-2/callback-originals'
mapping = json.loads((recorded / 'mapping.json').read_text())
first = {path: (recorded / name).read_text() for path, name in mapping.items()}
adapter = Path("internal/lower/callback_widening.go").read_text()
assert "if !changed {" in adapter
targets = Path("internal/ir/call_targets.go").read_text()
assert "value = effects.Result" in targets
mutants = [
    ("refusal-revert", first, "./internal/oracle", "^TestCallbackWideningRefuse", 8),
    ("adaptation-omitted", {"internal/lower/callback_widening.go": adapter.replace("if !changed {", "if true || !changed {")}, "./internal/oracle", "^TestCallbackWideningNumber(Map|Filter|ForEach|Find|Some|Every|Reduce|Sort|GenericMap)$", 9),
    ("nullable-token-omitted", {"internal/lower/callback_widening.go": adapter.replace("if nullPointer {", "if false && nullPointer {")}, "./internal/oracle", "^TestCallbackWideningNullableObjectMap(Release|Sanitized|JavaScript)$", None),
    ("unrepresented-parameter-admitted", {"internal/lower/callback_widening.go": adapter.replace("if fit(ir.Read{Of: given}, of).Type() != of {", "if false && fit(ir.Read{Of: given}, of).Type() != of {")}, "./internal/oracle", "^TestCallbackWideningUnrepresentedParameter$", 1),
    ("adapter-target-omitted", {"internal/ir/call_targets.go": targets.replace("value = effects.Result", "_ = effects; break")}, "./internal/ir", "^TestCallbackAdapterTargets$", 1),
]
results = json.loads((scratch / "results.json").read_text()) if (scratch / "results.json").exists() else []
for name, files, package, pattern, failures in mutants:
    if len(sys.argv) > 1 and name != sys.argv[1]:
        continue
    if name == 'adaptation-omitted':
        pattern = "^TestCallbackWideningNumberEvery(Release|Sanitized|JavaScript)$"
        failures = None
    directory = scratch / name
    directory.mkdir(exist_ok=True)
    replacements = {}
    for index, (path, source) in enumerate(files.items()):
        target = directory / (str(index) + ".go.txt")
        target.write_text(source)
        replacements[str((root / path).resolve())] = str(target)
    overlay = directory / "overlay.json"
    overlay.write_text(json.dumps({"Replace": replacements}))
    log = directory / "test.jsonl"
    command = ["go", "test", "-json", "-count=1", "-timeout=90s", "-overlay=" + str(overlay), package, "-run", pattern]
    with log.open("w") as output:
        ran = subprocess.run(command, env=environment, stdout=output, stderr=subprocess.STDOUT, timeout=300)
    events = [json.loads(line) for line in log.read_text().splitlines() if line.startswith("{")]
    failed = [event["Test"] for event in events if event["Action"] == "fail" and event.get("Test")]
    build_failed = any(event["Action"] == "build-fail" or event.get("FailedBuild") for event in events)
    caught = ran.returncode != 0 and (len(failed) == failures if failures is not None else len(failed) > 0) and not build_failed and "test timed out" not in log.read_text()
    row = {"mutant": name, "caught": caught, "exit": ran.returncode, "catchers": failed, "command": command}
    results = [r for r in results if r["mutant"] != name] + [row]
    print(json.dumps(row), flush=True)
    (scratch / "results.json").write_text(json.dumps(results, indent=2) + "\n")
    assert caught, (name, str(log))
print("All executed callback mutants caught; evidence:", scratch)
