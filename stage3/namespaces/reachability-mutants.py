"""Mutate namespace call reachability and readiness separately, restoring both."""
from pathlib import Path
import os
import subprocess

root = Path(__file__).resolve().parents[2]
subject = root / 'internal/lower/namespaces.go'
original = subject.read_text()
scratch = Path('/tmp/namespaces-reach-mutants')
scratch.mkdir(exist_ok=True)
start = original.index('func (l *lowering) namespaceReadyReads(')
end = original.index('func (l *lowering) namespaceReadyValue(', start)
header = original[start:original.index('{', start)+1]
mutations = [
    ('reaching-as-unreaching', original.replace('case ast.KindFunctionDeclaration:\n\t\t\tif declaration.Body() != nil {\n\t\t\t\treturn declaration\n\t\t\t}', 'case ast.KindFunctionDeclaration:\n\t\t\treturn nil'), './internal/lower', '^TestNamespaceInitializationReachability$', 'reaching call lost its initialization refusal'),
    ('drop-runtime-readiness', original[:start] + header + '\n\treturn nil\n}\n\n' + original[end:], './internal/oracle', '^TestNativeAgreesWithNode$/internal/oracle/testdata/namespaces_unknown_function_before.a$', 'exit codes differ'),
]
for name, changed, package, pattern, expected in mutations:
    assert changed != original, name
    try:
        subject.write_text(changed)
        log = scratch / (name + '.log')
        with log.open('w') as output:
            result = subprocess.run(['go', 'test', package, '-run', pattern, '-count=1', '-v'], cwd=root, env=dict(os.environ, ADAMIC_GATE_UNCACHED='1'), stdout=output, stderr=subprocess.STDOUT)
        observed = log.read_text()
        assert result.returncode != 0 and expected in observed, (name, result.returncode, observed)
        assert 'build failed' not in observed and 'error: unused' not in observed, (name, observed)
        print(name, 'caught:', expected, log)
    finally:
        subject.write_text(original)

# Removing only the const-enum hook must also fail by behavior, not by a warning.
enums = root / 'internal/lower/enums.go'
enum_original = enums.read_text()
needle = '\t\t\tif err == nil {\n\t\t\t\tvalue = l.namespaceReadyValue(node, value)\n\t\t\t}\n'
assert needle in enum_original
try:
    enums.write_text(enum_original.replace(needle, '', 1))
    log = scratch / 'inline-away-namespace-read.log'
    with log.open('w') as output:
        result = subprocess.run(['go', 'test', './internal/oracle', '-run', '^TestNativeAgreesWithNode$/internal/oracle/testdata/namespaces_unknown_enum_before.a$', '-count=1', '-v'], cwd=root, env=dict(os.environ, ADAMIC_GATE_UNCACHED='1'), stdout=output, stderr=subprocess.STDOUT)
    observed = log.read_text()
    assert result.returncode != 0 and 'exit codes differ' in observed and 'build failed' not in observed, observed
    print('inline-away-namespace-read caught: exit codes differ', log)
finally:
    enums.write_text(enum_original)
