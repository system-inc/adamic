"""Prove the main replay preserves terminal checks without classifying new payloads."""
import os
from pathlib import Path
import subprocess

ROOT = Path.cwd()
LOGS = ROOT / 'review/compiler/lowering-chain/source-member-17/evidence'
mutants = [
    ('changed-source-keeps-terminal-mode', 'oracle/node.mjs',
     ".digest('hex') === expectedHash", ".digest('hex') !== undefined",
     '^TestStep21TerminalOracleSourceHash$', 'changed source borrowed terminal convention'),
    ('changed-dependency-keeps-terminal-mode', 'oracle/node.mjs',
     '(terminalDependencies.get(key) ?? [])', '[]',
     '^TestStep21TerminalOracleDependencyHash$', 'changed dependency borrowed terminal convention'),
    ('backend-drops-entry-source', 'internal/oracle/oracle_test.go',
     'onJavaScriptBackend(t, program, path)', 'onJavaScriptBackend(t, program)',
     'TestNativeAgreesWithNode/internal/oracle/testdata/from_code_point_fails[.]a$', 'JavaScript backend: exit codes differ'),
    ('readiness-runs-handler', 'internal/javascript/javascript.go',
     "panic(`ReferenceError: Cannot access '${name}' before initialization`);",
     "throw new ReferenceError(`Cannot access '${name}' before initialization`);",
     '^TestStep21ReadinessIsTerminal$', 'JavaScript readiness guard ran catch/finally'),
]
for name, relative, old, new, test, expected in mutants:
    path = ROOT / relative
    original = path.read_bytes()
    source = original.decode()
    assert source.count(old) == 1, (name, source.count(old))
    try:
        path.write_text(source.replace(old, new, 1))
        with (LOGS / (name + '.log.txt')).open('w') as output:
            result = subprocess.run(['go', 'test', './internal/oracle', '-run', test,
                '-count=1', '-v', '-timeout=5m'], cwd=ROOT,
                env=dict(os.environ, ADAMIC_GATE_UNCACHED='1'), stdout=output,
                stderr=subprocess.STDOUT, timeout=360)
        observed = (LOGS / (name + '.log.txt')).read_text()
        assert result.returncode != 0 and expected in observed, (name, result.returncode, observed)
        assert '[build failed]' not in observed and 'clang failed' not in observed
        print(f'{name}: exit {result.returncode}, caught by {expected}', flush=True)
    finally:
        path.write_bytes(original)
