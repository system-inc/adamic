"""Check the first-read table against pinned observations and stock source bytes."""
import collections
import hashlib
import json
import pathlib
import subprocess
import sys

root = pathlib.Path(__file__).resolve().parents[4]
report = json.loads(pathlib.Path(sys.argv[1]).read_text())
stock = pathlib.Path(sys.argv[2])
prefix = 'stage3/scouts/step09/nonnull/'
raw = subprocess.check_output(['git', 'show', report['measurement_commit'] + ':' + prefix + 'slot-observations.json'], cwd=root)
assert hashlib.sha256(raw).hexdigest() == report['observation_sha256']
expected = {r['site']: r for r in json.loads(raw) if r['before_write']}
assert len(report['slots']) == len(expected) == 19
assert {r['placeholder'] for r in report['slots']} == set(expected)
for row in report['slots']:
    measured = expected[row['placeholder']]
    assert row['slot'] == measured['slot']
    assert row['before_write_reads'] == measured['before_write']
    assert row['after_write_first_reads'] == measured['after_write']
    assert row['before_write_inputs'] == measured['before_write_inputs']
    assert {r['location'] + '|before-write|' + r['node_value']: r['count'] for r in row['first_reads_before_write']} == {k: v for k, v in measured['locations'].items() if '|before-write|' in k}
    for read in row['first_reads_before_write']:
        assert read['node_value'] == 'undefined'
        file, line, column = read['location'].rsplit(':', 2)
        data = (stock / file).read_bytes()
        assert hashlib.sha256(data).hexdigest() == report['source_files'][file]
        assert read['source_line'] == data.decode().splitlines()[int(line) - 1].strip()
        assert read['category'] in {'presence', 'forwarded-presence', 'save-restore', 'copy'}
assert sum(r['before_write_reads'] for r in report['slots']) == report['before_write_reads'] == 344015
assert report['distinct_read_locations'] == len({r['location'] for s in report['slots'] for r in s['first_reads_before_write']}) == 32
print('19 slots and 344015 reads match pinned Node observations; 32 source locations match pinned bytes')
