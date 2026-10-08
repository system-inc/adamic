"""Run the pinned shared a-check method; prove every expected-error header matters."""
import argparse
import importlib.util
import json
import pathlib
import subprocess
import tempfile

GATE = '1247da58f065ba1a29737bf010cdd1e43090d519'
parser = argparse.ArgumentParser()
parser.add_argument('--mutants', action='store_true')
parser.add_argument('--out', type=pathlib.Path, required=True)
args = parser.parse_args()
root = pathlib.Path.cwd()
args.out.mkdir(parents=True, exist_ok=True)
runner = args.out / 'shared-gate.py'
runner.write_bytes(subprocess.check_output(['git', 'show', GATE + ':cloud/fast-gate/run.py']))
spec = importlib.util.spec_from_file_location('shared_gate', runner)
gate = importlib.util.module_from_spec(spec)
spec.loader.exec_module(gate)

class Harness:
    def __init__(self, label):
        self.arguments = argparse.Namespace(tree=str(root))
        self.result, self.steps, self.exits = {}, {}, {}
        self.failure, self.label = None, label

    def step(self, name, command):
        with (args.out / (self.label + '-build.log')).open('w') as log:
            status = subprocess.run(command, cwd=root, stdout=log, stderr=subprocess.STDOUT).returncode
        assert status == 0, (name, status)
        return True

    def spawn(self, command, stdout, stderr):
        return subprocess.Popen(command, cwd=root, stdout=stdout, stderr=stderr, text=True)

    def fail(self, name, detail):
        self.failure = detail

paths = sorted(str(p.relative_to(root)) for p in (root / 'stage3/shape-conformance').rglob('*.a'))
assert len(paths) == 13, paths
normal = Harness('normal')
gate.Gate.aCheck(normal, paths)
assert normal.exits['a-check'] == 0, normal.failure
results = {'tools_sha': GATE, 'normal': normal.result['a_check'], 'mutants': []}
if args.mutants:
    with tempfile.TemporaryDirectory(prefix='shape-header-mutants-') as directory:
        for index, path in enumerate(paths):
            source = (root / path).read_text()
            assert source.startswith('// a-check: type error TS2322\n'), path
            for mutation, content in (
                ('wrong-code', source.replace('TS2322', 'TS999999', 1)),
                ('missing-header', source.split('\n', 1)[1]),
            ):
                mutant = pathlib.Path(directory) / ('fixture-' + str(index) + '.a')
                mutant.write_text(content)
                h = Harness(str(index) + '-' + mutation)
                gate.Gate.aCheck(h, [str(mutant)])
                row = h.result['a_check'][str(mutant)]
                assert h.exits['a-check'] == 1 and h.failure, (path, mutation)
                assert row['outcome'] == 'type error' and 'error TS2322' in row['first'], row
                results['mutants'].append({'file': path, 'mutation': mutation, 'catcher': row})
                print(path + ': ' + mutation + ' caught by shared a-check', flush=True)
(args.out / 'results.json').write_text(json.dumps(results, indent=2) + '\n')
print('PASS shared a-check: %d fixtures, %d mutants' % (len(paths), len(results['mutants'])))
