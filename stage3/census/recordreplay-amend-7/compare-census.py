from pathlib import Path
import json,collections
D=Path('/workspace/scratch/b7phwab-amend7')
def normalize(x):
 if isinstance(x,str):
  for p in ['/workspace/scratch/b7phwab-amend5/control-tree',str(D/'preflight'),str(D/'fresh-tree')]:x=x.replace(p,'$TREE')
  return x
 if isinstance(x,list):return list(map(normalize,x))
 if isinstance(x,dict):return {k:normalize(v) for k,v in x.items()}
 return x
def read(name):
 rows=[json.loads(l) for l in (D/(name+'-census.jsonl')).read_text().splitlines()]
 assert rows[0]['latent_mode']=='full' and rows[0]['module_format']=='CommonJSBundle'
 assert rows[-1].get('status')=='relation_report' and not rows[-1].get('findings')
 files=[r for r in rows if r.get('file','').endswith('/sys.ts')];assert len(files)==1
 return normalize(files[0])
a,b=read('control'),read('candidate')
oldUnits={(u['where'],u['kind']) for u in a['units']}
newUnits={(u['where'],u['kind']) for u in b['units']}
assert not oldUnits-newUnits, 'an original unit disappeared'
addedUnits=[u for u in b['units'] if (u['where'],u['kind']) not in oldUnits]
assert len(addedUnits)==1 and addedUnits[0]['kind']=='KindImportDeclaration' and addedUnits[0]['status']=='attempted'
line=int(addedUnits[0]['where'].rsplit(':',2)[1])
assert (D/'preflight/src/compiler/sys.ts').read_text().splitlines()[line-1]=='import type {} from "../recordreplay.js";'

def key(f):return tuple(f.get(k,'') for k in ['unit','phase','kind','where','reason','pass','piece_atom'])
old={key(f):f for f in a['findings']};new={key(f):f for f in b['findings']}
oldBy=collections.defaultdict(list);newBy=collections.defaultdict(list)
for f in a['findings']:oldBy[f['unit']].append(f)
for f in b['findings']:newBy[f['unit']].append(f)
hard=lambda f:f['kind'] in {'Refused','NotYet','error','panic'}
oldStop={u for u,fs in oldBy.items() if any(map(hard,fs))};newStop={u for u,fs in newBy.items() if any(map(hard,fs))}
added=[{'finding':new[k],'classification':'already stopped unit' if new[k]['unit'] in oldStop else 'newly stopped unit','before':oldBy[new[k]['unit']]} for k in sorted(new.keys()-old.keys())]
result={'control_cli':'66fc7356a943372caa1b494d8d05959f91e0fb7b','module_format':'CommonJSBundle','mode':'full','assert_no_output':True,'added_units':addedUnits,'before_units':len(a['units']),'after_units':len(b['units']),'before_findings':len(a['findings']),'after_findings':len(b['findings']),'newly_stopped_units':sorted(newStop-oldStop),'newly_any_finding_units':sorted(newBy.keys()-oldBy.keys()),'new_findings':added,'removed_findings':[old[k] for k in sorted(old.keys()-new.keys())]}
text=(D/'preflight/src/compiler/sys.ts').read_text();assert '(process as any).recordreplay' not in text and 'process.recordreplay' in text
result['recordreplay_cast_absent']=True
(D/'census-comparison.json').write_text(json.dumps(result,indent=2)+'\n')
assert not result['newly_stopped_units'] and not result['newly_any_finding_units']
print(json.dumps({k:v for k,v in result.items() if k not in ['new_findings','removed_findings']}))
