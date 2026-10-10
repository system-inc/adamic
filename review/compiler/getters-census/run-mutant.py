#!/usr/bin/env python3
"""Temporarily make a literal getter run at construction, then restore it."""
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
source = root / 'internal/lower/class_accessors.go'
evidence = root / 'review/compiler/getters-census'
original = source.read_text()
needle = '\t\t\tl.registerAccessor(class, property, function)\n'
mutation = '''\t\t\t// Census mutant: a memoized binary getter runs before its first read.
\t\t\tif property.Kind == ast.KindGetAccessor && name.Text() == "createComma" {
\t\t\t\tliteral.Fields = append(literal.Fields, ir.Field{
\t\t\t\t\tName: "#eager", Private: true,
\t\t\t\t\tValue: ir.CallClosure{Direct: function + 1, Closure: value,
\t\t\t\t\t\tArguments: []ir.Expression{ir.ObjectLiteral{}}, Returns: l.result.Functions[function].Returns},
\t\t\t\t})
\t\t\t}
'''
assert original.count(needle) == 1
command = "source /workspace/adamic-tools/env.sh; ADAMIC_GATE_UNCACHED=1 timeout 90 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^getters_census_binary_inline\\.a$' -count=1 -timeout 90s -parallel 4 -v"
try:
    source.write_text(original.replace(needle, mutation + needle))
    diff = subprocess.run(['git', 'diff', '--', 'internal/lower/class_accessors.go'], cwd=root, capture_output=True, text=True, check=True)
    (evidence / 'eager-getter.patch').write_text(diff.stdout)
    with (evidence / 'mutant.log').open('w') as log:
        result = subprocess.run(['bash', '-lc', command], cwd=root, stdout=log, stderr=subprocess.STDOUT, timeout=120)
    output = (evidence / 'mutant.log').read_text()
    assert result.returncode != 0, 'mutant survived'
    assert 'stdout differs' in output, 'mutant failed outside the intended comparison'
    assert 'Lower:' not in output and 'error:' not in output, 'mutant failed to compile'
    print('eager getter caught by Node stdout comparison; exit', result.returncode)
finally:
    source.write_text(original)
