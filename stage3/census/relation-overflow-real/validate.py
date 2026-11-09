"""Check the evidence's internal consistency without issuing checker queries."""
import gzip,json,pathlib
root=pathlib.Path(__file__).resolve().parent
read=lambda n:json.loads((root/n).read_text())
names={'subtypeRelation','strictSubtypeRelation','assignableRelation','comparableRelation','identityRelation','enumRelation'}
for mode,phases in [('real',['checked','lowered','units','rendered']),('census',['checked','units','rendered'])]:
 snapshots=[read(mode+'-'+phase+'.json') for phase in phases]
 assert len({x['checker'] for x in snapshots})==1
 for row in snapshots:
  assert set(row['caches'])==names
  assert all(0<=c['overflow']<=c['size'] for c in row['caches'].values())
  assert sum(c['overflow'] for c in row['caches'].values())==len(row['pairs'])
  assert row['RelationOverflowMarked']==bool(row['pairs'])
  assert bool(row['RelationOverflow'])==bool(row['pairs'] or row['TS2859'])
 units=read(mode+'-units.json')
 keys={(p['cache'],p['key']) for p in units['pairs']}
 pairs=read(mode+'-pairs.json')
 assert len(pairs)<=50
 assert all((p['cache'],p['key']) in keys for p in pairs)
 path=root/(mode+'-units.jsonl.gz')
 with gzip.open(path,'rt') as inp:
  records=[json.loads(line) for line in inp]
 assert records[0]['file'].endswith('/src/tsc/tsc.ts'), 'canonical entry must be first'
 work=read(mode+'-work.json')
 assert work['files']==len(records)
 assert work['units']==sum(len(r['units']) for r in records)
 assert all(r['findings'] is not None for r in records)
assert read('real-checked.json')['caches']==read('census-checked.json')['caches']
assert read('real-checked.json')['TS2859']==read('census-checked.json')['TS2859']
for name in ['snapshot.go.txt','main.go.txt','checker-hook.go.txt','prepare.py','run.py','write_report.py','validate.py','REPORT.md']:
 text=(root/name).read_text()
 assert '\u2013' not in text and '\u2014' not in text,name
print('Cache counts, checker identities, pair recovery, workload totals, loader agreement, and text checks pass.')
