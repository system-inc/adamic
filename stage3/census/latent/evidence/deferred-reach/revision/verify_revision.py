import json,sys,subprocess,os,re
from pathlib import Path
root=Path('/workspace/adamic');w=Path('/workspace/scratch/abrq61p-revision');base=Path('/workspace/scratch/abrq61p');rel=lambda s:'src/'+s.split('/src/',1)[1] if '/src/' in s else s
rows=lambda p:[json.loads(s) for s in p.read_text().splitlines()]
def stops(records):
 result={}
 for r in records:
  for f in r.get('findings',[]):
   if f.get('phase') in ('lowering','deferred'):result.setdefault(rel(f['unit']),f)
 return result
old,new=stops(rows(base/'base.jsonl')),stops(rows(w/'mapping.jsonl'));log=(w/'mapping-census.log').read_text();stages=rows(w/'mapping-stages.jsonl');requests=rows(w/'mapping-requests.jsonl');diagnosis=[]
for q in requests:
 target=rel(q['target']);unit=rel(q['unit'])
 if target not in ['src/compiler/executeCommandLine.ts:981:1','src/compiler/moduleNameResolver.ts:3414:1','src/compiler/utilities.ts:9031:1']:continue
 assert old.get(unit)==new.get(unit),(unit,old.get(unit),new.get(unit))
 assert f'declaration={q["target"]} unit={q["unit"]} call={q["call_site"]} declared=1 inferred=0 concrete=false' in log
 trace=[e for e in stages if e['unit']==q['unit'] and e['declaration']==q['target']]
 diagnosis.append(dict(target=target,unit=unit,call_site=rel(q['call_site']),declared_binders=1,recovered_binders=0,concrete=False,first_stop_matches_baseline=True,first_stop=old.get(unit),trace=trace))
assert len(diagnosis)==3
(w/'mapping-diagnosis.json').write_text(json.dumps(diagnosis,indent=2)+'\n')
sys.path.insert(0,str(root/'stage3/census/latent'));from audit_instantiated_by import compare
identities={}
for mode in ['legacy','full','speculative']:
 env=dict(os.environ,LATENT_ASSERT_NO_OUTPUT='1',LATENT_FULL='0' if mode=='legacy' else '1',LATENT_SPECULATIVE='1' if mode=='speculative' else '0')
 output=w/('baseline-controls-'+mode+'.jsonl')
 with (w/('baseline-controls-'+mode+'.log')).open('w') as logfile:
  child=subprocess.run([str(base/'base-census'),str(w/'controls/source'),str(output)],env=env,stdout=logfile,stderr=subprocess.STDOUT,timeout=45)
 print(mode,'baseline control child exit',child.returncode,flush=True);assert child.returncode==0
 identities[mode]=compare(output,w/'controls'/(mode+'.jsonl'))
(w/'controls-identity.json').write_text(json.dumps(identities,indent=2)+'\n')
print('PASS: all three diagnoses preserve first stops; initializer and failed-body controls preserve old flags in all three modes')
