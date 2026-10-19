import json,subprocess
from verify import run,out
with (out/'go-list.log').open('w') as f: subprocess.run(['go','list','-json','./...'],stdout=f,stderr=subprocess.STDOUT,timeout=120,check=True)
s=(out/'go-list.log').read_text();dec=json.JSONDecoder();items=[]
while s.strip():
 x,n=dec.raw_decode(s.lstrip()); items.append(x);s=s.lstrip()[n:]
readers=set('github.com/system-inc/adamic/'+p for p in ['internal/lower','internal/native','internal/javascript','internal/oracle'])
while True:
 new=readers|{x['ImportPath'] for x in items if readers.intersection(x.get('Imports',[])+x.get('TestImports',[])+x.get('XTestImports',[]))}
 if new==readers:break
 readers=new
packages=sorted(readers);(out/'reader-packages.json').write_text(json.dumps(packages,indent=2)+'\n')
raise SystemExit(run('readers',['go','test',*packages,'-run','^TestCallTargetReaders$','-count=1','-json','-timeout','90s'],300))
