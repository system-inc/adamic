"""Audit exact original-read overlap before recording lane 7 certificates."""
import hashlib
import json
from pathlib import Path
import sys

lane = Path(__file__).resolve().parent.parent
original = lane / 'original'
owner = lane.parent / 'lane7'
relative = '../../lane7/original/certification.json'
certificate_path = owner / 'original/certification.json'
certificate = json.loads(certificate_path.read_text())
assigned = json.loads((owner / 'assigned-lane4b-certificates.json').read_text())
manifest = json.loads((original / 'declaration-manifest.json').read_text())
local_path = original / 'certification.json'
local = json.loads(local_path.read_text())
pin = '050880ce59e30b356b686bd3144efe24f875ebc8'
assert all(value['upstream_commit'] == pin for value in (certificate, assigned, manifest, local))
assert certificate['declarations'] == manifest['declarations']
assert assigned['pairs'] == 7 and assigned['reads'] == 45

def key(row):
    return row['type_id'], row['field']

def sites(row):
    return sorted((site['file'], site['start'], site['end'], site['text'])
                  for site in row['sites'])

rows = []
for row in assigned['rows']:
    demand = next(pair for pair in manifest['pairs'] if key(pair) == key(row))
    certified = next(pair for pair in certificate['pairs'] if key(pair) == key(row))
    assert row['status'] == 'certified overlap' and certified['status'] == 'certified'
    assert demand['read_count'] == certified['read_count'] == row['candidate_reads']
    assert sites(demand) == sites(certified)
    rows.append(dict(type_id=demand['type_id'], type=demand['type'],
                     field=demand['field'], candidate_reads=demand['read_count'],
                     certificate=relative))
assert len(rows) == 7 and sum(row['candidate_reads'] for row in rows) == 45
digest = hashlib.sha256(certificate_path.read_bytes()).hexdigest()

if sys.argv[1:] == ['--update']:
    prior = {key(row) for row in local['pairs']}
    local['pairs'].extend(row for row in rows if key(row) not in prior)
    local.setdefault('shared_certificates', {})[relative] = digest
    local_path.write_text(json.dumps(local, indent=2) + '\n')
elif sys.argv[1:]:
    raise SystemExit('usage: adopt-intersection-certificates.py [--update]')

for expected in rows:
    assert next(row for row in local['pairs'] if key(row) == key(expected)) == expected
assert local['shared_certificates'][relative] == digest
assert len({key(row) for row in local['pairs']}) == len(local['pairs'])
assert len(local['pairs']) >= 33
print('Exact original overlap: seven pairs / 45 reads; unique lane total:',
      len(local['pairs']), '/', sum(row['candidate_reads'] for row in local['pairs']))
