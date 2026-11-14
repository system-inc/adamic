import json,re,difflib,sys
from pathlib import Path
s=Path('/workspace/scratch/diag-read-view-1');raw=s/'raw';side=sys.argv[1] if len(sys.argv)>1 else 'candidate';report={}
for file in ['program.ts','utilitiesPublic.ts','utilities.ts','types.ts']:
 a=s/'main-tree/src/compiler'/file;b=s/(side+'-tree')/'src/compiler'/file
 old=a.read_text().splitlines();new=b.read_text().splitlines();line={}
 for tag,i,j,k,l in difflib.SequenceMatcher(a=old,b=new,autojunk=False).get_opcodes():
  if tag=='equal':line.update({v+1:i+v-k+1 for v in range(k,l)})
  elif tag=='replace' and j-i==l-k:line.update({v+1:i+v-k+1 for v in range(k,l)})
 def read(label):
  p=raw/(label+'-'+file+'.jsonl');records=[json.loads(x) for x in p.read_text().splitlines()];assert records[0]['latent_mode']=='full',records[0].get('latent_mode');assert records[-1]['status']=='relation_report'
  findings=[f for r in records for f in r.get('findings',[]) if f.get('where','').startswith(str(s/(label+'-tree')/'src/compiler'/file)+':')]
  unique={}
  for f in findings:
   where=f['where'].removeprefix(str(s/(label+'-tree'))+'/');m=re.match(r'(.*):(\d+):(\d+)$',where)
   if label!='main' and m:where=f'{m[1]}:{line.get(int(m[2]),"inserted-"+m[2])}:{m[3]}'
   key=(f.get('phase'),f['kind'],where,f['reason']);unique[key]=f
  return unique
 before=read('main');after=read(side);added=after.keys()-before.keys();removed=before.keys()-after.keys();changed=[]
 for key in list(added):
  matches=[k for k in removed if k[:3]==key[:3]]
  if matches:changed.append({'before':[before[k] for k in matches],'after':after[key]})
 report[file]={'before':len(before),'after':len(after),'new':[after[k] for k in sorted(added)],'removed':[before[k] for k in sorted(removed)],'changed':changed}
(raw/('pair-'+side+'.json')).write_text(json.dumps(report,indent=2)+'\n')
print(json.dumps({f:{k:(len(v) if isinstance(v,list) else v) for k,v in row.items()} for f,row in report.items()},indent=2))
