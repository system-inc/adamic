import collections,json
from pathlib import Path
out=Path(__file__).resolve().parent
old=b'const changedFilesInLastPoll: WatchedFileWithUnchangedPolls[] = [];';new=b'const changedFilesInLastPoll: (WatchedFileWithUnchangedPolls | undefined)[] = [];'
source=(out/'main-tree/src/compiler/sys.ts').read_bytes();start=source.index(old);delta=len(new)-len(old)
def rows(label):
 raw=[json.loads(s) for s in (out/(label+'.jsonl')).read_text().splitlines()]
 assert raw[0]['diagnostics']==[] and raw[0]['checker_rejected']==False
 assert raw[0]['module_format']=='CommonJSBundle' and raw[0]['latent_mode']=='full'
 assert raw[-1]['status']=='relation_report'
 def norm(v):
  if isinstance(v,str):return v.replace(str(out/(label+'-tree')),'<TREE>')
  if isinstance(v,list):return [norm(x) for x in v]
  if isinstance(v,dict):return {k:norm(x) for k,x in v.items()}
  return v
 fs=[]
 for row in raw:
  for finding in row.get('findings',[]):
   f=norm(finding)
   if label=='after':
    for a,b in [('site_start','site_where'),('site_end','site_where'),('start','where'),('end','where')]:
     if a in f and '/src/compiler/sys.ts:' in f.get(b,''):
      if f[a]>=start+len(new):f[a]-=delta
      elif f[a]>start:raise AssertionError(('span inside edited declaration',f))
   fs.append(f)
 return raw,fs
def key(f):return tuple(f.get(k,'') for k in ['kind','where','reason','text'])
def diff(a,b):
 da={key(f):f for f in a};db={key(f):f for f in b}
 return dict(gone=[da[k] for k in sorted(da.keys()-db.keys())],new=[db[k] for k in sorted(db.keys()-da.keys())],unchanged=len(da.keys()&db.keys()))
def observations(a,b):
 ca=collections.Counter(json.dumps(f,sort_keys=True) for f in a);cb=collections.Counter(json.dumps(f,sort_keys=True) for f in b)
 return dict(gone=[dict(count=n,finding=json.loads(k)) for k,n in sorted((ca-cb).items())],new=[dict(count=n,finding=json.loads(k)) for k,n in sorted((cb-ca).items())])
ra,a=rows('main');rb,b=rows('after')
report={'normalization':'Only tree-path prefixes and the exact 14-byte sys.ts span relocation after the declaration; no reason, text, unit or phase is ignored. Boundary events listed separately.', 'source_byte_delta':delta,'raw_counts':{'main':len(a),'after':len(b)},'sites':diff([f for f in a if f['kind'] in ['Refused','NotYet']],[f for f in b if f['kind'] in ['Refused','NotYet']]),'all_events':diff(a,b),'observations':observations(a,b)}
if (out/'restoring-mutant-receipt.json').exists():
 rm,m=rows('restoring-mutant');report['mutant_sites_vs_main']=diff(a,m);report['mutant_observations_vs_main']=observations(a,m)
 assert any(f['kind']=='Refused' and f['where'].endswith('sys.ts:290:19') for f in m)
assert any(f['kind']=='Refused' and f['where'].endswith('sys.ts:290:19') for f in a)
assert not any(f['where'].endswith('sys.ts:290:19') for f in b)
(out/'census-delta.json').write_text(json.dumps(report,indent=2)+'\n')
print(json.dumps({k:report[k] for k in ['raw_counts','sites']},indent=2))
