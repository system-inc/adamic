#!/usr/bin/env python3
"""Recount saved raw reason sites and check all source and outside-file coverage."""
import collections,gzip,json
from pathlib import Path
here=Path(__file__).resolve().parent;e=here/'evidence';r=json.loads((here/'RESULT.json').read_text());m=r['closure'];expected={f['file'] for f in m['files']};root=Path(m['root'])
assert len(expected)==81 and sorted(expected)==m['independent']
assert {f for f in expected if not f.startswith('src/compiler/')}=={'src/tsc/tsc.ts','src/tsc/_namespaces/ts.ts'}
with gzip.open(e/'stock.json.gz','rt') as stream:stock=json.load(stream)
assert set(stock)==expected
for f in m['files']:assert (f['bytes'],f['sha256'])==(stock[f['file']]['bytes'],stock[f['file']]['sha256'])
for mode in ['latent','full']:
 with gzip.open(e/(mode+'.jsonl.gz'),'rt') as stream:rows=[json.loads(l) for l in stream]
 assert rows[0]['checker_rejected'] and {Path(x['file']).relative_to(root).as_posix() for x in rows[1:]}==expected
 keys={(f['kind'],f['where'],f['reason'],f['text']) for row in rows[1:] for f in row['findings'] if f['kind']!='Boundary'}
 actual=dict(sorted(collections.Counter(k[0]+': '+k[2] for k in keys).items()));assert actual==r[mode]['all']['per_reason']
 outside=[k for k in keys if Path(k[1].rsplit(':',2)[0]).is_relative_to(root) and not Path(k[1].rsplit(':',2)[0]).relative_to(root).as_posix().startswith('src/compiler/')]
 assert dict(sorted(collections.Counter(k[0]+': '+k[2] for k in outside).items()))==r[mode]['outside_compiler']['per_reason']
 for group in ['all','compiler','outside_compiler']:
  assert sum(r[mode][group]['counts'].values())==sum(r[mode][group]['per_reason'].values())
h=r['hidden'];assert sum(f['hidden_bytes'] for f in h['files'].values())==h['hidden_bytes']==5430761
assert h['groups']['outside_compiler']['hidden_bytes']==564 and h['groups']['outside_compiler']['bytes']==712
for group in h['groups'].values():assert group['hidden_bytes']==group['blocked_union_bytes']-group['independently_examined_bytes']
p=json.loads((e/'provenance.json').read_text());assert not p['compiler_changes'] and not p['topic_only']
print('PASS: raw unique reason recount, whole closure, outside coverage, byte identities, subtraction and unchanged compiler')
