#!/usr/bin/env python3
"""Run independent unit A guard mutants. Restore every file after each run."""
import os
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
logs = Path('/tmp/adamic-type-rule-mutants')
logs.mkdir(exist_ok=True)
mutants = [
    ('cohere-invariant', 'internal/lower/refusals.go',
     '[]string{"adamic/invariant-mutable", "adamic/nominal-class"}',
     '[]string{"adamic/nominal-class"}',
     ['./internal/oracle', '-run', 'TestTypeRuleProbesStayRefused/conditional_copy'],
     None, 'want adamic/invariant-mutable, got <nil>'),
    ('cohere-nominal', 'internal/lower/refusals.go',
     '[]string{"adamic/invariant-mutable", "adamic/nominal-class"}',
     '[]string{"adamic/invariant-mutable"}',
     ['./internal/lower', '-run', 'TestCohereNominalInvocation'],
     None, 'want cohere nominal refusal, got <nil>'),
    ('constraint-source', 'internal/lower/invariance.go',
     'constrained := own.Flags()&checker.TypeFlagsTypeParameter != 0',
     'constrained := false',
     ['./internal/oracle', '-run', 'TestNativeAgreesWithNode/internal/oracle/testdata/type_rules_parameter_(array|map|field)'],
     '1', 'AddressSanitizer: heap-buffer-overflow'),
    ('weak-nominal', 'internal/lower/class_inheritance.go',
     'func (l *lowering) retainedNominalTarget(proven *checker.Type, seen map[*checker.Type]bool) bool {',
     'func (l *lowering) retainedNominalTarget(proven *checker.Type, seen map[*checker.Type]bool) bool {\n if proven != nil && l.weakTarget(proven) != nil { return false }',
     ['./internal/oracle', '-run', 'TestNativeAgreesWithNode/internal/oracle/testdata/type_rules_weak_(method|callback)'],
     '1', 'stdout differs'),
    ('union-nominal', 'internal/lower/class_inheritance.go',
     'func (l *lowering) retainedNominalTarget(proven *checker.Type, seen map[*checker.Type]bool) bool {',
     'func (l *lowering) retainedNominalTarget(proven *checker.Type, seen map[*checker.Type]bool) bool {\n if proven != nil && l.weakTarget(proven) == nil && proven.Flags()&checker.TypeFlagsUnion != 0 { return false }',
     ['./internal/oracle', '-run', 'TestNativeAgreesWithNode/internal/oracle/testdata/type_rules_union_(method|callback)'],
     '1', 'stdout differs'),
]
for name, filename, before, after, args, probes, catcher in mutants:
    path = root / filename
    original = path.read_text()
    if original.count(before) != 1:
        raise SystemExit(f'{name}: mutation anchor must occur once')
    env = dict(os.environ, ADAMIC_GATE_UNCACHED='1')
    if probes:
        env['ADAMIC_TYPE_RULE_PROBES'] = probes
    else:
        env.pop('ADAMIC_TYPE_RULE_PROBES', None)
    try:
        path.write_text(original.replace(before, after))
        with (logs / (name + '.log')).open('w') as out:
            result = subprocess.run(['go', 'test', *args, '-v', '-count=1', '-timeout', '30m'], cwd=root, env=env, stdout=out, stderr=subprocess.STDOUT)
        output = (logs / (name + '.log')).read_text()
        if result.returncode == 0 or catcher not in output or '[build failed]' in output:
            raise SystemExit(f'{name}: not caught by {catcher}; see {logs/name}.log')
        print(f'{name}: caught by {catcher}; exit={result.returncode}', flush=True)
    finally:
        path.write_text(original)
print('5 independent mutants caught; all files restored')
