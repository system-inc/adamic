"""Supplement same-input pair with explicitly source-changing native B observations.

No headers, byte hashes, unit eligibility or relation trailers are rewritten.
The amendment preserves every existing declaration's line/column. Compare those
identities directly and report the one new erased import unit separately.
"""
from pathlib import Path
import json,gzip,hashlib
D=Path('/workspace/scratch/recordreplay');O=D/'base/stage3/census/recordreplay'
def rows(p):
 data=gzip.decompress(p.read_bytes()) if p.suffix=='.gz' else p.read_bytes()
 return [json.loads(l) for l in data.decode().splitlines()]
def rel(x):
 for name in ['control','candidate2']:x=x.replace(str(D/name)+'/', '')
 return x
def normal(v):
 if isinstance(v,str):return rel(v)
 if isinstance(v,list):return [normal(x) for x in v]
 if isinstance(v,dict):return {k:normal(x) for k,x in v.items()}
 return v
def summarize(before,after):
 def collect(rs):
  us={};fs={};first={}
  for r in rs[1:]:
   for u in r.get('units',[]):us[rel(u['where'])]=normal(u)
   for f in r.get('findings',[]):
    f=normal(f);key=(f['where'],f['kind'],f['reason'],f.get('unit'),f.get('phase'),f.get('text'));fs.setdefault(key,f)
    if f.get('phase') in ['lowering','deferred','prepass']:first.setdefault(f.get('unit'),f)
  return us,fs,first
 bu,bf,bs=collect(before);hu,hf,hs=collect(after)
 stop=lambda f:(f['where'],f['kind'],f['reason']) if f else None
 changes=[{'unit':bu[u],'base':bs.get(u),'head':hs.get(u)} for u in sorted(bu.keys()&hu.keys()) if stop(bs.get(u))!=stop(hs.get(u))]
 new=[hf[k] for k in hf.keys()-bf.keys()]
 return {'comparison':'native original-main source against amended-head source; source identity is explicitly different','method':'existing declaration file/line/column identities, preserved by this amendment; exact raw finding context; no synthetic same-input stream','gone':[bf[k] for k in bf.keys()-hf.keys()],'new':new,'unit_first_stop_delta':changes,'removed_units':[bu[u] for u in bu.keys()-hu.keys()],'added_units':[hu[u] for u in hu.keys()-bu.keys()],'new_process_findings':[f for f in new if 'process' in (f.get('reason','')+' '+f.get('text','')).lower()]}
full=summarize(rows(O/'gate/A-main-sys-full.jsonl.gz'),rows(O/'gate/candidate2-sys-full.jsonl.gz'))
for side,name in [('before','control'),('after','candidate2')]:full[side+'_sys_sha256']=hashlib.sha256((D/name/'src/compiler/sys.ts').read_bytes()).hexdigest()
(O/'full-sys/B-native-source.json').write_text(json.dumps(full,indent=2)+'\n')
lines=['# B native source-changing full sys.ts comparison','',full['method'],'']
for field in ['gone','new','unit_first_stop_delta','added_units','removed_units','new_process_findings']:lines+=['## '+field,'','```json',json.dumps(full[field],indent=2),'```','']
(O/'full-sys/B-native-every-change.md').write_text('\n'.join(lines))
pieces=[]
for g in range(6):
 a=O/'first-error'/f'A-group-{g}'/'base.jsonl.gz';b=O/'first-error'/f'B-group-{g}'/'head.jsonl.gz'
 if not a.exists() or not b.exists():break
 x=summarize(rows(a),rows(b));x['group']=g;pieces.append(x)
if len(pieces)==6:
 result={'complete':True,'method':full['method'],'pieces':pieces,'moved_units':[{'group':p['group'],**u} for p in pieces for u in p['unit_first_stop_delta']],'new_process_findings':[{'group':p['group'],**u} for p in pieces for u in p['new_process_findings']],'added_units':[u for p in pieces for u in p['added_units']],'removed_units':[u for p in pieces for u in p['removed_units']]}
 assert not result['removed_units'];assert len(result['added_units'])==1,result['added_units'];added=result['added_units'][0];assert added['where']=='src/compiler/sys.ts:56:1' and added['kind']=='KindImportDeclaration',added
 (O/'first-error/B-native-source.json').write_text(json.dumps(result,indent=2)+'\n')
 lines=['# B native source-changing all-unit first-error comparison','',result['method'],'']
 for field in ['moved_units','new_process_findings','added_units','removed_units']:lines+=['## '+field,'','```json',json.dumps(result[field],indent=2),'```','']
 (O/'first-error/B-native-every-change.md').write_text('\n'.join(lines))
print('native full source comparison complete;',len(pieces),'of6 all-unit groups available')
