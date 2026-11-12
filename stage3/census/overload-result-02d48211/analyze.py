import pathlib,json,gzip,collections,re
root=pathlib.Path('/workspace/or3-report');out=root/'stage3/census/overload-result-02d48211';old=root/'stage3/census/overload-result-10b6c33'
def read(p):
 op=gzip.open if str(p).endswith('.gz') else open
 with op(p,'rt') as f:return [json.loads(s) for s in f if s.strip()]
def relative(s):return 'src/'+s.split('/src/',1)[1] if '/src/' in s else s
b=read('/workspace/or3-build-parent/latent.jsonl');h=read('/workspace/or3-build-head/latent.jsonl');ob=read(old/'parent-latent.jsonl.gz')
assert b[0]==h[0] and b[0]['latent_mode']=='first-error' and not b[0]['checker_rejected']
assert b[-1]['status']==h[-1]['status']=='relation_report'
def view(rows):
 units={u['where']:u for r in rows for u in r.get('units') or []};first={}
 for r in rows:
  for f in r.get('findings') or []:
   if f['phase']=='lowering' and f['unit'] in units:first.setdefault(f['unit'],f)
 return units,first
bu,bf=view(b);hu,hf=view(h);ou,of=view(ob)
def ordinals(units):
 c=collections.Counter();bywhere={};bykey={}
 for w,u in units.items():
  key=(relative(w).rsplit(':',2)[0],u['kind'],u.get('name',''));n=c[key];c[key]+=1;k=(*key,n);bywhere[w]=k;bykey[k]=w
 return bywhere,bykey
ow,ok=ordinals(ou);bw,bk=ordinals(bu)
oldanalysis=json.loads((old/'unit-analysis.json').read_text())
def stop(f):return (f['where'],f['kind'],f['reason']) if f else None
def isresult(f):return bool(f and re.search(r'overload \d+ of .+ result .+ cannot be served by implementation result ',f['reason']))
def cohort(items):
 result=[]
 for r in items:
  original=r['unit'];key=ow[original['where']];w=bk[key];assert original.get('name')==bu[w].get('name')
  result.append({'unit':bu[w],'original_unit':original,'mapping_key':key,'original_overload_stop':r['base'],'base':bf.get(w),'head':hf.get(w),'unchanged':bf.get(w)==hf.get(w),'result_stop_clear_base':not isresult(bf.get(w)),'result_stop_clear_head':not isresult(hf.get(w)),'lowered_base':bu[w]['status']=='attempted' and w not in bf,'lowered_head':hu[w]['status']=='attempted' and w not in hf})
 return result
c43=cohort(oldanalysis['targets']);c67=cohort(oldanalysis['other_overload']);assert len(c43)==43 and len(c67)==67
changes=[{'unit':bu[w],'base':bf.get(w),'head':hf.get(w),'regression_previously_lowered':bu[w]['status']=='attempted' and w not in bf and w in hf} for w in bu if stop(bf.get(w))!=stop(hf.get(w))]
values=[{'unit':hu[w],'finding':f,'base':bf.get(w),'unchanged':bf.get(w)==f} for w,f in hf.items() if f['reason']=='an overloaded function as a value']
summary={'header':b[0],'unit_count':len(bu),'attempted_before':sum(u['status'] in ('attempted','panic') for u in bu.values()),'attempted_after':sum(u['status'] in ('attempted','panic') for u in hu.values()),'lowered_before':sum(u['status']=='attempted' and w not in bf for w,u in bu.items()),'lowered_after':sum(u['status']=='attempted' and w not in hf for w,u in hu.items()),'overload_before':sum(isresult(f) for f in bf.values()),'overload_after':sum(isresult(f) for f in hf.values()),'targets':c43,'other_overload':c67,'first_stop_changes':changes,'overloaded_value_units':values}
(out/'unit-analysis.json').write_text(json.dumps(summary,indent=2)+'\n')
(out/'classification-request.json').write_text(json.dumps({'entry':b[0]['roots'][0],'sources':b[0]['sources'],'rows':values},indent=2)+'\n')
def esc(s):return str(s).replace('|','\\|').replace('\n','<br>')
def fmt(f):return 'Lowered' if not f else f"{f['kind']}: {f['reason']} at {relative(f['where'])}"
def table(items,target):
 lines=['| Unit | Unit location | Parent first stop | Head first stop | '+('Original result stop at head' if target else 'Unchanged from parent')+' |','| --- | --- | --- | --- | --- |']
 for r in sorted(items,key=lambda r:(r['unit'].get('name',''),r['unit']['where'])):
  u=r['unit'];flag=('clear' if r['result_stop_clear_head'] else 'present') if target else ('yes' if r['unchanged'] else 'MOVED')
  lines.append('| '+' | '.join(esc(s) for s in [u.get('name',u['kind']),relative(u['where']),fmt(r['base']),fmt(r['head']),flag])+' |')
 return '\n'.join(lines)+'\n'
(out/'43-units.md').write_text(table(c43,True));(out/'67-units.md').write_text(table(c67,False))
print(json.dumps({k:v for k,v in summary.items() if k not in ['header','targets','other_overload','first_stop_changes','overloaded_value_units']},indent=2))
print('changed first stops',len(changes),'value stops',len(values),'43 cohort lower',sum(r['lowered_head'] for r in c43),'67 moved',sum(not r['unchanged'] for r in c67))
