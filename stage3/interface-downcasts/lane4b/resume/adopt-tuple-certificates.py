"""Audit four original tuple overlaps; direct cast admission stays separate."""
import hashlib
import json
from pathlib import Path
import sys

lane = Path(__file__).resolve().parent.parent
original = lane / 'original'
owner = lane.parent / 'tuples'
manifest = json.loads((original / 'declaration-manifest.json').read_text())
local_path = original / 'certification.json'
local = json.loads(local_path.read_text())
candidate = json.loads((owner / 'candidate-pairs.json').read_text())
certificate = json.loads((owner / 'certifications.json').read_text())
assert manifest['upstream_commit'] == local['upstream_commit'] == '050880ce59e30b356b686bd3144efe24f875ebc8'

def key(row):
    return row['type_id'], row['field']

def sites(row):
    return sorted((s['file'], s['start'], s['end'], s['text']) for s in row['sites'])

rows = []
for identity in (97898, 97913, 97931, 97934):
    demand = next(row for row in manifest['pairs'] if row['type_id'] == identity)
    source = next(row for row in candidate['pairs'] if key(row) == key(demand))
    certified = next(row for row in certificate['certified'] if row['type_id'] == identity)
    assert demand['read_count'] == source['read_count'] == certified['reads']
    assert sites(demand) == sites(source)
    rows.append(dict(type_id=identity, type=demand['type'], field=demand['field'],
                     candidate_reads=demand['read_count'], certificate='../../tuples/certifications.json'))
assert len(rows) == 4 and sum(row['candidate_reads'] for row in rows) == 7
hashes = {'../../tuples/' + name: hashlib.sha256((owner / name).read_bytes()).hexdigest()
          for name in ('candidate-pairs.json', 'certifications.json')}
if sys.argv[1:] == ['--update']:
    prior = {key(row) for row in local['pairs']}
    local['pairs'].extend(row for row in rows if key(row) not in prior)
    local.setdefault('shared_certificates', {}).update(hashes)
    local_path.write_text(json.dumps(local, indent=2) + '\n')
elif sys.argv[1:]:
    raise SystemExit('usage: adopt-tuple-certificates.py [--update]')
for row in rows:
    assert next(value for value in local['pairs'] if key(value) == key(row)) == row
for name, digest in hashes.items():
    assert local['shared_certificates'][name] == digest
assert len({key(row) for row in local['pairs']}) == len(local['pairs']) == 37
assert sum(row['candidate_reads'] for row in local['pairs']) == 165
print('Exact tuple overlap: four pairs / seven reads; unique lane total: 37 / 165')
