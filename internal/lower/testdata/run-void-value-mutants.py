#!/usr/bin/env python3
"""Each mutant starts from and restores the final lowering implementation."""
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
logs = Path('/tmp/adamic-void-value-mutants')
logs.mkdir(exist_ok=True)
value_path = root / 'internal/lower/void_value.go'
statement_path = root / 'internal/lower/statements.go'
value_source = value_path.read_text()
statement_source = statement_path.read_text()
mutants = []
for name, kind, fixture in [
    ('drop-call', 'ir.Call', 'binder'),
    ('drop-closure', 'ir.CallClosure', 'closure'),
    ('drop-array-visit', 'ir.ArrayVisit', 'builtin'),
    ('drop-map-clear', 'ir.MapClear', 'builtin'),
    ('drop-map-visit', 'ir.MapForEach', 'builtin'),
]:
    old = 'b.body = append(b.body, ir.Evaluate{Value: value})'
    new = f'if _, dropped := value.({kind}); !dropped {{\n\t\t{old}\n\t}}'
    mutants.append((name, value_path, value_source.replace(old, new),
                    './internal/oracle', f'TestNativeAgreesWithNode/internal/oracle/testdata/void_value_{fixture}'))
mutants.append(('defined-result', value_path,
                value_source.replace('ir.Undefined{}', 'ir.StringConstant{Index: l.constant("wrong")}'),
                './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/void_value_binder'))
mutants.append(('lose-return', statement_path,
                statement_source.replace('ir.Evaluate{Value: value}, ir.Return{}', 'ir.Evaluate{Value: value}'),
                './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/void_value_checker'))
# Erasing a callable's actual result would make answer() print undefined instead of 7.
start = value_source.index('\t\ttargets := l.result.ClosureTargets(call)')
end = value_source.index('\t\targuments = append', start)
closure = value_source[:start] + value_source[end:]
mutants.append(('erase-callable-result', value_path, closure,
                './internal/lower', 'TestVoidValueErasedResultsStayNotYet/closure'))
unknown_start = value_source.index('\t\tif targets.Unknown || len(targets.Functions) == 0 {')
unknown_end = value_source.index('\t\tfor _, target', unknown_start)
unknown = value_source[:unknown_start] + value_source[unknown_end:]
mutants.append(('accept-unknown-target', value_path, unknown,
                './internal/lower', 'TestVoidValueErasedResultsStayNotYet/unknown_parameter'))
for name, path, mutated, package, test in mutants:
    try:
        original = value_source if path == value_path else statement_source
        assert mutated != original, name
        path.write_text(mutated)
        command = ['go', 'test', package, '-run', test, '-count=1', '-timeout', '10m']
        with (logs / f'{name}.log').open('w') as output:
            result = subprocess.run(command, cwd=root, stdout=output, stderr=subprocess.STDOUT)
        output = (logs / f'{name}.log').read_text()
        assert result.returncode != 0, f'{name} survived'
        assert '[build failed]' not in output and 'error: ' not in output, f'{name} broke a build'
        if package == './internal/oracle':
            assert 'stdout' in output, f'{name} was not caught by the Node output comparison: {output}'
        else:
            assert 'got <nil>' in output, f'{name} was not caught by the refusal assertion: {output}'
        print(f'{name}: caught (exit {result.returncode}), {logs / (name + ".log")}', flush=True)
    finally:
        value_path.write_text(value_source)
        statement_path.write_text(statement_source)
