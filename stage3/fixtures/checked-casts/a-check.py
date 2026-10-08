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
# All fixtures now lower. A false refusal header must be rejected for each,
# independently of its Node or native runtime contract.
for mutation, reason in [('false-cast-refusal', 'adamic/no-unchecked-cast'), ('false-unrelated-refusal', 'unrelated-rule')]:
    mutant_paths = []
    for row in fixtures:
        text = (HERE / row['file']).read_text()
        if text.startswith('// a-check:'):
            text = text.split('\n', 1)[1]
        target = scratch_path / (mutation + '-' + row['file'])
        target.write_text('// a-check: refused ' + reason + '\n' + text)
        mutant_paths.append(str(target))
    harness = Harness(mutation)
    gate.Gate.aCheck(harness, mutant_paths)
    assert harness.exits['a-check'] == 1, 'false refusal headers survived'
    assert len(harness.result['a_check']) == 20
    for name, observed in harness.result['a_check'].items():
        assert observed['outcome'] == 'checked'
        assert observed['expected'] == 'refused ' + reason
        result['mutants'].append({'file': name, 'mutation': mutation, 'caught_by': 'unchanged Gate.aCheck expected-error predicate'})
(HERE / 'a-check-results.json').write_text(json.dumps(result, indent=2) + '\n')
print(json.dumps({'a_check_passed': 20, 'header_mutants_caught': len(result['mutants'])}))
