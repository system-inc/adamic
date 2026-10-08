#!/usr/bin/env python3
"""Prove fresh error construction and string guard precision with source overlays."""
import json
from pathlib import Path
import subprocess
import tempfile

root = Path(__file__).resolve().parents[1]
mutants = [
    ('fresh-error-initializer-proof-removed', 'internal/lower/error_classes.go',
     'ir.Function{Name: name + "_initialize", LibraryGuarded: true}',
     'ir.Function{Name: name + "_initialize"}', './internal/oracle',
     'TestNativeAgreesWithNode/internal/oracle/testdata/57f2d04_with_frozen.a$',
     'a write to a potentially frozen object'),
    ('concat-bound-ignored', 'internal/lower/error_runtime_ranges.go',
     'if bound(node) <= maximumStringLength {',
     'if true || bound(node) <= maximumStringLength {', './internal/oracle',
     'TestNativeAgreesWithNode/internal/oracle/testdata/catchability-limits/d96d304_try_finally_concat.a$',
     'stdout differs'),
    ('unprotected-finally-guarded', 'internal/lower/error_runtime_ranges.go',
     'if tried.HasFinally {',
     'tried.Finally = rewriteChecks(tried.Finally, guard).([]ir.Statement); if tried.HasFinally {',
     './internal/lower', 'TestProtectedStringGuardPrecision$', 'cannot throw on this bounded protected path'),
    ('unknown-object-bound-ignored', 'internal/lower/error_runtime_ranges.go',
     'if !unknownFields {', 'if true || !unknownFields {', './internal/lower',
     'TestStringLengthBoundsIncludeUnknownObjects$', 'unknown object must retain the full string bound'),
    ('repeat-constant-length-ignored', 'internal/lower/exceptions.go',
     'safe = safe || l.stringLengthBounds()(node.Value)*count <= maximumStringLength',
     'safe = true || l.stringLengthBounds()(node.Value)*count <= maximumStringLength', './internal/lower',
     'TestRepeatRefusalChecksResultLength$', 'oversized constant repeat must remain a named refusal'),
    ('protected-catch-root-omitted', 'internal/lower/exceptions.go',
     'if lowered.HasCatch && lowered.HasFinally {',
     'if false && lowered.HasCatch && lowered.HasFinally {', './internal/oracle',
     'TestRuntimeStackCatchFinally$', 'stdout differs'),
    ('javascript-stack-error-identity-omitted', 'internal/javascript/javascript.go',
     'fmt.Sprintf("adamicCatch(%s)", caught)', 'fmt.Sprintf("%s", caught)', './internal/oracle',
     'TestRuntimeStackCleanup$', 'stdout differs'),
]
logs = Path('/tmp/adamic-runtime-source-mutants')
logs.mkdir(exist_ok=True)
with tempfile.TemporaryDirectory(prefix='adamic-runtime-overlay-') as directory:
    scratch = Path(directory)
    for name, relative, old, new, package, test, assertion in mutants:
        source_path = root / relative
        source = source_path.read_text()
        assert source.count(old) == 1, name
        variant = scratch / f'{name}.go'
        variant.write_text(source.replace(old, new))
        overlay = scratch / f'{name}.json'
        overlay.write_text(json.dumps({'Replace': {str(source_path): str(variant)}}))
        log = logs / f'{name}.log'
        with log.open('w') as output:
            result = subprocess.run(['go', 'test', '-overlay', str(overlay), package,
                                     '-run', test, '-count=1', '-v'], cwd=root,
                                    stdout=output, stderr=subprocess.STDOUT)
        observation = log.read_text()
        caught = result.returncode != 0 and assertion in observation and '[build failed]' not in observation
        print(f'{name}: exit {result.returncode}, intended assertion caught={caught}', flush=True)
        if not caught:
            raise RuntimeError(f'mutant survived or failed before its intended assertion: {log}')
