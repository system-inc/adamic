#!/usr/bin/env python3
"""Compare native placement with unchanged production Go placeDeferrals."""
import json
import os
from pathlib import Path
import subprocess
import sys

repo = Path(__file__).resolve().parents[5]
source = Path(__file__).resolve().parent
out = Path(sys.argv[1]).resolve()
out.mkdir(parents=True, exist_ok=True)
compiler = Path(sys.argv[2]).resolve()

def run(label, command, cwd=repo):
    with (out / (label + '.stdout')).open('wb') as stdout, (out / (label + '.stderr')).open('wb') as stderr:
        result = subprocess.run([str(v) for v in command], cwd=cwd, stdout=stdout, stderr=stderr,
                                env=dict(os.environ, ASAN_OPTIONS='detect_leaks=1', UBSAN_OPTIONS='halt_on_error=1'))
    assert result.returncode == 0, (label, result.returncode)
    return (out / (label + '.stdout')).read_bytes(), (out / (label + '.stderr')).read_bytes()

virtual = repo / 'cohere/adamic_wave01_deferrals.go'
overlay = out / 'overlay.json'
overlay.write_text(json.dumps({'Replace': {
    str(virtual): str(source / 'atomic_deferrals_main.go'),
    str(repo / 'cohere/internal/lint/rules/core/adamic_wave01_deferrals.go'): str(source / 'atomic_deferrals_capture.go'),
}}))
oracle = out / 'oracle'
run('go-build', ['go', 'build', '-overlay', overlay, '-o', oracle, virtual], repo / 'cohere')
expected, error = run('go', [oracle])
assert not error
native = out / 'native'
run('native-build', [compiler, 'build', source / 'atomic_deferrals.a', '-o', native, '--sanitize'])
actual, error = run('native', [native])
assert not error and actual == expected
assert len(actual.splitlines()) == 32
# Dropping the original floor changes placement, while leaving valid native code.
original = (source.parent / 'require_atomic_updates/deferrals.a').read_text()
mutant = out / 'deferrals-mutant.a'
mutant.write_text(original.replace('let slot = deferral.originalIndex;', 'let slot = 0;'))
entry = out / 'mutant-entry.a'
entry.write_text((source / 'atomic_deferrals.a').read_text().replace('../require_atomic_updates/deferrals.a', str(mutant)))
mutant_binary = out / 'mutant'
run('mutant-build', [compiler, 'build', entry, '-o', mutant_binary, '--sanitize'])
actual, error = run('mutant', [mutant_binary])
assert not error and actual != expected
print('PASS 32 placements; sanitizers clean; lost-floor mutant caught by Go bytes')
