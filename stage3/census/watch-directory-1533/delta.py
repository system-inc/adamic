import json,re
from pathlib import Path
r=Path(__file__).parent

def norm(x):
 if isinstance(x,str):
  for name in ['baseline','candidate','mutant']: x=x.replace(str(r/name)+'/','<tree>/')
  return x
 if isinstance(x,list):return [norm(i) for i in x]
 if isinstance(x,dict):return {k:norm(v) for k,v in x.items()}
 return x

def read(name):
 rows=[norm(json.loads(l)) for l in (r/(name+'.jsonl')).read_text().splitlines()]
 assert rows[0]['module_format']=='CommonJSBundle' and rows[0]['latent_mode']=='full'
 assert rows[-1]['status']=='relation_report'
 findings=[f for row in rows for f in row.get('findings',[]) or []]
 exact={json.dumps({k:v for k,v in f.items() if k not in ['start','end']},sort_keys=True):f for f in findings}
 return rows,exact
base,a=read('main');change,b=read('change');mutant,m=read('mutant')
new=[b[k] for k in sorted(b.keys()-a.keys())];gone=[a[k] for k in sorted(a.keys()-b.keys())]
# Compare full contextual records after excluding shifted byte offsets; retain original offsets in each reported record and raw streams.
def site(f):return tuple(f.get(k) for k in ['kind','where','reason','text','phase'])
sa={site(f):f for f in a.values()};sb={site(f):f for f in b.values()}
site_new=[sb[k] for k in sorted(sb.keys()-sa.keys(),key=str)];site_gone=[sa[k] for k in sorted(sa.keys()-sb.keys(),key=str)]
# changed means the same location/kind/phase has a different reason or text.
loc=lambda f:(f.get('kind'),f.get('where'),f.get('phase'))
changed=[]
for f in site_gone:
 for g in site_new:
  if loc(f)==loc(g):changed.append({'before':f,'after':g})
result={'module_format':'CommonJSBundle','full':True,'no_output_assertions':True,'contextual_new':new,'contextual_gone':gone,'site_new':site_new,'site_gone':site_gone,'changed':changed,'counts':{'contextual_new':len(new),'contextual_gone':len(gone),'site_new':len(site_new),'site_gone':len(site_gone),'changed':len(changed)},'restoring_mutant_contextual_findings_equal_main':a==m,'targets':{}}
for line in [1532,1533]:
 result['targets'][str(line)]={name:[f for f in items.values() if f.get('where','').endswith(f'sys.ts:{line}:13')] for name,items in [('main',a),('change',b),('mutant',m)]}
(r/'delta.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result['counts']));print('mutant equality:',a==m)
for name,items in [('new',site_new),('gone',site_gone)]:
 print(name,json.dumps(items,indent=2))
