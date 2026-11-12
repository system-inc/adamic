from pathlib import Path
import json
from collections import Counter
out=Path(__file__).resolve().parent
records={}
for name in ['before','after','mutant']:
 rows=[json.loads(l)for l in (out/('census-'+name+'.jsonl')).read_text().splitlines()];assert rows[0]['latent_mode']=='full' and not rows[0]['diagnostics'];files=[r for r in rows if 'file' in r];assert len(files)==1 and files[0]['file'].endswith('/src/compiler/path.ts');records[name]=files[0]
def keys(name):
 result=set()
 for f in records[name]['findings']:
  file,line,col=f['where'].rsplit(':',2);line=int(line)
  if name=='after' and line>=1094:line+=1
  result.add((f['kind'],'path.ts:'+str(line)+':'+col,f['reason']))
 return result
b,a,m=keys('before'),keys('after'),keys('mutant');removed=sorted(b-a);added=sorted(a-b)
def raw_keys(name):
 result=Counter()
 for f in records[name]['findings']:
  file,line,col=f['where'].rsplit(':',2);line=int(line)
  if name=='after' and line>=1094:line+=1
  result[(f['kind'],'path.ts:'+str(line)+':'+col,f['reason'],f['phase'])]+=1
 return result
raw_before,raw_after,raw_mutant=map(raw_keys,['before','after','mutant'])
assert not (raw_after-raw_before)
assert sum((raw_before-raw_after).values())==1
assert raw_mutant==raw_before
expected=('Refused','path.ts:1105:28',"a cast the runtime can't check")
assert removed==[expected],removed;assert not added,added;assert m==b
result={'rawFindingEventsRemoved':1,'rawFindingEventsAdded':0,'rawMutantMultisetEqualsBefore':True,'removed':removed,'added':added,'mutantRestoresCompleteBeforeFindingSet':True,'restoredRefusal':expected,'normalization':'after lines >=1094 map to before line +1','units':{n:[u for u in r['units']if u.get('name')=='forEachAncestorDirectory']for n,r in records.items()}}
(out/'census-delta.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({'removed':removed,'added':added,'mutantRestored':True}))
