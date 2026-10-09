#!/usr/bin/env python3
"""Audit the observation contract and demonstrate that byte comparison can fail."""
import json
from pathlib import Path

root = Path(__file__).resolve().parent
status = json.loads((root / 'status.json').read_text())
evidence = json.loads((root / 'evidence.json').read_text())
mutants = json.loads((root / 'mutants.json').read_text())
manifest = json.loads((root / 'manifest.json').read_text())
assert len(status) == len(manifest) == len(evidence) == 24
assert len({row['file'] for row in status}) == 24
assert [row['file'] for row in status] == [row['file'] for row in manifest]
compiled = []
for row, raw in zip(status, evidence):
    assert set(row) == {'file', 'tsc', 'reason', 'node', 'stage0'}
    assert set(row['node']) == {'stdout', 'stderr', 'exit'}
    assert row['node']['exit'] == 0 and row['node']['stderr'] == ''
    assert set(row['stage0']) == {'outcome', 'what'}
    assert row['stage0']['outcome'] in {'NotYet', 'Refused', 'Checker', 'Compiles'}
    assert raw['file'] == row['file']
    source = (root / row['file']).read_text()
    for location in row['tsc']:
        assert '// From TypeScript 6.0.3, ' + location in source
    for branch in ['main', 'taste']:
        build = raw[branch]['build']
        outcome = row['stage0'] if branch == 'main' else raw['taste_stage0']
        if outcome['outcome'] == 'Compiles':
            assert build['exit'] == 0 and outcome['what'] == ''
            assert raw[branch]['native'] == row['node'], ('silent miscompile', row['file'], branch)
            assert raw[branch]['matches_node'] is True
            compiled.append((row['file'], branch))
        else:
            assert build['exit'] != 0
            assert outcome['what'] == build['stdout'] + build['stderr']
for row in mutants:
    assert row['flips_outcome'] is True, ('masked rewrite', row['form'])
    assert row['same_node'] is True, ('rewrite changed semantics', row['form'])
    if 'native' in row:
        assert row['matches_node'] is True
        assert row['native'] == row['node']
proved = {row['form'] for row in mutants if row['flips_outcome']}
assert {'operand', 'object_condition', 'string_condition', 'double_not', 'logical_assignment', 'comma', 'void', 'label_break', 'named_export', 'star_export'} <= proved
print('24 Node observations; native matches:', compiled)
print('Outcome flips:', sorted(proved))
sample = json.loads((root / 'sample.json').read_text())
assert sample['budget'] == len(sample['sample']) == 25
assert sample['actual'] == sample['quotas'] == {'condition': 19, 'assignment': 2, 'comma': 1, 'label': 1, 'void': 1, 'export': 1}
for site in sample['sample']:
    assert site['fixture'] in {row['file'] for row in status}
print('Primary sample:', sample['actual'])
stress = json.loads((root / 'stress-mutants.json').read_text())
assert len(stress) == 4
for mutant in stress:
    assert mutant['build']['exit'] == 0 and mutant['native']['exit'] == 0
    assert mutant['caught_by_reference_output'] is True
    reference = next(row['node'] for row in status if row['file'] == mutant['file'])
    assert mutant['native'] != reference
assignment = next(row['node']['stdout'] for row in status if row['file'] == '22_assignment_once.a')
assert all(line.startswith('1:1:1:') for line in assignment.splitlines()[:10])
finally_output = next(row['node']['stdout'] for row in status if row['file'] == '23_labels_finally.a')
assert finally_output == 'try:0:0\nfinally:0:0\ntry:0:1\nfinally:0:1\ntry:1:0\nfinally:1:0\nouter-tail:1\ntry:2:0\nfinally:2:0\ndone\n'
print('Stress mutants caught:', [row['name'] for row in stress])
# This exercises the exact equality check used above, without changing saved evidence.
original = evidence[0]['taste']['native']
wrong = {**original, 'stdout': original['stdout'][:-1]}
try:
    assert wrong == status[0]['node'], 'byte comparison rejected a deleted output byte'
except AssertionError as error:
    print('Output mutant caught:', error)
else:
    raise AssertionError('Output mutant escaped')
# Prove the status schema assertion can fail independently of output comparison.
wrong_row = {**status[0], 'taste': {}}
try:
    assert set(wrong_row) == {'file', 'tsc', 'reason', 'node', 'stage0'}, 'schema rejected an extra key'
except AssertionError as error:
    print('Schema mutant caught:', error)
else:
    raise AssertionError('Schema mutant escaped')
