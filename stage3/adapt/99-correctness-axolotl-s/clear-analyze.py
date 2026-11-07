import sys
sys.dont_write_bytecode=True
import json,collections
from pathlib import Path
from importlib.machinery import SourceFileLoader
here=Path(__file__).resolve().parent
s=SourceFileLoader('supervise',str(here/'clear-supervise.py')).load_module()
results={}
for label,tree in [('baseline','baseline'),('after','before'),('mutant','mutant')]:
 out=s.scratch/('census-'+label)
 js=(s.scratch/'reference/ast-map.cjs').read_text().replace('/workspace/after-wave-adapted',str(s.scratch/tree)).replace('/workspace/scratch/after-wave',str(out))
 (out/'ast-map.cjs').write_text(js)
 r=s.run('ast-'+label,['node',str(out/'ast-map.cjs')],limit=1800)
 if r['exit_code']!=0:sys.exit(r['exit_code'])
 text=(s.scratch/'reference/analyze.py').read_text().replace("pathlib.Path('/workspace/scratch/after-wave')",'pathlib.Path('+repr(str(out))+')').replace("dest=repo/'stage3/census/core-sites-af42d3dc'",'dest=scratch').replace('/workspace/after-wave-adapted',str(s.scratch/tree))
 exec(compile(text,'reference/analyze.py','exec'),{})
 results[label]=json.loads((out/'rows.json').read_text())
def key(data,f):
 u=next((u for u in data['units'] if u['id']==f['site_unit_id']),None)
 return (u['source_key'] if u else f['site_unit_name'],f['kind'],f['reason'],f['phase'])
def diff(a,b):
 x=collections.Counter(key(a,f) for f in a['sites']);y=collections.Counter(key(b,f) for f in b['sites'])
 return {'gone':[list(k)+[n] for k,n in (x-y).items()],'new':[list(k)+[n] for k,n in (y-x).items()]}
result={'baselineToAfter':diff(results['baseline'],results['after']),'afterToMutant':diff(results['after'],results['mutant']), 'referenceToAfter':diff(json.loads((s.scratch/'reference/rows.json').read_text()),results['after'])}
for label,data in results.items():
 result[label]={'units':len(data['units']),'statuses':dict(collections.Counter(u['reported_status'] for u in data['units'])),'sites':len(data['sites']),'firstWalls':len(data['first_walls']),'diagnostics':data['measurement']['diagnostics'],'clearUnits':[u for u in data['units'] if u['display_name']=='clear'],'clearEvents':[f for f in data['findings'] if f['site_unit_name']=='clear']}
assert len(result['baselineToAfter']['gone'])==1 and not result['baselineToAfter']['new'],result
assert 'erased element storage' in result['baselineToAfter']['gone'][0][2]
assert not result['afterToMutant']['gone'] and result['afterToMutant']['new']==result['baselineToAfter']['gone']
assert not result['referenceToAfter']['new'],result
(here/'clear-census-results.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result,indent=2))
