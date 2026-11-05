"""Prepare disjoint first-error scheduling atoms and two guarded selector binaries."""
from collections import Counter
import difflib
import gzip
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess

OUT = Path(__file__).resolve().parent
E = OUT / 'evidence/pieces'
S = OUT / 'scratch'
TREE = S / 'adapted'
E.mkdir(parents=True, exist_ok=True)

def digest(path): return hashlib.sha256(path.read_bytes()).hexdigest()
def write(path, obj): path.write_text(json.dumps(obj, indent=2) + '\n')
with gzip.open(OUT / 'evidence/base-latent.jsonl.gz', 'rt') as f:
    whole = [json.loads(x) for x in f]
with gzip.open(OUT / 'evidence/base-times.jsonl.gz', 'rt') as f:
    times = {x['where']: x for x in map(json.loads, f)}
manifest = json.loads((OUT / 'evidence/source-manifest.json').read_text())
assert manifest == {str(p.relative_to(TREE)): digest(p) for p in sorted((TREE / 'src').rglob('*')) if p.is_file()}
records = {r['file']: r for r in whole[1:]}
assert len(records) == 79
units = {u['where']: u for r in records.values() for u in r['units']}
stock = json.loads((OUT / 'evidence/stock-units.json').read_text())
assert set(units) == {u['where'] for r in stock for u in r['units']}
weight = lambda u: times.get(u['where'], {}).get('wall_seconds', 0)
file_cost = {file: sum(weight(u) for u in r['units']) + 30 for file, r in records.items()}
raw, ordinary = [], []
for file, record in records.items():
    if file_cost[file] <= 600:
        ordinary.append(file)
        continue
    chunk, cost = [], 0
    for unit in record['units']:
        w = weight(unit)
        if chunk and cost + w > 300:
            raw.append(dict(files=[dict(file=file, all_units=False, units=chunk)], base_unit_seconds=cost, estimated_seconds=cost+60))
            chunk, cost = [], 0
        chunk.append(unit['where']); cost += w
    if chunk:
        raw.append(dict(files=[dict(file=file, all_units=False, units=chunk)], base_unit_seconds=cost, estimated_seconds=cost+60))
bins = []
for file in sorted(ordinary, key=lambda f: (-file_cost[f], f)):
    candidates = [b for b in bins if b['cost'] + file_cost[file] <= 600]
    if candidates:
        bucket = min(candidates, key=lambda b: b['cost'])
    else:
        bucket = dict(files=[], cost=0); bins.append(bucket)
    bucket['files'].append(file); bucket['cost'] += file_cost[file]
for b in bins:
    raw.append(dict(files=[dict(file=f, all_units=True, units=[]) for f in sorted(b['files'])],
                    base_unit_seconds=sum(file_cost[f]-30 for f in b['files']), estimated_seconds=b['cost']))
raw.sort(key=lambda p: (p['files'][0]['file'], p['files'][0]['units'][0] if p['files'][0]['units'] else ''))
coverage = Counter()
scans = Counter()
for number, piece in enumerate(raw):
    piece.update(schema=1, id=f'p{number:02d}')
    expected = []
    for selected in piece['files']:
        file = selected['file']; scans[file] += 1
        scheduled = [u['where'] for u in records[file]['units']] if selected['all_units'] else selected['units']
        expected += scheduled
    piece['expected_units'] = expected
    piece['expected_timed_units'] = [u for u in expected if u in times]
    coverage.update(expected)
assert coverage == Counter({u: 1 for u in units})
assert set(scans) == set(records)
plan = dict(schema=1, mode='first-error', source_manifest_sha256=digest(OUT / 'evidence/source-manifest.json'),
    base_commit='b0efce6dce1cd7445ed8a570aecac387a4a7d684', head_commit='66fff4ecc05c3bfb2258cf4bb2a7b0a2b762c59b',
    files=79, units=len(units), timed_units=len(times), scan_multiplicity=dict(scans),
    cost_model='base unit wall sums; ordinary files add 30 seconds; files over 600 seconds split into contiguous top-level unit chunks of at most 300 baseline seconds plus 60 seconds setup reserve',
    timeout_seconds=1200, pieces=raw)
plan_path = E / 'plan.json'
if plan_path.exists(): assert json.loads(plan_path.read_text()) == plan
else: write(plan_path, plan)
for piece in raw:
    path = E / f'{piece["id"]}.selection.json'
    if path.exists(): assert json.loads(path.read_text()) == piece
    else: write(path, piece)

builds = {}
for label in ('base', 'head'):
    src, old, dest = S / label, S / f'{label}-bin/overlay', S / f'piece-{label}-bin'
    dest.mkdir(exist_ok=True)
    oldmap = json.loads((old / 'overlay.json').read_text())['Replace']
    newmap = {}
    for original, replacement in oldmap.items():
        target = dest / Path(replacement).name
        shutil.copyfile(replacement, target)
        newmap[original] = str(target)
    hook = dest / 'internal_lower_latent_hook.go'
    text = hook.read_text(); start = text.index('func latentRun(')
    prefix, body = text[:start], text[start:]
    floop = '\tfor _, file := range program.Files() {\n'
    uloop = '\t\tfor _, node := range file.Statements.Nodes {\n'
    assert body.count(floop) == body.count(uloop) == 1
    body = body.replace(floop, floop+'\t\tif !anyMembersCensusSelected(program.FileName(file), "") { continue }\n', 1)
    body = body.replace(uloop, uloop+'\t\t\tif !anyMembersCensusSelected(program.FileName(file), program.Where(node)) { continue }\n', 1)
    hook.write_text(prefix + body)
    (E / f'{label}-selector.patch.txt').write_text(''.join(difflib.unified_diff(text.splitlines(True), (prefix+body).splitlines(True), fromfile='original latent hook', tofile='selected latent hook', n=0)))
    helper = dest / 'any_members_census_piece.go'
    shutil.copyfile(OUT / 'first-error-piece.go.txt', helper)
    newmap[str(src / 'internal/lower/any_members_census_piece.go')] = str(helper)
    write(dest / 'overlay.json', {'Replace': newmap})
    with (E / f'{label}-build.log').open('w') as log:
        subprocess.run(['gofmt', '-w', str(hook), str(helper)], check=True, stdout=log, stderr=subprocess.STDOUT)
        (E / f'{label}-selector.patch.txt').write_text(''.join(difflib.unified_diff(text.splitlines(True), hook.read_text().splitlines(True), fromfile='original latent hook', tofile='selected latent hook', n=0)))
        subprocess.run(['go', 'build', '-buildvcs=false', '-overlay='+str(dest / 'overlay.json'), '-o', str(dest / 'latent-census'), './stage3/census/latent/tool'], cwd=src, check=True, stdout=log, stderr=subprocess.STDOUT)
    builds[label] = dict(compiler_commit=subprocess.check_output(['git','-C',str(src),'rev-parse','HEAD'],text=True).strip(),
        binary_sha256=digest(dest/'latent-census'), helper_sha256=digest(OUT/'first-error-piece.go.txt'),
        original_binary_sha256=digest(S/f'{label}-bin/latent-census'), patch='file and top-level unit selection only inside latentRun; full 79-root program loading and lowering unchanged')
write(E / 'builds.json', builds)
print(json.dumps({'pieces':len(raw),'files':len(scans),'units':len(units),'timed_units':len(times),'max_estimated_seconds':max(p['estimated_seconds'] for p in raw)},indent=2))
