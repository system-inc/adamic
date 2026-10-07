"""Run from the repository root after sourcing the toolchain environment."""
import json
import subprocess
import sys
from pathlib import Path

output = Path(sys.argv[1])
output.mkdir(parents=True, exist_ok=True)
results = []
failures = []
programs = sorted(Path("internal/oracle/testdata").glob("records_coverage_*.a"))
programs += sorted(Path("notes/records-lowering").glob("*.a"))

def run(arguments):
    result = subprocess.run(arguments, capture_output=True, text=True)
    return {"code": result.returncode, "stdout": result.stdout, "stderr": result.stderr}

for program in programs:
    node = run(["node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", str(program.resolve())])
    binary = output / program.stem
    build = run(["go", "run", "./cmd/adamic", "build", str(program), "-o", str(binary)])
    native = run([str(binary)]) if build["code"] == 0 else None
    emitted = run(["go", "run", "./cmd/adamic", "js", str(program)])
    javascript = None
    if emitted["code"] == 0:
        script = output / (program.stem + ".mjs")
        script.write_text(emitted["stdout"])
        javascript = run(["node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", str(script)])
    if program.name.startswith("records_coverage_"):
        passed = native is not None and node == native == javascript
    elif program.name.startswith("dynamic_"):
        passed = native is not None and native == javascript and native["code"] == 70
        passed = passed and "records hold own keys only" in native["stderr"]
    else:
        passed = build["code"] != 0 and emitted["code"] != 0
    if not passed:
        failures.append(str(program))
    results.append({"program": str(program), "node": node, "build": build, "native": native, "javascript": javascript, "passed": passed})
    (output / "results.json").write_text(json.dumps(results, indent=2))
    print(str(program), "PASS" if passed else "FAIL", flush=True)
if failures:
    raise SystemExit("Unexpected observations: " + ", ".join(failures))
