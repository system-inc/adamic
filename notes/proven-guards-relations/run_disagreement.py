import json, os, pathlib, subprocess
root = pathlib.Path(__file__).resolve().parents[2]
notes = pathlib.Path(__file__).resolve().parent
registration = root / 'internal/oracle/proven_disagreement_probe_test.go'
assert not registration.exists(), 'temporary registration already exists'
try:
    registration.write_text('package oracle\nfunc init() { fixtures = append(fixtures, struct { path string; lowers bool; checked bool }{"notes/proven-guards-relations/mutable_kind_guard.a", true, false}) }\n')
    command = ['go', 'test', './internal/oracle', '-run', 'TestNativeAgreesWithNode/notes/proven-guards-relations/mutable_kind_guard[.]a', '-count=1', '-timeout', '30m', '-v']
    run = subprocess.run(command, cwd=root, env=dict(os.environ, ADAMIC_GATE_UNCACHED='1'), text=True, capture_output=True)
    (notes / 'mutable-kind-oracle.log').write_text('Command: ' + ' '.join(command) + '\nExit: ' + str(run.returncode) + '\n' + run.stdout + run.stderr)
    assert run.returncode != 0 and 'exit codes differ' in run.stdout, run.stdout + run.stderr
finally:
    registration.unlink()
file = 'notes/proven-guards-relations/mutable_kind_guard.a'
binary = '/tmp/adamic-gate/mutable_kind_guard'
build_command = ['go', 'run', './cmd/adamic', 'build', file, '-o', binary]
build = subprocess.run(build_command, cwd=root, text=True, capture_output=True)
assert build.returncode == 0, build.stdout + build.stderr
node_command = ['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', file]
node = subprocess.run(node_command, cwd=root, text=True, capture_output=True)
native = subprocess.run([binary], cwd=root, text=True, capture_output=True)
js_command = ['go', 'run', './cmd/adamic', 'js', file]
js = subprocess.run(js_command, cwd=root, text=True, capture_output=True)
assert js.returncode == 0, js.stderr
js_file = pathlib.Path('/tmp/adamic-gate/mutable_kind_guard.mjs')
js_file.write_text(js.stdout)
backend_command = ['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(js_file)]
backend = subprocess.run(backend_command, cwd=root, text=True, capture_output=True)
results = dict(file=file, oracle_command=command, build_command=build_command, node_command=node_command, native_command=[binary], javascript_command=js_command, backend_command=backend_command, node=dict(exit=node.returncode, stdout=node.stdout, stderr=node.stderr), native=dict(exit=native.returncode, stdout=native.stdout, stderr=native.stderr), backend=dict(exit=backend.returncode, stdout=backend.stdout, stderr=backend.stderr), likely_cause='internal/lower/class.go:446 emits a discriminant store without proving the resulting complete variant; internal/lower/predicates.go:160 accepts checker narrowing that assumes that variant invariant')
(notes / 'disagreements.json').write_text(json.dumps([results], indent=2) + '\n')
assert node.returncode == 0 and node.stdout == 'branch undefined\n' and node.stderr == ''
assert native.returncode == 70 and native.stdout == '' and native.stderr == 'adamic: panic: compiler bug: a field the checker proved is there is missing\n'
assert (backend.returncode, backend.stdout, backend.stderr) == (node.returncode, node.stdout, node.stderr)
print('Observed Node/native disagreement; temporary oracle registration removed.')
