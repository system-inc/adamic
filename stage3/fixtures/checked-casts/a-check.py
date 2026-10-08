"""Run the established Gate.aCheck on only these fixtures and on header mutants."""
import importlib.util
import json
from pathlib import Path
import subprocess
import sys

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[2]
gate_path, scratch_path = map(Path, sys.argv[1:3])
scratch_path.mkdir(parents=True, exist_ok=True)
spec = importlib.util.spec_from_file_location('fastgate', gate_path)
gate = importlib.util.module_from_spec(spec)
spec.loader.exec_module(gate)


class Harness:
    def __init__(self, label):
        self.arguments = type('Arguments', (), {'tree': str(ROOT)})()
        self.result, self.steps, self.exits = {}, {}, {}
        self.failure = None
        self.label = label

    def step(self, name, command):
        with (scratch_path / (self.label + '-build.log')).open('w') as log:
            result = subprocess.run(command, cwd=ROOT, stdout=log, stderr=subprocess.STDOUT)
        return result.returncode == 0

    def spawn(self, command, stdout, stderr):
        return subprocess.Popen(command, cwd=ROOT, stdout=stdout, stderr=stderr, text=True)

    def fail(self, name, detail):
        self.failure = detail


fixtures = json.loads((HERE / 'fixtures.json').read_text())
paths = [str(HERE / row['file']) for row in fixtures]
harness = Harness('fixtures')
gate.Gate.aCheck(harness, paths)
assert harness.exits['a-check'] == 0, harness.failure
result = {'gate_revision': '914ea6d7d3ef0aeca2ac68baa76eb4604556479b', 'fixtures': harness.result['a_check'], 'mutants': []}
for mutation in ['remove', 'wrong']:
    mutant_paths = []
    for row in fixtures:
        if row['file'].startswith('02_'):
            continue  # These clean programs need no expected-error header.
        text = (HERE / row['file']).read_text()
        header, body = text.split('\n', 1)
        assert header == '// a-check: refused adamic/no-unchecked-cast'
        target = scratch_path / (mutation + '-' + row['file'])
        target.write_text(body if mutation == 'remove' else '// a-check: refused unrelated-rule\n' + body)
        mutant_paths.append(str(target))
    harness = Harness(mutation)
    gate.Gate.aCheck(harness, mutant_paths)
    assert harness.exits['a-check'] == 1, 'header mutants survived'
    assert len(harness.result['a_check']) == 18
    for name, observed in harness.result['a_check'].items():
        assert observed['outcome'] == 'refused'
        assert 'adamic/no-unchecked-cast' in observed['first'] or 'a cast the runtime' in observed['first']
        assert observed['expected'] in ['checked', 'refused unrelated-rule']
        result['mutants'].append({'file': name, 'mutation': mutation, 'caught_by': 'unchanged Gate.aCheck expected-error predicate'})
(HERE / 'a-check-results.json').write_text(json.dumps(result, indent=2) + '\n')
print(json.dumps({'a_check_passed': 20, 'header_mutants_caught': len(result['mutants'])}))
