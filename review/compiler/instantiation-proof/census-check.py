"""Assert independent generics stay inventoried and concrete callers still lower."""
import json
import sys
from pathlib import Path

records = [json.loads(line) for line in Path(sys.argv[1]).read_text().splitlines()]
units = [unit for record in records for unit in record.get('units', [])]
findings = [finding for record in records for finding in record.get('findings', [])]
expected = {'findBestPatternMatch', 'firstOrUndefined', 'lastOrUndefined', 'forEach', 'length'}
by_name = {unit.get('name'): unit for unit in units}
for name in expected:
    assert name in by_name, ('missing inventory', name)
    assert by_name[name]['status'] == 'deferred_generic_instantiation', by_name[name]
assert not [f for f in findings if f['kind'] == 'NotYet' and ('returning T | undefined' in f['reason'] or 'returning U | undefined' in f['reason'])], findings
print('generic inventory retained; independent bodies deferred')
