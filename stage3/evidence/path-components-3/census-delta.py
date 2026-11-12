import collections,json,sys
from pathlib import Path
sys.dont_write_bytecode=True
sys.path.insert(0,'/workspace/adamic/stage3/census/latent')
from stream_records import census_records
from unit_status import unit_counts
out=Path(__file__).parent
sets={};summary={}
for label,tree in [('before','/workspace/scratch/path-components-upstream'),('after','/workspace/scratch/path-components-after')]:
 def normalize(x):
  if isinstance(x,str):return x.replace(tree+'/', '')
  if isinstance(x,list):return [normalize(v) for v in x]
  if isinstance(x,dict):return {k:normalize(v) for k,v in x.items()}
  return x
 raw=census_records([normalize(json.loads(s)) for s in (out/('census-'+label+'.jsonl')).read_text().splitlines()])
 header=raw[0];assert header['status']=='measurement';assert header['latent_mode']=='full';assert not header['checker_rejected'];assert not header['diagnostics']
 files=[r for r in raw[1:] if 'file' in r]
 findings=[f for row in raw for f in row.get('findings',[])]
 sites={(f['kind'],f['where'],f['reason'],f['text']) for f in findings}
 assert not any(s[0] in ['error','panic'] for s in sites)
 sets[label]=sites
 counts=collections.Counter(s[0] for s in sites)
 units=[u for r in files for u in r.get('units',[])]
 own={s for s in sites if s[1].startswith('src/compiler/path.ts:')}
 summary[label]=dict(uniqueSites=len(sites),counts=dict(sorted(counts.items())),pathSites=len(own),pathCounts=dict(sorted(collections.Counter(s[0] for s in own).items())),events=len(findings),files=[r['file'] for r in files],units=unit_counts(units),checkerDiagnostics=0)
 def rows(sites):return [dict(kind=k,where=w,reason=r,text=t) for k,w,r,t in sorted(sites)]
 (out/('sites-'+label+'.json')).write_text(json.dumps(rows(sites),indent=2)+'\n')
removed=sets['before']-sets['after'];added=sets['after']-sets['before']
summary.update(removed=rows(removed),added=rows(added),delta=summary['after']['uniqueSites']-summary['before']['uniqueSites'],countDefinition='unique (kind, where, reason, text) across all attempts, with tree prefix removed')
(out/'census-delta.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps(summary,indent=2))
