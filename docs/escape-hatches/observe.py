"""Run main's compiler and the independent Node source oracle; write all output to logs."""
import json
from pathlib import Path
import subprocess
import sys

destination = Path(sys.argv[1])
destination.mkdir(parents=True, exist_ok=True)
results = {}
for source in sorted(Path("docs/escape-hatches").glob("*.a")):
    name = source.stem
    def run(arguments):
        result = subprocess.run(arguments, text=True, capture_output=True, timeout=120)
        return {"code": result.returncode, "stdout": result.stdout, "stderr": result.stderr}
    c = run(["go", "run", "./cmd/adamic", "c", str(source)])
    (destination / (name + ".c")).write_text(c["stdout"])
    results[name] = {"c": c}
    typescript = destination / (name + ".ts")
    typescript.write_text(source.read_text())
    results[name]["node"] = run(["node", str(typescript)])
    if c["code"] == 0:
        binary = destination / (name + ".native")
        results[name]["build"] = run(["go", "run", "./cmd/adamic", "build", str(source), "-o", str(binary)])
        if results[name]["build"]["code"] == 0:
            results[name]["native"] = run([str(binary)])
    # Generated C lives beside the log; do not embed it in the JSON.
    if c["code"] == 0:
        c["stdout"] = "[generated C]"
    (destination / "observations.json").write_text(json.dumps(results, indent=2) + "\n")
