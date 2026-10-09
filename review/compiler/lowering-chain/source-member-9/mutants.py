#!/usr/bin/env python3
"""Run refusal ruling mutants through Go overlays without changing the checkout."""
import json
import os
from pathlib import Path
import subprocess
import tempfile

ROOT = Path.cwd()
MUTANTS = [
    ("merged-computed-read", "internal/lower/expression.go",
     "if value, handled, err := l.mergedComputedRead(node); handled {",
     "if value, handled, err := l.mergedComputedRead(node); false && handled {",
     "^TestMergedFieldTypeScriptComputedChecked$"),
    ("merged-declaration", "internal/lower/merged_fields.go",
     "if adamic {", "if false && adamic {", "^TestMergedFieldAdamicUnusedRefused$"),
    ("merged-read", "internal/lower/merged_fields.go",
     "l.result.CheckedFields[field.Name] = true", "l.result.CheckedFields[field.Name] = false",
     "^TestMergedFieldTypeScriptMissingChecked$"),
    ("function-annotation", "internal/lower/function_rulings.go",
     'node.Kind == ast.KindTypeReference &&', 'false && node.Kind == ast.KindTypeReference &&',
     '^TestFunctionAdamicUnusedAnnotationRefused$'),
    ("function-call", "internal/lower/function_rulings.go",
     'if opaque {', 'if false && opaque {', '^TestFunctionTypeScriptCallRefused$'),
    ("function-length", "internal/lower/functions.go",
     'function.SourceLength = sourceFunctionLength(declaration)', 'function.SourceLength = 0',
     '^TestFunctionTypeScriptLengthAgreesWithNode$'),
]


def main():
    directory = Path(os.environ.get("REFUSAL_MUTANT_LOGS", "/tmp/refusal-rulings-mutants"))
    directory.mkdir(parents=True, exist_ok=True)
    for name, filename, before, after, selector in MUTANTS:
        original = ROOT / filename
        source = original.read_text()
        if source.count(before) != 1:
            raise RuntimeError(f"{name}: expected one mutation site")
        with tempfile.TemporaryDirectory(prefix="refusal-mutant-") as scratch:
            replacement = Path(scratch) / (original.name + '.txt')
            replacement.write_text(source.replace(before, after, 1))
            overlay = Path(scratch) / "overlay.json"
            overlay.write_text(json.dumps({"Replace": {str(original): str(replacement)}}))
            log = directory / f"{name}.log"
            with log.open("w") as output:
                result = subprocess.run(["go", "test", f"-overlay={overlay}", "./internal/oracle",
                                         "-run", selector, "-count=1", "-v"],
                                        cwd=ROOT, stdout=output, stderr=subprocess.STDOUT,
                                        env={**os.environ, "ADAMIC_GATE_UNCACHED": "1"})
            text = log.read_text()
            if result.returncode == 0 or "--- FAIL: Test" not in text or "[build failed]" in text or "clang failed" in text:
                raise RuntimeError(f"{name}: no behavioral catcher, see {log}")
            print(f"{name}: caught by {selector}; {log}", flush=True)


if __name__ == "__main__":
    main()
