"""Run the three requested probe-D mutants and restore the runtime after each."""
from pathlib import Path
import json
import subprocess

root = Path(__file__).resolve().parents[2]
runtime = root / "internal/native/runtime/json_stringify.c"
scratch = Path("/tmp/library-json-number-shapes-mutants")
scratch.mkdir(exist_ok=True)
original = runtime.read_bytes()
mutants = [
    ("replacer-null", " || value.reference == NULL) { return; }", ") { return; }", "runtime error:"),
    ("array-null", " || kind == adamic_json_array) && value.reference == NULL", ") && value.reference == NULL", "runtime error:"),
    ("union-closure", "case adamic_kind_closure: kind = adamic_json_function; break;", "case adamic_kind_closure: kind = adamic_json_map; break;", "stdout differs"),
]
results = []
try:
    for name, before, after, expected in mutants:
        source = original.decode()
        if source.count(before) != 1:
            raise RuntimeError(f"{name}: expected exactly one mutation site")
        runtime.write_text(source.replace(before, after))
        log = scratch / f"{name}.log"
        command = ["go", "test", "-v", "-count=1", "-timeout", "30m", "./internal/oracle", "-run", "TestNativeAgreesWithNode/internal/oracle/testdata/library_json_shapes.a"]
        with log.open("w") as output:
            result = subprocess.run(command, cwd=root, stdout=output, stderr=subprocess.STDOUT)
        text = log.read_text()
        killed = result.returncode != 0 and expected in text and "[build failed]" not in text
        results.append({"mutant": name, "exit": result.returncode, "caughtBy": expected, "killed": killed, "log": str(log)})
        runtime.write_bytes(original)
        if not killed:
            raise RuntimeError(f"{name}: mutant not caught by the expected check; see {log}")
finally:
    runtime.write_bytes(original)
    (scratch / "results.json").write_text(json.dumps(results, indent=2) + "\n")
print(json.dumps(results, indent=2))
