"""Check status coverage and Node observations, including two source mutants."""
import json
import subprocess
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parent
REPO = ROOT.parents[2]
rows = json.loads((ROOT / 'status.json').read_text())
assert {r['file'] for r in rows} == {f.name for f in ROOT.glob('*.a')}
assert len(rows) == len({r['file'] for r in rows})
for row in rows:
    assert set(row) == {'file', 'tsc', 'reason', 'node', 'stage0'}
    assert set(row['node']) == {'stdout', 'stderr', 'exit'}
    assert set(row['stage0']) == {'outcome', 'what'}
    assert row['stage0']['outcome'] in {'NotYet', 'Refused', 'Checker', 'Compiles'}
    run = subprocess.run(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(ROOT / row['file'])], cwd=REPO, capture_output=True, text=True)
    assert (run.stdout, run.stderr, run.returncode) == (row['node']['stdout'], row['node']['stderr'], row['node']['exit']), row['file']
print('PASS:', len(rows), 'fixture observations and exact status schema')

by_name = {row['file']: row for row in rows}
mutants = [
    ('02_get_property.a', 'hasOwnProperty.call(map, key) ? map[key] : undefined', 'map[key]', 'remove own-property guard'),
    ('05_integer_order.a', 'return keys;', 'return keys.reverse();', 'reverse key enumeration'),
    ('17_compare_missing_object.a', 'const src = { other: { value: 1 } };', 'const src = { [key]: { value: 1 } };', 'replace missing object key with own key'),
    ('18_compare_missing_scalar.a', 'const src = { other: 1 };', 'const src = { [key]: 1 };', 'replace missing scalar key with own key'),
    ('19_compare_empty_objects.a', 'for (const e in dst) {', 'for (const e in dst) {\n        if (!Object.prototype.hasOwnProperty.call(src, e)) return false;', 'check ownership before the empty-object comparison'),
]
with tempfile.TemporaryDirectory(prefix='records-mutants-') as scratch:
    for name, before, after, label in mutants:
        source = (ROOT / name).read_text()
        assert source.count(before) == 1
        path = Path(scratch) / name
        path.write_text(source.replace(before, after))
        run = subprocess.run(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(path)], cwd=REPO, capture_output=True, text=True)
        assert run.returncode == 0, run.stderr
        assert run.stdout != by_name[name]['node']['stdout'], label
        print('CAUGHT by stdout: ' + label)
print('Runtime panic and ownership mutants are not claimed: stage 0 does not compile these inputs.')
