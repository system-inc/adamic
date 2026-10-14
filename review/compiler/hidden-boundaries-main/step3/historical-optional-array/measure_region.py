"""Measure only the assigned, hash-pinned interval using the pinned hidden census."""
import gzip
import hashlib
import importlib.util
import json
from pathlib import Path
import sys

pin, compiler, ledger, output = map(Path, sys.argv[1:])
start, end = 620304, 627479
directory = pin / 'stage3/census/hidden'
spec = importlib.util.spec_from_file_location('pinned_hidden', directory / 'hidden.py')
hidden = importlib.util.module_from_spec(spec)
spec.loader.exec_module(hidden)
stock = hidden.read_json(directory / 'evidence/stock.json.gz')
actual = {path.relative_to(compiler).as_posix() for path in compiler.rglob('*') if path.is_file()}
assert actual == set(stock), 'incomplete adapted source'
for name, metadata in stock.items():
    data = (compiler / name).read_bytes()
    assert len(data) == metadata['bytes'] and hashlib.sha256(data).hexdigest() == metadata['sha256'], name

catalog = stock['checker.ts']['units']
overlap = {unit['where'] for unit in catalog if unit['start'] < end and start < unit['end']}
assert overlap == {'checker.ts:1486:1', 'checker.ts:6309:5', 'checker.ts:9346:9', 'checker.ts:10648:13'}, overlap

def clip(spans):
    return hidden.union((max(left, start), min(right, end)) for left, right in spans if left < end and start < right)

def window(rows):
    root = Path(rows[1]['file']).parent
    got = {hidden.local_where(unit['where'], root) for unit in rows[1]['units']}
    assert got == overlap, (got, overlap)
    result = hidden.calculate(rows, {'checker.ts': stock['checker.ts']}, root)
    return clip(result['files']['checker.ts']['hidden_ranges'])

historical = hidden.read_rows(directory / 'evidence/full.jsonl.gz')
old_record = next(row for row in historical[1:] if row['file'].endswith('/checker.ts'))
old_root = Path(old_record['file']).parent
selected = dict(old_record)
selected['units'] = [unit for unit in old_record['units'] if hidden.local_where(unit['where'], old_root) in overlap]
selected['findings'] = [finding for finding in old_record['findings']
    if finding.get('unit') and hidden.local_where(finding['unit'], old_root) in overlap]
old = window([historical[0], selected])
recorded = hidden.read_json(directory / 'RESULT.json')
assert old == clip(recorded['files']['checker.ts']['hidden_ranges']), 'partial interval differs from historical full census'
assert hidden.size(old) == 7175
rows = hidden.read_rows(ledger)
assert rows[0]['region_start'] == start and rows[0]['region_end'] == end
assert len(rows) == 2 and rows[1]['file'].endswith('/checker.ts')
new = window(rows)
result = dict(region={'file':'checker.ts', 'start':start, 'end':end, 'source_sha256':stock['checker.ts']['sha256']},
    old_hidden_intersection=old, old_hidden_bytes=hidden.size(old),
    new_hidden_intersection=new, new_hidden_bytes=hidden.size(new),
    byte_difference=hidden.size(old)-hidden.size(new),
    overlapping_units=sorted(overlap), adapted_files_verified=len(stock),
    historical_partial_matches_full=True,
    ledger_sha256=hashlib.sha256(ledger.read_bytes()).hexdigest(),
    scope='Assigned interval only. No whole-file, group, or corpus total is measured.')
output.write_text(json.dumps(result, indent=2)+'\n')
print(json.dumps(result, indent=2))
