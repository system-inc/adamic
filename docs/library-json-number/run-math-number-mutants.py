"""Run integration's four Math/Number mutants, requiring only Node to catch them."""
from pathlib import Path
import json
import math
import os
import subprocess

root = Path(__file__).resolve().parents[2]
runtime = root / "internal/native/runtime/library_math_number.c"
scratch = Path("/tmp/library-json-number-math-mutants")
scratch.mkdir(exist_ok=True)
original = runtime.read_bytes()
negative_boundary = -3.4028235677973362e38
mutated_boundary = math.nextafter(negative_boundary, -math.inf)
assert mutated_boundary == -3.4028235677973366e38
mutants = [
    ("fround-negative-ulp", "math_edges", "value >= -3.4028235677973362e38", "value >= -3.4028235677973366e38"),
    ("epsilon-name", "own_names", '"EPSILON"', '"Epsilon"'),
    ("octal-binary-prefixes", "convert_sources", "start[1] == 'o' || start[1] == 'O' ? 8 : start[1] == 'b' || start[1] == 'B' ? 2 : 0", "0"),
    ("null-string-zero", "convert_sources", "if (text == NULL) return NAN;", "if (text == NULL) return 0;"),
]
results = []
try:
    for name, fixture, before, after in mutants:
        source = original.decode()
        if source.count(before) != 1:
            raise RuntimeError(f"{name}: expected exactly one mutation site")
        runtime.write_text(source.replace(before, after))
        log = scratch / f"{name}.log"
        command = ["go", "test", "-v", "-count=1", "-timeout", "30m", "./internal/oracle", "-run", f"TestNativeAgreesWithNode/internal/oracle/testdata/library_number_{fixture}.a"]
        with log.open("w") as output:
            result = subprocess.run(command, cwd=root, env={**os.environ, "ADAMIC_GATE_UNCACHED": "1"}, stdout=output, stderr=subprocess.STDOUT)
        text = log.read_text()
        selected = f"--- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/library_number_{fixture}.a" in text
        only_node = ("stdout differs" in text and "node:   exit 0," in text and "native: exit 0," in text
                     and text.count('stderr ""') == 2
                     and all(marker not in text for marker in ["[build failed]", "runtime error:", "leaks:", "the release build:", "JavaScript backend:", "no tests to run"]))
        killed = result.returncode != 0 and selected and only_node
        results.append({"mutant": name, "fixture": fixture, "exit": result.returncode, "caughtBy": "Node stdout comparison only", "killed": killed, "log": str(log)})
        runtime.write_bytes(original)
        if not killed:
            raise RuntimeError(f"{name}: not caught solely by Node's comparison; see {log}")
finally:
    runtime.write_bytes(original)
    (scratch / "results.json").write_text(json.dumps(results, indent=2) + "\n")
print(json.dumps(results, indent=2))
