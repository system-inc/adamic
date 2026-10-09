import json,pathlib,subprocess,time
out=pathlib.Path('review/compiler/chain-slice-5');groups=json.loads((out/'remaining-selected.json').read_text());results=[]
for pkg,names in groups.items():
 for i in range(0,len(names),8):
  group=names[i:i+8];cmd=['go','test',pkg,'-run','^('+'|'.join(group)+')$','-count=1','-json','-timeout','90s'];start=time.monotonic();log=out/'logs'/('remaining-'+pkg.rsplit('/',1)[-1]+'-%02d.jsonl'%(i//8))
  with log.open('w') as f:p=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,timeout=105)
  results.append(dict(package=pkg,tests=group,command=cmd,exit=p.returncode,seconds=time.monotonic()-start));(out/'remaining-results.json').write_text(json.dumps(results,indent=2)+'\n');print(pkg,i//8,'exit',p.returncode,flush=True)
  assert p.returncode==0
