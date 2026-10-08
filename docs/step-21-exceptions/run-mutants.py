"""Restore each production-source mutant and require its intended semantic catcher."""
import os
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parents[2]
LOGS = ROOT / 'docs/step-21-exceptions/evidence'
mutants = [
    ('missing-exception-edge', 'internal/flow/build.go',
     'if CanThrow(b.program, instruction) {', 'if false && CanThrow(b.program, instruction) {',
     'TestNativeAgreesWithNode/internal/oracle/testdata/step21_liveness[.]a$', 'runtime error:'),
    ('assignment-kills-handler-value', 'internal/flow/liveness.go',
     'ok && position == len(block.Instructions)-1', 'false && ok && position == len(block.Instructions)-1',
     '^TestStep21ThrowPathLiveness$', 'old text is dead'),
    ('throw-temporaries-leak', 'internal/native/exceptions.go',
     'e.line("adamic_release(%s);", e.owned[index])',
     'e.line("adamic_retain(%s);", e.owned[index])',
     'TestNativeAgreesWithNode/internal/oracle/testdata/exceptions[.]a$', 'leaks:'),
    ('pending-finally-leak', 'internal/native/exceptions.go',
     'e.hold(pending)', '// mutant: pending completion has no owner',
     'TestNativeAgreesWithNode/internal/oracle/testdata/step21_finally_completion[.]a$', 'leaks:'),
    ('error-message-not-retained', 'internal/native/runtime/exceptions.c',
     'error->slots[1].reference = adamic_retain(message);', 'error->slots[1].reference = message;',
     'TestNativeAgreesWithNode/internal/oracle/testdata/step21_catch_callback[.]a$', 'AddressSanitizer: heap-use-after-free'),
]
for name, relative, old, new, test, expected in mutants:
    target = ROOT / relative
    original = target.read_bytes()
    source = original.decode()
    assert source.count(old) == 1, (name, source.count(old))
    try:
        target.write_text(source.replace(old, new, 1))
        log = LOGS / (name + '.log.txt')
        with log.open('w') as output:
            result = subprocess.run(['go', 'test', './internal/oracle', '-run', test,
                '-count=1', '-v', '-timeout', '5m'], cwd=ROOT,
                env=dict(os.environ, ADAMIC_GATE_UNCACHED='1'), stdout=output,
                stderr=subprocess.STDOUT, timeout=360)
        observed = log.read_text()
        assert result.returncode != 0 and expected in observed, (name, result.returncode, observed)
        assert 'clang failed' not in observed and '[build failed]' not in observed, (name, observed)
        print(f'{name}: exit {result.returncode}, caught by {expected}', flush=True)
    finally:
        target.write_bytes(original)
