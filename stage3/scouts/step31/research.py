#!/usr/bin/env python3
"""Attribute fresh meter sites to each piece and its gathered declaration closure."""
import argparse
from collections import Counter, defaultdict
import gzip
import hashlib
import json
from pathlib import Path
import re

parser = argparse.ArgumentParser()
parser.add_argument('meter', type=Path)
parser.add_argument('tree', type=Path)
parser.add_argument('slices', type=Path, help='parent of step31-{binder,checker,emitter}-slice')
parser.add_argument('output', type=Path)
args = parser.parse_args()
if args.output.exists():
    parser.error('output must be new')
args.output.mkdir(parents=True)

def load_rows(path):
    if path.exists():
        return [json.loads(line) for line in path.read_text().splitlines()]
    with gzip.open(str(path) + '.gz', 'rt') as stream:
        return [json.loads(line) for line in stream]

def relative(value):
    return value.replace(str(args.tree.resolve()) + '/', '')

meter = json.loads((args.meter / 'main/report.json').read_text())
ordinary = load_rows(args.meter / 'main/census.jsonl')
latent = load_rows(args.meter / 'main/latent.jsonl')
files = {relative(row['file']): row for row in latent[1:]}
checker = {row['file']: row for row in meter['files']}
output = {'compiler_commit': meter['adamic_commit'], 'tree_commit': meter.get('tree_commit'),
    'meter_totals': meter['totals'], 'measurement': latent[0]['measurement'], 'pieces': {}}
all_sites = set()
for row in files.values():
    all_sites.update((f['kind'], relative(f['where']), f.get('reason', ''), relative(f.get('text', ''))) for f in row['findings'])
output['latent_totals'] = dict(Counter(site[0] for site in all_sites))
scout = Path(__file__).resolve().parent
static = json.loads((scout.parents[1] / 'census/data/rankings.json').read_text())
for piece in ('binder', 'checker', 'emitter'):
    name = f'src/compiler/{piece}.ts'
    manifest = json.loads((args.slices / f'step31-{piece}-slice/slice.json').read_text())
    code_files = sorted({d['file'] for d in manifest['declarations'] if d['kind'] != 'ExportDeclaration'})
    groups = defaultdict(set)
    row = files[name]
    sites = {site for site in all_sites if site[1].rsplit(':', 2)[0] == name}
    for kind, where, reason, text in sites:
        if kind in ('NotYet', 'Refused'):
            groups[(kind, reason)].add(where)
    families = sorted([{'kind': kind, 'reason': reason,
        'count': sum(1 for site in sites if site[0] == kind and site[2] == reason),
        'locations': sorted(locations, key=lambda value: (value.rsplit(':', 2)[0], int(value.rsplit(':', 2)[1]), int(value.rsplit(':', 2)[2])))} for (kind, reason), locations in groups.items()],
        key=lambda item: (-item['count'], item['kind'], item['reason']))
    skipped = []
    for file in code_files:
        for unit in files[file]['units']:
            if unit['status'] != 'skipped_checker_body':
                continue
            source = (args.tree / file).read_bytes()
            body = source[unit['body_start']:unit['body_end']]
            declaration = next((d for d in files[file]['declarations'] if d['where'] == unit['where']), {})
            skipped.append({'file': file, 'where': relative(unit['where']), 'name': declaration.get('name', '').strip(),
                'bytes': len(body), 'lines': body.count(b'\n') + 1, 'diagnostics': len(unit['checker_diagnostics']),
                'reached_in_slice': any(d['file'] == file and declaration.get('name', '').strip() in d['names']
                    for d in manifest['declarations'])})
    closure_sites = {site for site in all_sites if site[1].rsplit(':', 2)[0] in code_files}
    own_diags = sorted({relative(diagnostic) for record in ordinary for diagnostic in record.get('diagnostics', [])
        if diagnostic.startswith(str(args.tree.resolve() / name) + ':')})
    old_families = []
    for item in static:
        locations = [location for location in item['locations'] if location[0] == name]
        if locations:
            old_families.append({'reason': item['reason'], 'layer': item['layer'],
                'sites_in_file': len(locations), 'locations': locations})
    old_families.sort(key=lambda item: (-item['sites_in_file'], item['reason']))
    output['pieces'][piece] = {'source_sha256': hashlib.sha256((args.tree / name).read_bytes()).hexdigest(),
        'meter': checker[name], 'own_diagnostics': own_diags, 'latent_counts': dict(Counter(site[0] for site in sites)),
        'families': families, 'slice_summary': manifest['summary'], 'code_files': code_files,
        'closure_meter': [checker[file] for file in code_files],
        'closure_latent_full_file_counts': dict(Counter(site[0] for site in closure_sites)),
        'hidden_body_ranking': sorted(skipped, key=lambda item: (-item['lines'], item['where'])),
        'historical_original_source_ranking': old_families}
(args.output / 'research.json').write_text(json.dumps(output, indent=2) + '\n')
print(json.dumps({'meter_totals': output['meter_totals'], 'latent_totals': output['latent_totals'],
    'pieces': {piece: {'meter': data['meter'], 'latent': data['latent_counts'],
        'families': [{k:v for k,v in family.items() if k!='locations'} | {'example':family['locations'][0]} for family in data['families'][:7]],
        'hidden_top': data['hidden_body_ranking'][:3], 'closure_files':len(data['code_files'])}
        for piece, data in output['pieces'].items()}}, indent=2))
