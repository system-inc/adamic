#!/usr/bin/env python3
"""Run isolated compiler guard mutations and restore exact source bytes."""
from pathlib import Path
import subprocess
import sys

root = Path(__file__).resolve().parents[4]
logs = Path(sys.argv[1] if len(sys.argv) > 1 else "/tmp/views-callables-factory-mutants")
logs.mkdir(parents=True, exist_ok=True)
cases = [
    ("native-overload-result", "internal/native/runtime/view_callables_contract.h",
     "    if (!adamic_callable_representation_compatible(recorded->result, expected->result, recorded->result_mask, expected->result_mask)) return false;",
     "    /* Mutant: omit result variance. */",
     "^TestCheckedViewCallableFactory$/^createIdentifier$/^wrong-result$",
     'negative exit=0 stdout="read\\n" stderr=""'),
    ("javascript-overload-result", "internal/javascript/view_callables_contract.go",
     "    if (!adamicCallableRepresentationCompatible(recorded.result, expected.result, recorded.resultMask, expected.resultMask)) return false;",
     "    // Mutant: omit result variance.",
     "^TestCheckedViewCallableFactory$/^createIdentifier$/^wrong-result$",
     'negative exit=0 stdout="read\\n" stderr=""'),
    ("overload-union-read", "internal/lower/view_callables_read.go",
     "if len(l.checker.GetSignaturesOfType(member, checker.SignatureKindCall)) > 1 {",
     "if len(l.checker.GetSignaturesOfType(member, checker.SignatureKindCall)) > 100 {",
     "^TestPrepareViewCallableRead$", "unsupported read admitted"),
    ("resolved-declaration-membership", "internal/lower/view_callable_calls.go",
     "if !member {", "if false {",
     "^TestResolvedCallableDeclarationMembership$",
     "undeclared resolution admitted: <nil>"),
]
for name, relative, before, after, test, failure in cases:
    path = root / relative
    original = path.read_bytes()
    try:
        source = original.decode()
        if source.count(before) != 1:
            raise SystemExit(f"{name}: expected one guard")
        path.write_text(source.replace(before, after, 1))
        package = "./internal/lower" if name in {"resolved-declaration-membership", "overload-union-read"} else "./internal/oracle"
        log = logs / (name + ".log")
        with log.open("w") as output:
            result = subprocess.run(["go", "test", package, "-run", test, "-count=1", "-v", "-timeout", "5m"], cwd=root, stdout=output, stderr=subprocess.STDOUT)
        observed = log.read_text()
        if result.returncode == 0 or failure not in observed:
            raise SystemExit(f"{name}: not caught by its semantic assertion; inspect {log}")
        if name == "native-overload-result" and 'sanitized result exit=0 stdout="read\\n" stderr=""' not in observed:
            raise SystemExit(f"{name}: sanitizer must finish with the same forbidden output")
        print(f"{name}: caught by {failure}; log {log}")
    finally:
        path.write_bytes(original)
