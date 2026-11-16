from pathlib import Path
import json,re,collections
D=Path('/workspace/scratch/b7phwab-amend5')
def read(name):
 rows=[json.loads(l) for l in (D/(name+'-census.jsonl')).read_text().splitlines()]
 assert rows[-1].get('status')=='relation_report', 'census did not complete'
 assert rows[0]['module_format']=='CommonJSBundle'
 files=[r for r in rows if r.get('file','').endswith('/sys.ts')]
 assert len(files)==1, len(files)
 return files[0]
a,b=read('control'),read('candidate')
def normalize(v):
 if isinstance(v,str):return v.replace(str(D/'control-tree'),'$TREE').replace(str(D/'candidate-tree'),'$TREE').replace(str(D/'final-composed'),'$TREE')
 if isinstance(v,list):return [normalize(x) for x in v]
 if isinstance(v,dict):return {k:normalize(x) for k,x in v.items()}
 return v
a,b=normalize(a),normalize(b)
assert {(u['where'],u['kind']) for u in a['units']}=={(u['where'],u['kind']) for u in b['units']}, 'unit set changed'
def key(f):return tuple(f.get(k,'') for k in ['unit','phase','kind','where','reason','pass','piece_atom'])
old={key(f):f for f in a['findings']};new={key(f):f for f in b['findings']}
old_by=collections.defaultdict(list);new_by=collections.defaultdict(list)
for f in a['findings']:old_by[f['unit']].append(f)
for f in b['findings']:new_by[f['unit']].append(f)
hard=lambda f:f['kind'] in {'Refused','NotYet','error','panic'}
old_hard={u for u,fs in old_by.items() if any(map(hard,fs))};new_hard={u for u,fs in new_by.items() if any(map(hard,fs))}
newly=sorted(new_hard-old_hard)
classified=[]
for k in sorted(new.keys()-old.keys()):
 f=new[k];u=f['unit']
 category='already stopped unit; newly exposed finding' if u in old_hard else 'previously dependency-skipped unit' if u in old_by else 'newly stopped unit'
 classified.append(dict(classification=category,finding=f,before=old_by[u]))
result=dict(module_format='CommonJSBundle',mode='full',assert_no_output=True,before_units=len(a['units']),after_units=len(b['units']),before_findings=len(a['findings']),after_findings=len(b['findings']),newly_stopped_units=newly,newly_any_finding_units=sorted(new_by.keys()-old_by.keys()),removed_findings=[old[k] for k in sorted(old.keys()-new.keys())],new_findings=classified)
sys=(D/'final-composed/src/compiler/sys.ts').read_text()
assert '(process as any).recordreplay' not in sys and 'process.recordreplay' in sys
result['recordreplay_cast_absent']=True
(D/'census-comparison.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps({k:v for k,v in result.items() if k not in ['removed_findings','new_findings']}))
assert not newly,newly
assert not result['newly_any_finding_units'],result['newly_any_finding_units']
