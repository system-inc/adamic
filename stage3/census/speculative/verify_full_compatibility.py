"""Check inherited controls against the untouched candidate and its actual rollback witness."""
import json
from pathlib import Path
import sys

current, base = map(Path, sys.argv[1:3])
def read(directory, name):
    text = (directory/(name+'.jsonl')).read_text()
    source = Path(json.loads(text.splitlines()[1])['file']).parent
    text = text.replace(str(source), '$SOURCE')
    rows = [json.loads(line) for line in text.splitlines()]
    for row in rows[1:]:
        for finding in row['findings']:
            finding.pop('root_kind', None)
            finding.pop('depth', None)
    return rows

def rollback(rows):
    units = {unit['where']:unit.get('name') for row in rows[1:] for unit in row['units']}
    assert any(f['kind']=='NotYet' and f['reason']=='reading poison' and units.get(f['unit'])=='rollback'
               for row in rows[1:] for f in row['findings']), 'rollback witness missing'

for name in ('full', 'first-error-mutant', 'rollback-mutant'):
    assert read(current,name)==read(base,name), name+' differs from untouched candidate'
rollback(read(current,'full'))
for name in ('first-error-mutant', 'rollback-mutant'):
    try:
        rollback(read(current,name))
    except AssertionError:
        print(name+' caught by the actual candidate rollback witness')
    else:
        raise AssertionError(name+' survived')
print('PASS: inherited observations identical to untouched candidate; first-error and retained-state mutants both caught')
