#!/usr/bin/env python3
"""Target metadata must cover the exact original resolved call coordinates."""
import copy, importlib.util, json, sys
from pathlib import Path
spec = importlib.util.spec_from_file_location('target_join', Path(__file__).with_name('merge-target-metadata.py'))
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)
measured = json.loads(Path(sys.argv[1]).read_bytes())
metadata = json.loads(Path(sys.argv[2]).read_bytes())
module.join(measured, metadata)

def missing_target(rows):
    rows.pop()
def change_coordinate(rows):
    rows[0]['where'] += '-wrong'
def change_boolean(rows):
    rows[0]['viewTarget'] = 'yes'

results = []
for name, mutate, catcher in [
    ('drop target metadata', missing_target, 'target metadata coverage drift'),
    ('change target coordinate', change_coordinate, 'target metadata coordinate drift'),
    ('replace boolean target flag', change_boolean, 'invalid target metadata'),
]:
    rows = copy.deepcopy(metadata)
    mutate(rows)
    try:
        module.join(measured, rows)
    except AssertionError as failure:
        assert catcher in str(failure), (name, str(failure))
        results.append({'mutant': name, 'catcher': str(failure)})
    else:
        raise AssertionError('mutant escaped: ' + name)
print(json.dumps(results, indent=2))
