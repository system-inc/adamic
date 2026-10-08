#!/usr/bin/env python3
"""Run collection-order and single-match lowering mutants and restore the source."""
from pathlib import Path
import os
import subprocess

root = Path(__file__).resolve().parents[3]
source = root / "internal/lower/regexp.go"
original = source.read_text()
mutants = {
    "collection-before-callbacks": (
        'ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: b.read(global)}, Then: []ir.Statement{ir.Break{}}},',
        'ir.Evaluate{Value: ir.CallClosure{Closure: callback, Returns: ir.String, Arguments: []ir.Expression{whole}}},\n' +
        'ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: b.read(global)}, Then: []ir.Statement{ir.Break{}}},'),
    "nonglobal-single-match": (
        'ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: b.read(global)}, Then: []ir.Statement{ir.Break{}}},',
        'ir.Break{},'),
}
for name, (before, after) in mutants.items():
    assert original.count(before) == 1
    log = Path('/tmp/library-small-regex-mutant-' + name + '.log')
    try:
        source.write_text(original.replace(before, after))
        env = dict(os.environ, ADAMIC_GATE_UNCACHED="1")
        with log.open("w") as output:
            result = subprocess.run([
                "go", "test", "./internal/oracle", "-run",
                "TestNativeAgreesWithNode/internal/oracle/testdata/notyet_library_regex_callback",
                "-count=1", "-timeout", "2m", "-v"],
                cwd=root, env=env, stdout=output, stderr=subprocess.STDOUT)
        text = log.read_text()
        assert result.returncode != 0 and 'stdout differs' in text, text[-4000:]
        assert 'clang failed' not in text and 'build failed' not in text, text[-4000:]
        print(name + ': caught by source Node stdout comparison; log ' + str(log), flush=True)
    finally:
        source.write_text(original)
