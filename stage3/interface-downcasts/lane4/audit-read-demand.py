#!/usr/bin/env python3
"""Independent span and aggregate audit; this is not allocation-flow analysis."""
import collections, gzip, json, pathlib, sys
root=pathlib.Path(sys.argv[1]); lane=pathlib.Path(sys.argv[2])
def read(name):
 p=lane/name
 return json.loads(gzip.decompress(p.read_bytes()) if p.suffix=='.gz' else p.read_bytes())
rows=read('read-demand-sites.json.gz'); pairs=read('read-demand-pairs.json.gz'); summary=read('read-demand-summary.json')
assert len(rows)==summary['read_sites']==summary['unknown_fallback']
source={}; sites=set(); aggregate=collections.defaultdict(list)
for r in rows:
 if r['file'] not in source: source[r['file']]=(root/r['file']).read_bytes().decode('utf-8').encode('utf-16-le')
 assert source[r['file']][2*r['start']:2*r['end']].decode('utf-16-le')==r['text'],r
 identity=(r['file'],r['start'],r['end'],r['field']); assert identity not in sites,identity
 sites.add(identity); assert r['flow']=='Unknown'
 aggregate[(r['receiver_type_id'],r['field'])].append(r)
assert len(aggregate)==len(pairs)==summary['pairs']
for p in pairs:
 rs=aggregate[(p['receiver_type_id'],p['field'])]
 assert len(rs)==p['reads']
 assert sorted(set(f for r in rs for f in r['families']))==p['families']
 assert sorted(set(r['declared_type'] for r in rs))==p['declared_types']
for f in summary['families']:
 assert f['read_sites']==sum(f['family'] in r['families'] for r in rows)
 assert f['pairs']==sum(f['family'] in p['families'] for p in pairs)
# This receiver is inside a shared helper, with no lexical cast in its body.
assert any(r['file']=='src/compiler/utilities.ts' and r['line']==993 and r['text']=='module.valueDeclaration' and set(r['families'])=={'nullish members','optional properties'} for r in rows),'shared helper read omitted'
print(f"PASS: {len(rows)} exact UTF-16 spans; {len(pairs)} pair counts; family aggregates; shared-helper witness")
