#!/usr/bin/env python3
"""Independently recount evidence and reject report mutants."""
import argparse
from collections import Counter
import gzip
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent
parser = argparse.ArgumentParser()
parser.add_argument('tree', type=Path)
parser.add_argument('--mutant', choices=['meter', 'family', 'hidden', 'attribution'])
args = parser.parse_args()

def read(name):
    with gzip.open(ROOT / 'evidence' / name, 'rt') as stream:
        return json.load(stream)

report = read('research.json.gz')
meter = read('main-meter.json.gz')
with gzip.open(ROOT / 'evidence/main-latent.jsonl.gz', 'rt') as stream:
    records = [json.loads(line) for line in stream]
prefix = meter['adapted_tree'] + '/'
def relative(value):
    return value.removeprefix(prefix)
site_set = {(finding['kind'], relative(finding['where']), finding.get('reason', ''), relative(finding.get('text', '')))
    for record in records[1:] for finding in record['findings']}
if args.mutant == 'meter':
    report['meter_totals']['checker_own_file'] += 1
elif args.mutant == 'family':
    report['pieces']['binder']['families'][0]['count'] += 1
elif args.mutant == 'hidden':
    report['pieces']['checker']['hidden_body_ranking'][0]['lines'] += 1
elif args.mutant == 'attribution':
    report['pieces']['binder']['latent_counts']['NotYet'] += 1
assert report['meter_totals'] == meter['totals'], 'own-file meter recount'
assert report['latent_totals'] == dict(Counter(site[0] for site in site_set)), 'unique latent totals'
for piece, row in report['pieces'].items():
    file = f'src/compiler/{piece}.ts'
    assert hashlib.sha256((args.tree / file).read_bytes()).hexdigest() == row['source_sha256'], 'source hash'
    expected = {site for site in site_set if site[1].rsplit(':', 2)[0] == file}
    assert row['latent_counts'] == dict(Counter(site[0] for site in expected)), 'actual file attribution'
    for family in row['families']:
        matching = {site for site in expected if site[0] == family['kind'] and site[2] == family['reason']}
        assert family['count'] == len(matching), 'family unique-site recount'
        assert set(family['locations']) == {site[1] for site in matching}, 'family location recount'
    manifest = read(piece + '-slice.json.gz')
    assert row['code_files'] == sorted({d['file'] for d in manifest['declarations'] if d['kind'] != 'ExportDeclaration'}), 'code closure'
    for hidden in row['hidden_body_ranking']:
        record = next(r for r in records[1:] if relative(r['file']) == hidden['file'])
        unit = next(u for u in record['units'] if relative(u['where']) == hidden['where'])
        assert unit['status'] == 'skipped_checker_body', 'hidden body eligibility'
        body = (args.tree / hidden['file']).read_bytes()[unit['body_start']:unit['body_end']]
        assert hidden['bytes'] == len(body) and hidden['lines'] == body.count(b'\n') + 1, 'hidden source extent'
        reached = any(d['file'] == hidden['file'] and hidden['name'] in d['names'] for d in manifest['declarations'])
        assert hidden['reached_in_slice'] == reached, 'actual reached declaration'
print('PASS: meter, site attribution, refusal families, closure and hidden-source byte extents')
