#!/usr/bin/env python3
"""Intersect pinned hidden.py's exact union/subtraction result with assigned regions."""
import argparse
import hashlib
import importlib.util
import json
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument('source_pin', type=Path)
parser.add_argument('adapted', type=Path)
parser.add_argument('before', type=Path)
parser.add_argument('after', type=Path)
parser.add_argument('output', type=Path)
a = parser.parse_args()
spec = importlib.util.spec_from_file_location('pinned_hidden', a.source_pin / 'stage3/census/hidden/hidden.py')
hidden = importlib.util.module_from_spec(spec)
spec.loader.exec_module(hidden)
stock = hidden.read_json(a.source_pin / 'stage3/census/hidden/evidence/stock.json.gz')
for name, metadata in stock.items():
 data = (a.adapted / name).read_bytes()
 assert len(data) == metadata['bytes'] and hashlib.sha256(data).hexdigest() == metadata['sha256'], name
regions = [('hidden-01','transformers/declarations.ts',68180,81805), ('hidden-05','transformers/es2018.ts',35690,41768), ('hidden-06','transformers/esDecorators.ts',61324,72741)]
selected = {name:stock[name] for _,name,_,_ in regions}
results = {}
for side, path in [('before',a.before),('after',a.after)]:
 rows = hidden.read_rows(path)
 computed = hidden.calculate(rows, selected, a.adapted)
 results[side] = computed
output = {'source_pin':'388096e6a83a4e9d287fb827f793c599ba1bf0ad','delivery_base':'4885cec50290686df487b62aac47c85d871ed40c','all_adapted_hashes_verified':len(stock),'scope':'All independent declarations in the three named files; entire project loaded and registered. No corpus-wide claim.','regions':[]}
for label,name,start,end in regions:
 row = {'label':label,'file':name,'sha256':stock[name]['sha256'],'start':start,'end':end,'bytes':end-start}
 for side in ['before','after']:
  ranges = results[side]['files'][name]['hidden_ranges']
  intersect = [(max(start,left),min(end,right)) for left,right in ranges if left < end and right > start]
  row[side] = {'hidden_bytes':hidden.size(intersect),'ranges':intersect}
 row['revealed_bytes'] = row['before']['hidden_bytes'] - row['after']['hidden_bytes']
 output['regions'].append(row)
a.output.write_text(json.dumps(output,indent=2)+'\n')
for row in output['regions']:
 print(row['label'], row['before']['hidden_bytes'], '->', row['after']['hidden_bytes'], 'revealed',row['revealed_bytes'])
