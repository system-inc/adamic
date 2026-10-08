#!/usr/bin/env python3
"""Prove that neither emitter nor lowering may substitute physical Set storage."""
from pathlib import Path
import os
import subprocess

root = Path(__file__).resolve().parents[4]
evidence = root / "stage3/interface-downcasts/lane5/code/logs"
mutants = [
    ("native-set-domain", "internal/native/runtime/view_set_intrinsics.c", "set->view_set_element = callable_element;", "set->view_set_element = element; (void)callable_element;"),
    ("javascript-set-domain", "internal/javascript/javascript.go", "expression.CallableElement", "expression.Element"),
    ("lower-set-domain", "internal/lower/view_set_domains.go", "element := l.concrete(arguments[0])", "element := l.concrete(arguments[0]); if of, known := l.representation(element); known { return of }"),
]
for name, relative, old, new in mutants:
    path = root / relative
    original = path.read_text()
    assert old in original
    try:
        path.write_text(original.replace(old, new))
        log = evidence / (name + "-mutant.log")
        with log.open("w") as output:
            result = subprocess.run(
                ["go", "test", "./internal/oracle", "-run", "^TestCheckedViewCallableCodeSetDomains$|^TestCheckedViewCallableCodeSetPrimitivePairs$|^TestCheckedViewCallableCodeSetInstantiations$", "-count=1", "-v", "-timeout", "5m"],
                cwd=root, env=dict(os.environ, ADAMIC_GATE_UNCACHED="1"), stdout=output, stderr=subprocess.STDOUT,
            )
        text = log.read_text()
        assert result.returncode != 0, f"{name}: mutant survived"
        assert 'negative exit 0 stdout "completed\\n"' in text, f"{name}: no forbidden completing writer: {log}"
        assert "clang failed" not in text and "[build failed]" not in text, f"{name}: build failure is not evidence"
        assert text.count("--- FAIL: TestCheckedViewCallableCodeSetPrimitivePairs/") == 8, f"{name}: every pair must catch the mutant"
        assert "generic domain exit 0" in text, f"{name}: instantiation certificate must catch the mutant"
        print(f"{name}: forbidden literal-domain writer completed; all eight pair negatives caught the mutant")
    finally:
        path.write_text(original)
