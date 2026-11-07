from pathlib import Path
import json
s=Path('/workspace/scratch/arrayof');reason='a member typed any ([]) related to T at the return site'
def read(label):
 records=[json.loads(l) for l in (s/(label+'-census.jsonl')).read_text().splitlines()];meta=records[0];assert len(meta['diagnostics'])==0;rows=[r for r in records if r.get('file','').endswith('/core.ts')];assert len(rows)==1
 events=[]
 def walk(v):
  if isinstance(v,dict):
   if v.get('kind') in ['Refused','NotYet'] and 'where' in v and 'reason' in v:events.append((v['where'].split('/core.ts:')[-1],v['kind'],v['reason'],v.get('phase','')))
   for x in v.values():walk(x)
  elif isinstance(v,list):
   for x in v:walk(x)
 walk(rows[0]);return set(events),dict(units=len(rows[0]['units']),events=len(events),sites=len(set(events)),checkerDiagnostics=0)
before,binfo=read('before');after,ainfo=read('after');mutant,minfo=read('mutant');target={e for e in before if e[2]==reason};assert target,'baseline site absent';assert not any(e[2]==reason for e in after),'arrayOf site remains';assert not after-before,after-before;assert before-after==target,before-after;assert mutant==before,'mutant did not restore exact baseline sites'
result=dict(before=binfo,after=ainfo,mutant=minfo,removedSites=sorted(target),newSites=[],mutantRestoresSite=True,mutantAllSitesEqualBaseline=True,scope='entire core.ts; all raw Refused and NotYet events, not only the ranked range')
(s/'CENSUS-COMPARISON.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
