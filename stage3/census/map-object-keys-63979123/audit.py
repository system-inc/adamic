import collections,gzip,hashlib,json,pathlib
p=pathlib.Path(__file__).resolve().parent

def read(name):
 f=p/name
 if not f.exists():f=pathlib.Path(str(f)+'.gz')
 with (gzip.open(f,'rt') if f.suffix=='.gz' else f.open()) as s:return s.read()
def load(name):return json.loads(read(name))
def rows(name):return [json.loads(x) for x in read(name).splitlines()]
def identity(f):return tuple(f[k] for k in ['kind','where','reason','text'])
checks={}
for filename,record in load('source-names.json').items():
 assert hashlib.sha256(pathlib.Path(filename).read_bytes()).hexdigest()==record['sha256'],filename
checks['all_adapted_src_hashes_unchanged']=True
for mode in ['tracing','entry']:
 b,h=(rows(f'{mode}-{s}.jsonl') for s in ['parent','head'])
 assert b[0]==h[0]
 assert b[0]['checker_rejected'] is False
 assert b[0]['latent_mode']==('full' if mode=='tracing' else 'first-error')
 bf={identity(f):f for r in b[1:] for f in r.get('findings',[])}
 hf={identity(f):f for r in h[1:] for f in r.get('findings',[])}
 d=load(f'{mode}-delta.json')['raw_findings']
 assert dict(collections.Counter(f['kind'] for f in bf.values()))==d['totals']['before']
 assert dict(collections.Counter(f['kind'] for f in hf.values()))==d['totals']['after']
 assert len(bf.keys()-hf.keys())==len(d['lowered'])+len(d['moved'])
 assert len(hf.keys()-bf.keys())==len(d['new_findings'])
 assert all(load(f'{mode}-{s}.time.json')['exit']==0 for s in ['parent','head'])
 checks[mode]={'raw_before':len(bf),'raw_after':len(hf),'regressions':len(d['regressions']),'new_sites':len(d['new_sites'])}
 if mode=='tracing':
  removed=[f for k,f in bf.items() if k not in hf and f['kind']!='Boundary']
  assert sorted((f['where'].rsplit(':',2)[1],f['reason']) for f in removed)==sorted([
   ('224',"a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions"),
   ('309',"a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions"),
   ('312',"a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions"),
   ('311','reading recursionIdentityMap')])
  assert not d['new_findings']
inventory=load('entry-inventory.json')
for side in ['parent','head']:
 timings=rows(f'entry-{side}-units.jsonl')
 assert len(timings)==inventory['total_expected_attempts'],(side,len(timings))
 assert len({r['where'] for r in timings})==len(timings)
checks['entry_unit_attempts_each_side']=inventory['total_expected_attempts']
checks['entry_reachable_implementation_sources']=len(rows('entry-parent.jsonl')[0]['sources'])
(p/'audit.json').write_text(json.dumps(checks,indent=2)+'\n')
print(json.dumps(checks,indent=2))
