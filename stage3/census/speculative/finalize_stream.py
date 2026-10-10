"""Combine completed whole-project records and resolve depths after the stream.

Usage: finalize_stream.py COMPILER_ROOT RUN_DIRECTORY OUTPUT_DIRECTORY
Independent stock ancestry recounts depths from the complete typed boundary union.
A separate report.py audit rechecks each final depth; the two required stream
mutants must fail before publication. Missing successful records are errors.
"""
from collections import defaultdict
from copy import deepcopy
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys

root, runs, output = (Path(x).resolve() for x in sys.argv[1:4])
output.mkdir(parents=True, exist_ok=True)
identity = json.loads((runs / 'INPUT.json').read_text())
assert identity['root'] == str(root)
records = []
completed = []
incomplete = []
header = None
for item in identity['files']:
    path = root / item['file']
    assert len(path.read_bytes()) == item['bytes']
    assert hashlib.sha256(path.read_bytes()).hexdigest() == item['sha256']
    name = item['file'].replace('/', '__')
    record_path = runs / 'records' / (name + '.jsonl')
    metrics_path = runs / (name + '.metrics.json')
    metrics = json.loads(metrics_path.read_text()) if metrics_path.exists() else None
    if not record_path.exists():
        assert metrics is None or metrics['exit'] != 0, 'completed records dropped: ' + item['file']
        incomplete.append(dict(item, metrics=metrics,
                               reason='no completed measurement' if metrics is None else
                               'RSS limit' if metrics['rss_limit_killed'] else
                               'unit deadline' if metrics.get('unit_deadline_limited') and metrics['exit'] == 124 else
                               'wall timeout' if metrics['exit'] == 124 else 'measurement failed'))
        continue
    assert hashlib.sha256(record_path.read_bytes()).hexdigest() == (record_path.with_suffix('.sha256')).read_text().strip()
    rows = [json.loads(line) for line in record_path.read_text().splitlines()]
    assert len(rows) == 2 and rows[1]['file'] == str(path)
    assert not rows[1].get('piece'), 'partial piece cannot claim complete file coverage'
    if header is None:
        header = rows[0]
    assert header == rows[0], 'whole-project checker header changed'
    records.append(rows[1])
    completed.append(dict(item, metrics=metrics))
assert header is not None, 'no completed files'
header['stream_inventory'] = dict(inputs=identity, completed=completed, incomplete=incomplete)
header['compiler_base'] = '946a8f095a7fa419a92117406314b7b3d44630f0'
header['mapper_fix'] = 'already on main: 6b33ec61da399ae79cfba0900c9f753ed38faa1a equivalent generic binder inference; census copier preserves opaque mapper'

def coverage(rows):
    observed = [str(Path(row['file']).relative_to(root)) for row in rows[1:]]
    expected = [item['file'] for item in rows[0]['stream_inventory']['completed']]
    assert len(observed) == len(set(observed)) and set(observed) == set(expected), 'completed file records dropped'
    whole = {item['file'] for item in identity['files']}
    assert set(expected).isdisjoint(item['file'] for item in incomplete)
    assert whole == set(expected) | {item['file'] for item in incomplete}
    for row in rows[1:]:
        assert row['speculative_coverage']['unvisited_nodes'] == 0
        assert row['speculative_coverage']['source_bytes'] == (Path(row['file'])).stat().st_size


def write(path, rows):
    with path.open('w') as stream:
        for row in rows:
            stream.write(json.dumps(row) + '\n')

raw = output / 'speculative.jsonl'
rows = [header, *records]
coverage(rows)
mutant = deepcopy(rows)
mutant.pop()
try:
    coverage(mutant)
except AssertionError:
    print('dropped-file stream mutant caught by independent completed-file coverage audit', flush=True)
else:
    raise AssertionError('dropped-file stream mutant survived')
write(raw, rows)
scripts = Path(__file__).resolve().parent
with (output / 'stock.log').open('w') as log:
    subprocess.run(['node', str(scripts / 'stock.cjs'), str(root), str(raw), str(output / 'stock.json')],
                   env=dict(os.environ, LATENT_STOCK_ALL_FILES='1'), timeout=90,
                   stdout=log, stderr=subprocess.STDOUT, check=True)
stock = json.loads((output / 'stock.json').read_text())
kind_codes = {name: int(code) for code, names in stock['syntax_kinds'].items() for name in names}
boundaries = defaultdict(set)
for row in records:
    for filename, inventory in row.get('project_failed_boundaries', {}).items():
        for boundary in inventory:
            boundaries[filename].add((boundary['start'], boundary['end'], kind_codes[boundary['kind'].removeprefix('Kind')]))
indices = {}
for file in stock['files']:
    index = defaultdict(list)
    for node in file['nodes']:
        index[(node['start'], node['end'], node['kind_code'])].append(node)
    indices[str(root / file['file'])] = index
changed = 0
matched = 0
for row in records:
    for finding in row['findings']:
        site = finding.get('site_where', '').rsplit(':', 2)[0]
        if site not in indices or not finding.get('site_kind'):
            continue
        key = (finding.get('site_start', 0), finding['site_end'], kind_codes[finding['site_kind'].removeprefix('Kind')])
        candidates = indices[site].get(key, [])
        depths = {sum((*span, kind) in boundaries[site] for span, kind in zip(node['ancestors'], node['ancestor_kinds'], strict=True)) for node in candidates}
        assert len(depths) == 1, (site, key, depths)
        depth = depths.pop()
        changed += finding['depth'] != depth
        finding['depth'] = depth
        matched += 1
write(raw, rows)
# Release the large stock graph before running the independent report auditor.
del stock, indices
with (output / 'report.log').open('w') as log:
    subprocess.run([sys.executable, str(scripts / 'report.py'), str(root), str(raw), str(output / 'stock.json'), str(output)],
                   stdout=log, stderr=subprocess.STDOUT, timeout=90, check=True)
mutant = deepcopy(rows)
finding = next(f for row in mutant[1:] for f in row['findings'] if f.get('site_kind') and f.get('site_where', '').startswith(str(root) + '/'))
finding['depth'] += 1
mutant_path = output / 'shifted-depth-mutant.jsonl'
write(mutant_path, mutant)
with (output / 'shifted-depth-mutant.log').open('w') as log:
    result = subprocess.run([sys.executable, str(scripts / 'report.py'), str(root), str(mutant_path), str(output / 'stock.json'), str(output / 'mutant-report')],
                            stdout=log, stderr=subprocess.STDOUT, timeout=90)
assert result.returncode != 0 and 'AssertionError' in (output / 'shifted-depth-mutant.log').read_text(), 'shifted depth mutant survived'
print('depth shifted by one caught by independent stock ancestry recount', flush=True)
print(json.dumps(dict(completed=len(completed), incomplete=len(incomplete), matched_depths=matched, depths_changed_after_boundary_union=changed)), flush=True)
