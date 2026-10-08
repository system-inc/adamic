#!/usr/bin/env python3
"""Prove exception precision and method receiver assertions under Go source overlays."""
import json
from pathlib import Path
import subprocess
import tempfile

root = Path(__file__).resolve().parents[1]
logs = Path('/tmp/adamic-counts-mutants')
logs.mkdir(exist_ok=True)
mutants = [
    ('callback-method-receiver-unbound', 'internal/lower/expression.go',
     'property.Method = true', 'property.Method = false',
     './internal/lower', 'TestCallbackTypedClassMethodCallBindsReceiver$', 'must bind its receiver'),
    ('loop-presence-unproved', 'internal/lower/exception_paths.go',
     'if node, ok := value.Interface().(ir.Loop); ok {',
     'if node, ok := value.Interface().(ir.Loop); ok && false {',
     './internal/lower', 'TestLoopPresenceSurvivesRuntimeMutation$', 'loop condition proves cursor present'),
    ('unknown-store-ignored', 'internal/lower/exception_bounds.go',
     'unknown[name] = true\n\t\t\t\tbreak', 'continue',
     './internal/lower', 'TestGeneratedGuardFactsIncludeUnknownStores$', 'must remain throwing'),
    ('counter-body-write-ignored', 'internal/lower/exception_bounds.go',
     '!written[read.Local]', 'true',
     './internal/lower', 'TestGeneratedGuardCounterWritesRemainThrowing$', 'must remain throwing'),
    ('presence-call-mutation-ignored', 'internal/lower/exception_paths.go',
     '!root || writes[target][local]', 'false && (!root || writes[target][local])',
     './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/error_checks.a$', 'exit codes differ'),
    ('all-calls-pure', 'internal/native/borrow.go',
     'return expression.Pure', 'return true || expression.Pure',
     './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/lent_reads.a$', 'heap-use-after-free'),
    ('all-calls-throw', 'internal/ir/ir.go',
     'func (p *Program) CallMayThrow(call Call) bool {',
     'func (p *Program) CallMayThrow(call Call) bool { return true /* mutant */; ',
     './internal/lower', 'TestMayThrowPrecision$', 'cannot throw'),
    ('all-functions-throw', 'internal/lower/exceptions.go',
     'if !functions[index].MayThrow && l.throwsOut(functions[index].Body) {',
     'if !functions[index].MayThrow {',
     './internal/lower', 'TestMayThrowPrecision$', 'cannot throw'),
    ('readiness-unproved', 'internal/lower/exceptions.go',
     'l.preciseChecks()', '',
     './internal/lower', 'TestMayThrowPrecision$', 'cannot throw'),
]
with tempfile.TemporaryDirectory(prefix='adamic-counts-overlay-') as directory:
    scratch = Path(directory)
    for name, relative, old, new, package, test, assertion in mutants:
        source = (root / relative).read_text()
        if source.count(old) != 1:
            raise RuntimeError(f'{name}: expected exactly one target')
        variant = scratch / f'{name}.go'
        variant.write_text(source.replace(old, new))
        overlay = scratch / f'{name}.json'
        overlay.write_text(json.dumps({'Replace': {str(root / relative): str(variant)}}))
        log = logs / f'{name}.log'
        with log.open('w') as output:
            result = subprocess.run(['go', 'test', '-overlay', str(overlay), package,
                                     '-run', test, '-count=1', '-v'], cwd=root,
                                    stdout=output, stderr=subprocess.STDOUT)
        observation = log.read_text()
        caught = result.returncode != 0 and assertion in observation and '[build failed]' not in observation
        print(f'{name}: exit {result.returncode}, assertion caught={caught}', flush=True)
        if not caught:
            raise RuntimeError(f'{name} survived or failed before its assertion: {log}')
