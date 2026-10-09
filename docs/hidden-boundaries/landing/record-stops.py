#!/usr/bin/env python3
"""Record the first observed boundary, keeping dependencies and checker exclusions distinct."""
import json, sys
from pathlib import Path
raw, regions, output = map(Path, sys.argv[1:])
rows = [json.loads(line) for line in raw.read_text().splitlines()]
root = rows[1]['file'].split('/src/compiler/')[0]+'/src/compiler/'
counts = json.loads(regions.read_text())
result=[]
for region in counts['regions']:
 row = next(row for row in rows if row.get('file') == root+region['file'])
 left,right = region['start'],region['end']
 candidates=[]
 for finding in row['findings']:
  if finding.get('kind') == 'Boundary' and finding['start'] < right and finding['end'] > left:
   candidates.append(dict(start=finding['start'],end=finding['end'],owner='lowering',boundary=finding['where'],diagnostic=finding['text'],raising=('overloadDirectUses' if 'an indirect value of an overload' in finding['text'] else 'property / notYet' if 'union or optional field read in a program with record storage' in finding['text'] else 'censusOverload')))
 for unit in row['units']:
  if unit.get('status') == 'split_checker_body' and unit['body_start'] < right and unit['body_end'] > left:
   candidates.append(dict(start=unit['body_start'],end=unit['body_end'],owner='checker',boundary=unit['where'],diagnostic=unit['checker_diagnostics'][0],raising=('latentFullSelected / LatentOwnDiagnosticsIn; checker checkNonNullTypeWithReporter / reportObjectPossiblyNullOrUndefinedError' if 'TS18048' in unit['checker_diagnostics'][0] else 'latentFullSelected / LatentOwnDiagnosticsIn; checker isSignatureApplicable')))
 inside=[candidate for candidate in candidates if left <= candidate['start'] < right]
 if inside: chosen=min(inside,key=lambda c:(c['start'],c['end']-c['start']))
 elif candidates: chosen=min(candidates,key=lambda c:c['end']-c['start'])
 else: chosen=dict(start=left,end=right,owner='none',boundary='none inside region',diagnostic='All selected independent units are covered; no remaining boundary inside the interval.',raising='none')
 result.append(dict(region=region['label'],hidden_before=region['before']['hidden_bytes'],hidden_after=region['after']['hidden_bytes'],revealed=region['revealed_bytes'],**chosen))
output.write_text(json.dumps(result,indent=2)+'\n')
for row in result: print(row['region'],row['revealed'],row['boundary'],row['diagnostic'])
