"""Require intended assertions to kill implementation and report mutants."""
import copy
import json
import os
from pathlib import Path
import subprocess
import sys

raw, stock, result, output = map(Path, sys.argv[1:])
output.mkdir(parents=True, exist_ok=True)
territory = Path(__file__).resolve().parent
checks = [
    ('ignore-dependency', 'test_dependency_skip_exposes_no_expression_body', '0 != 17'),
    ('ignore-skipped', 'test_fake_skipped_span_grows_exactly', '0 != 17'),
    ('double-count', 'test_overlapping_skipped_span_does_not_double_count', '40 != 30'),
    ('forget-independent', 'test_nested_attempt_exposes_only_its_examined_part', '80 != 45'),
    ('forget-checker-child', 'test_attempted_parent_cannot_expose_skipped_child', '0 != 20'),
]
for name, check, catcher in checks:
    log = output / (name + '.log')
    with log.open('w') as stream:
        run = subprocess.run([sys.executable, str(territory / 'test_hidden.py'), 'Arithmetic.' + check],
            env=dict(os.environ, HIDDEN_MUTANT=name), stdout=stream, stderr=stream)
    assert run.returncode == 1 and catcher in log.read_text(), f'{name} was not killed by intended assertion'
    print(f'caught {name}: {catcher}')
original = json.loads(result.read_text())
for name, catcher in [('headline', 'headline hidden total'), ('file', 'file hidden bytes'),
                      ('ranking', 'top ten sizes and ranking'), ('line', 'top ten start line')]:
    mutated = copy.deepcopy(original)
    if name == 'headline':
        mutated['hidden_bytes'] += 17
    elif name == 'file':
        next(iter(mutated['files'].values()))['hidden_bytes'] += 1
    elif name == 'ranking':
        mutated['largest_regions'][0]['bytes'] += 1
    else:
        mutated['largest_regions'][0]['start_line'] += 1
    path, log = output / (name + '.json'), output / (name + '.log')
    path.write_text(json.dumps(mutated))
    with log.open('w') as stream:
        run = subprocess.run([sys.executable, str(territory / 'audit.py'), str(raw), str(stock), str(path)],
            stdout=stream, stderr=stream)
    assert run.returncode == 1 and catcher in log.read_text(), f'{name} corruption was not killed by intended assertion'
    print(f'caught {name} report mutant: {catcher}')
print('PASS: all nine mutants failed their intended assertions')
