"""Prove iterator ownership and borrowed cleanup against the real sanitizer oracle."""
import os
import pathlib
import subprocess

root = pathlib.Path(__file__).resolve().parents[3]
path = root / 'internal/native/emit_statements.go'
original = path.read_text()
logs = pathlib.Path(os.environ.get('ADAMIC_HOLD_LOGS', '/tmp/adamic-loop-array-hold-mutants'))
logs.mkdir(parents=True, exist_ok=True)
environment = os.environ.copy()
environment['ADAMIC_GATE_UNCACHED'] = '1'
oracle = ['go', 'test', './internal/oracle', '-run',
          'TestNativeAgreesWithNode/internal/oracle/testdata/borrow_loop', '-count=1']
checks = ['go', 'test', './internal/native', '-run', 'TestLoopArrayHoldC', '-count=1']
cases = []
# Bypass only the iterator proof, leaving element binding ownership unchanged.
# Captured/global reads and call results must still expire in statement cleanup.
for name in ['reassignBody', 'closureReassignBody', 'globalReassignBody', 'freshBody']:
    condition = f'(e.elementBorrows[e.at] || e.function != nil && e.function.Name == "{name}")'
    changed = original.replace('} else if e.elementBorrows[e.at] {', '} else if ' + condition + ' {', 1)
    changed = changed.replace('if !e.elementBorrows[e.at] {', 'if !' + condition + ' {', 1)
    cases.append((name, changed, oracle, 'AddressSanitizer: heap-use-after-free'))
# Keep cleanup scheduled but remove the count it releases. This function throws before the
# next iterator step; ASan must catch the unwind release, not a normal-path length read.
cases.append(('throwHeldBody', original.replace('held, e.kept(iterable))',
    'held, func() string { if e.function != nil && e.function.Name == "throwHeldBody" { return iterable }; return e.kept(iterable) }())', 1),
    oracle, 'AddressSanitizer: heap-use-after-free'))
# Borrowed arrays have no iterator count to release even on an exceptional exit.
cases.append(('borrowed-throw-cleanup', original.replace('if !e.elementBorrows[e.at] {',
    'if !e.elementBorrows[e.at] || e.function != nil && e.function.Name == "throwBody" {', 1),
    oracle, 'AddressSanitizer: heap-use-after-free'))
cases.append(('hold-elision-disabled', original.replace('} else if e.elementBorrows[e.at] {',
    '} else if false && e.elementBorrows[e.at] {', 1).replace('if !e.elementBorrows[e.at] {',
    'if true || !e.elementBorrows[e.at] {', 1), checks, '--- FAIL: TestLoopArrayHoldC'))
for name, changed, command, evidence in cases:
    try:
        assert changed != original, name
        path.write_text(changed)
        with (logs / (name + '.log')).open('w') as log:
            result = subprocess.run(command, cwd=root, env=environment, stdout=log, stderr=subprocess.STDOUT)
        output = (logs / (name + '.log')).read_text()
        if result.returncode == 0 or evidence not in output:
            raise RuntimeError(f'{name}: mutant survived or failed for another reason; see {logs / (name + ".log")}')
        print(name, result.returncode, evidence, flush=True)
    finally:
        path.write_text(original)
