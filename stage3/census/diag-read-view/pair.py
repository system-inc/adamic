import json,re,difflib,sys
from pathlib import Path
s=Path('/workspace/scratch/diag-read-view-1');raw=s/'raw';side=sys.argv[1] if len(sys.argv)>1 else 'candidate';report={};files=['program.ts','utilitiesPublic.ts','utilities.ts','types.ts'];maps={}
for file in files:
 a=s/'main-tree/src/compiler'/file;b=s/(side+'-tree')/'src/compiler'/file
 old=a.read_text().splitlines();new=b.read_text().splitlines();line={}
 for tag,i,j,k,l in difflib.SequenceMatcher(a=old,b=new,autojunk=False).get_opcodes():
  if tag=='equal' or (tag=='replace' and j-i==l-k):line.update({v+1:i+v-k+1 for v in range(k,l)})
 maps['src/compiler/'+file]=line
for file in (sys.argv[2:] or files):
 def read(label):
  p=raw/(label+'-'+file+'.jsonl');records=[json.loads(x) for x in p.read_text().splitlines()];assert records[0]['latent_mode']=='full';assert records[-1]['status']=='relation_report'
  # Retain every finding, including dependencies and every phase/kind.
  findings=[f for r in records for f in r.get('findings',[])];unique={}
  for f in findings:
   where=f.get('where','').removeprefix(str(s/(label+'-tree'))+'/');m=re.match(r'(.*):(\d+):(\d+)$',where)
   if label!='main' and m and m[1] in maps:where=f'{m[1]}:{maps[m[1]].get(int(m[2]),"inserted-"+m[2])}:{m[3]}'
   reason=f.get('reason','').replace(str(s/(label+'-tree'))+'/', '')
   def mapped_reference(match):
    name,ln,column=match.groups()
    if label!='main' and name in maps:ln=str(maps[name].get(int(ln),'inserted-'+ln))
    return name+':'+ln+':'+column
   reason=re.sub(r'(src/[A-Za-z0-9_./-]+\.(?:ts|a)):(\d+):(\d+)',mapped_reference,reason)
   key=(f.get('phase',''),f.get('kind',''),where,reason);unique[key]=f
  return unique
 before=read('main');after=read(side);added=set(after)-set(before);removed=set(before)-set(after);changed=[];changedAdded=set();changedRemoved=set()
 for site in sorted({k[:3] for k in added}&{k[:3] for k in removed}):
  aa={k for k in added if k[:3]==site};rr={k for k in removed if k[:3]==site};changed.append({'mappedSite':site,'before':[before[k] for k in sorted(rr)],'after':[after[k] for k in sorted(aa)]});changedAdded|=aa;changedRemoved|=rr
 report[file]={'before':len(before),'after':len(after),'new':[after[k] for k in sorted(added-changedAdded)],'removed':[before[k] for k in sorted(removed-changedRemoved)],'changed':changed}
(raw/('pair-'+side+'.json')).write_text(json.dumps(report,indent=2)+'\n')
print(json.dumps({f:{k:(len(v) if isinstance(v,list) else v) for k,v in row.items()} for f,row in report.items()},indent=2))
