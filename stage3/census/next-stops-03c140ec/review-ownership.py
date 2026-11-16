"""Record owner/target pin searches without duplicating whole pinned functions."""
from pathlib import Path
import json,subprocess
b=Path(__file__).resolve().parent;results=[];terms=['getNodeSystem','getFirstJSDocTag','getAllJSDocTags']
for p in sorted(Path('stage3/adapt').rglob('owners.json'))+sorted(Path('stage3/adapt').rglob('targets.json')):
 v=json.loads(p.read_text())
 def walk(x,path=''):
  if isinstance(x,dict):
   for k,a in x.items():
    if isinstance(a,str) and any(q in a for q in terms):
     results.append(dict(file=str(p),json_path=path+'/'+k,match_terms=[q for q in terms if q in a],first_line=a.splitlines()[0][:220],metadata={kk:vv for kk,vv in x.items() if isinstance(vv,(str,int,bool)) and len(str(vv))<180}))
    elif isinstance(a,(list,dict)):walk(a,path+'/'+k)
  elif isinstance(x,list):
   for i,a in enumerate(x):walk(a,path+'/'+str(i))
 walk(v)
(b/'ownership-search.json').write_text(json.dumps(results,indent=2)+'\n')
p=subprocess.run(['rg','-n','|'.join(terms),'stage3/adapt','--glob','owners.json','--glob','targets.json'],stdout=subprocess.PIPE,text=True)
assert p.returncode in (0,1)
(b/'ownership-grep.txt').write_text('\n'.join(line[:350] for line in p.stdout.splitlines())+'\n')
for r in results:print(r['file'],r['json_path'],r['first_line'])
