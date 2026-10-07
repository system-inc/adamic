"""Prove a call that can free a loop's array is not treated as harmless.

Two independent mistakes:
- a direct call counts as unchanging, so a pop through a callee borrows the element;
- global-argument lending does not follow callees, so a parameter loop borrows an
  array whose only count the callee then drops.
"""
import os
import pathlib
import subprocess

root = pathlib.Path(__file__).resolve().parents[3]
logs = pathlib.Path(os.environ.get("ADAMIC_CALL_LOGS", "/tmp/adamic-loop-call-mutants"))
logs.mkdir(parents=True, exist_ok=True)
environment = os.environ.copy()
environment["ADAMIC_GATE_UNCACHED"] = "1"
oracle = [
    "go", "test", "./internal/oracle", "-run",
    "TestNativeAgreesWithNode/internal/oracle/testdata/borrow_loop_calls", "-count=1",
]
element = root / "internal/native/element_borrow.go"
touches = root / "internal/native/reuse.go"
originals = {element: element.read_text(), touches: touches.read_text()}
cases = [
    (
        "call-unchanging",
        element,
        originals[element].replace(
            "for _, target := range program.CallTargets(expression) {\n"
            "\t\t\tif changing[target] {\n"
            "\t\t\t\treturn false\n"
            "\t\t\t}\n"
            "\t\t}\n"
            "\t\treturn true",
            "return true",
            1,
        ),
    ),
    (
        "touches-ignores-callees",
        touches,
        originals[touches].replace(
            "if touches(program, target, global, seen) {",
            "if false && touches(program, target, global, seen) {",
            1,
        ),
    ),
]
try:
    for name, path, changed in cases:
        assert changed != originals[path], name
        path.write_text(changed)
        try:
            with (logs / (name + ".log")).open("w") as log:
                result = subprocess.run(
                    oracle, cwd=root, env=environment, stdout=log, stderr=subprocess.STDOUT
                )
            output = (logs / (name + ".log")).read_text()
            evidence = "AddressSanitizer: heap-use-after-free"
            if result.returncode == 0 or evidence not in output:
                raise RuntimeError(
                    f"{name}: mutant survived or failed for another reason; see {logs / (name + '.log')}"
                )
            print(name, result.returncode, evidence, flush=True)
        finally:
            path.write_text(originals[path])
finally:
    for path, original in originals.items():
        path.write_text(original)
