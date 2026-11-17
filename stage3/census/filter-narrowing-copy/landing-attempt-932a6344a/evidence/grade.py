from pathlib import Path
import json,re,difflib,collections
p=Path('/workspace/scratch/ev14ch9-landing-proof');c=p/'census';maps={}
def positions(file):
 if file not in maps:
  a=(p/'before'/file).read_text().splitlines();b=(p/'after'/file).read_text().splitlines();m={}
  for tag,i,j,k,l in difflib.SequenceMatcher(None,a,b,autojunk=False).get_opcodes():
   if tag=='equal':
    for x,y in zip(range(i,j),range(k,l)):m[y+1]=(x+1,None)
   elif tag=='replace' and j-i==l-k:
    for x,y in zip(range(i,j),range(k,l)):m[y+1]=(x+1,(a[x],b[y]))
  maps[file]=m
 return maps[file]
def canonical(side,text):
 def replace(m):
  file,line,col=m.group(1),int(m.group(2)),int(m.group(3))
  if side=='after':
   item=positions(file).get(line)
   if item is None:return 'TREE/'+file+':NEW:'+str(line)+':'+str(col)
   line,changed=item
   if changed:
    old,new=changed;offset=col-1
    for tag,i,j,k,l in difflib.SequenceMatcher(None,old,new,autojunk=False).get_opcodes():
     if k<=offset<l:
      offset=i+(offset-k if tag=='equal' else min(offset-k,max(j-i-1,0)));break
    col=offset+1
  return 'TREE/'+file+':'+str(line)+':'+str(col)
 return re.sub(re.escape(str(p/side))+'/([^:\n]+):(\d+):(\d+)',replace,text)
def key(side,r):return tuple(canonical(side,str(r.get(k,''))) for k in ('unit','phase','kind','where','reason'))
results=[]
for old in sorted(c.glob('before-*.jsonl')):
 name=old.name[len('before-'):];new=c/('after-'+name);be=c/('before-'+name.replace('.jsonl','.exit'));ae=c/('after-'+name.replace('.jsonl','.exit'))
 if not (new.exists() and be.exists() and ae.exists() and be.read_text().strip()=='0' and ae.read_text().strip()=='0'):continue
 data={side:next((x for x in map(json.loads,(c/(side+'-'+name)).read_text().splitlines()) if 'file' in x),None) for side in ('before','after')}
 assert all(data.values());before=data['before']['findings'];after=data['after']['findings'];ac=collections.Counter(key('before',r) for r in before);bc=collections.Counter(key('after',r) for r in after);unmatched_before=[];unmatched_after=[]
 for r in before:
  k=key('before',r)
  if bc[k]:bc[k]-=1
  else:unmatched_before.append(r)
 for r in after:
  k=key('after',r)
  if ac[k]:ac[k]-=1
  else:unmatched_after.append(r)
 blocked={canonical('before',r['unit']) for r in before};reached=[];new_findings=[];changed=[]
 for r in list(unmatched_after):
  same=[b for b in unmatched_before if key('before',b)[:4]==key('after',r)[:4]]
  if same:
   previous=same[0];changed.append({'before':previous,'after':r});unmatched_before.remove(previous);unmatched_after.remove(r)
 for r in unmatched_after:
  (reached if canonical('after',r['unit']) in blocked else new_findings).append(r)
 results.append({'file':data['before']['file'].split('/before/')[-1],'before':len(before),'after':len(after),'unchanged':len(before)-len(unmatched_before),'gone':unmatched_before,'changed':changed,'reached_in_already_blocked_unit':reached,'new_in_previously_clean_units':new_findings,'comparison_key':['unit','phase','kind','where','reason'],'classification_note':'Reaching a previously blocked unit is not counted as a new finding. Reached stops in the new helper still fail the separate requirement that its body lowers without a stop. Raw sources and spans are retained.'})
 print(results[-1]['file'],'before',len(before),'after',len(after),'gone',len(unmatched_before),'reached',len(reached),'new',len(new_findings))
(p/'graded-censuses.json').write_text(json.dumps(results,indent=2)+'\n')
