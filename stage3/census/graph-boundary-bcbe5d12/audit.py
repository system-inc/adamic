"""Independent raw-stream recount, source checks and final evidence scope checks."""
import collections,gzip,hashlib,json,subprocess
from pathlib import Path
here=Path(__file__).resolve().parent;root=here.parents[2];raw=here/'raw'
def read(p):return [json.loads(line) for line in gzip.decompress(p.read_bytes()).decode().splitlines()]
header=read(raw/'first-latent.jsonl.gz')[0]
assert header['roots']==['/workspace/graph-boundary-adapted/src/tsc/tsc.ts'] and header['latent_mode']=='first-error'
assert len(header['files'])==81
for path in raw.glob('full-*-latent.jsonl.gz'):
 records=read(path)
 if not records:continue
 h=records[0]
 assert h['roots']==header['roots'] and h['files']==header['files'] and h['diagnostics']==header['diagnostics']
 assert h['latent_mode']=='full'
first=read(raw/'first-boundary.jsonl.gz');full=[]
source_streams={(r['file'],r['unit']):'first-boundary.jsonl' for r in first}
for p in sorted(raw.glob('full-*-boundary.jsonl.gz')):
 part=read(p);full.extend(part)
 for r in part:source_streams[r['file'],r['unit']]=p.name[:-3]
seen=set()
for r in full:
 key=(r['file'],r['unit']);assert key not in seen,key;seen.add(key)
units={(r['file'],r['unit']):r for r in first}
for r in full:units[r['file'],r['unit']]=r
counts=collections.Counter();distinct=collections.defaultdict(set);statuses=collections.Counter();by_scope=collections.defaultdict(collections.Counter)
for r in units.values():
 assert r['graph_types_nonempty']==(r['graph_types_count']>0)
 statuses[r['status']]+=1
 if r['status']=='stopped' and not r['full_mode']:assert r['sites']==[]
 for s in r['sites']:
  assert s['where']==r['unit'] and 'fallback' in s['position_source']
  path,line,column=s['where'].rsplit(':',2);assert Path(path).is_file() and int(line)>0 and int(column)>0
  counts[s['kind']]+=1;distinct[s['kind']].add(s['where'])
  by_scope['compiler_79' if '/src/compiler/' in r['file'] else 'tsc_entry_2'][s['kind']]+=1
flat_counts=collections.Counter();flat_seen=set()
with gzip.open(here/'sites.jsonl.gz','rt') as stream:
 for line in stream:
  site=json.loads(line);key=(site['file'],site['unit']);row=units[key];index=site['observation_index']
  identity=(*key,index);assert identity not in flat_seen;flat_seen.add(identity)
  for name,value in row['sites'][index].items():assert site[name]==value
  assert site['source_stream']==source_streams[key] and site['full_mode']==row['full_mode']
  assert site['graph_types_nonempty']==row['graph_types_nonempty']
  flat_counts[site['kind']]+=1
assert flat_counts==counts
summary=json.loads((here/'summary.json').read_text())
for kind,count in summary['counts'].items():assert counts[kind]==count
for kind,count in summary['distinct_fallback_locations'].items():assert len(distinct[kind])==count
assert len(units)==summary['observed_unit_results']
assert dict(statuses)==summary['observed_statuses']
assert sum(r['graph_types_nonempty'] for r in units.values())==summary['graph_types_nonempty_units']
manifest=json.loads((here/'source-manifest.json').read_text())
for name,digest in manifest.items():assert hashlib.sha256((Path('/workspace/graph-boundary-adapted')/name).read_bytes()).hexdigest()==digest
scope='stage3/census/graph-boundary-bcbe5d12/'
changed=subprocess.check_output(['git','diff','--name-only','bcbe5d12'],cwd=root,text=True).splitlines()
assert all(name.startswith(scope) for name in changed),changed
for p in here.rglob('*'):assert '__pycache__' not in p.parts
result=dict(raw_recount_passed=True,source_hashes_passed=True,scope_passed=True,counts_by_scope={k:dict(v) for k,v in by_scope.items()},units=len(units))
(here/'validation.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
