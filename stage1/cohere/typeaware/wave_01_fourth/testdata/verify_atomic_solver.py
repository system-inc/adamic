#!/usr/bin/env python3
"""Compare the native atomic solver with unchanged Go Solve and rule lattice."""
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

virtual = repo / 'cohere/adamic_wave01_solver.go'
overlay = out / 'overlay.json'
overlay.write_text(json.dumps({'Replace': {
    str(virtual): str(source / 'atomic_solver_main.go'),
    str(repo / 'cohere/internal/lint/rules/core/adamic_wave01_solver.go'): str(source / 'atomic_solver_capture.go'),
    str(repo / 'cohere/internal/lint/ecmascript/control_flow_graph/adamic_wave01_solver.go'): str(source / 'atomic_solver_graph.go'),
}}))
oracle = out / 'oracle'
run('go-build', ['go', 'build', '-overlay', overlay, '-o', oracle, virtual], repo / 'cohere')
expected, error = run('go', [oracle])
assert not error
native = out / 'native'
run('native-build', [compiler, 'build', source / 'atomic_solver.a', '-o', native, '--sanitize'])
actual, error = run('native', [native])
assert not error and actual == expected
assert len(actual.splitlines()) == 96
# One sweep misses information propagated around the back edge. It remains valid
# native code and exits cleanly, so only the production-Go bytes reject it.
original = (source.parent / 'require_atomic_updates/solver.a').read_text()
mutant = out / 'solver-mutant.a'
mutant.write_text(original.replace('while(changed) {', 'while(changed && rounds < 1) {'))
entry = out / 'mutant-entry.a'
entry.write_text((source / 'atomic_solver.a').read_text().replace('../require_atomic_updates/solver.a', str(mutant)))
# Preserve the state import in the scratch mutant without changing the real file.
mutant.write_text(mutant.read_text().replace('./state.a', str(source.parent / 'require_atomic_updates/state.a')))
mutant_binary = out / 'mutant'
run('mutant-build', [compiler, 'build', entry, '-o', mutant_binary, '--sanitize'])
actual, error = run('mutant', [mutant_binary])
assert not error and actual != expected
print('PASS 16 graphs, 96 block boundaries; sanitizers clean; one-sweep mutant caught by Go bytes')
