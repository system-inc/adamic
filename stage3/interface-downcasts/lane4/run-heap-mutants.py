#!/usr/bin/env python3
"""Test isolated runtime normalization defects, restoring production each time."""
import json, pathlib, subprocess
repo = pathlib.Path(__file__).resolve().parents[3]
lane = pathlib.Path(__file__).resolve().parent
source = repo / "internal/native/runtime/view_unions_mixed.c"
original = source.read_bytes()
mutants = {
    "number-tag": ("snapshot.kind = adamic_view_union_number;", "snapshot.kind = adamic_view_union_boolean;"),
    "boolean-payload": ("snapshot.payload.boolean = ((const adamic_boolean_box *)value)->boolean;", "snapshot.payload.boolean = false;"),
}
results = []
try:
    for name, (before, after) in mutants.items():
        text = original.decode()
        assert text.count(before) == 1
        source.write_text(text.replace(before, after))
        log = lane / "logs" / ("heap-mutant-" + name + ".log")
        with log.open("w") as output:
            result = subprocess.run(["go", "test", "./internal/oracle", "-run", "^TestCheckedViewPrimitiveHeap$", "-count=1", "-v", "-timeout", "10m"], cwd=repo, stdout=output, stderr=subprocess.STDOUT)
        observed = log.read_text()
        assert result.returncode != 0 and "--- FAIL: TestCheckedViewPrimitiveHeap" in observed
        assert "clang failed" not in observed and "[build failed]" not in observed
        assert "sanitized=false" in observed, observed
        results.append({"mutation": name, "caught": True, "log": str(log.relative_to(repo))})
        source.write_bytes(original)
finally:
    source.write_bytes(original)
(lane / "heap-mutants.json").write_text(json.dumps(results, indent=2) + "\n")
