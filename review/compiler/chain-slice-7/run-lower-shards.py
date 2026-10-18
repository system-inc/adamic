import subprocess,json,pathlib,time,re,os
p=pathlib.Path('review/compiler/chain-slice-7')
def run(args,name,bound=85):
 started=time.monotonic()
 with (p/name).open('w') as out:
  try: code=subprocess.run(args,stdout=out,stderr=subprocess.STDOUT,timeout=bound).returncode
  except subprocess.TimeoutExpired: code=124
 return dict(command=args,exit=code,seconds=round(time.monotonic()-started,3),log=name)
r=run(['go','test','./internal/lower','-list','^Test','-timeout','80s'],'lower-list.log')
assert r['exit']==0,r
names=[x for x in (p/'lower-list.log').read_text().splitlines() if re.fullmatch(r'Test\w+',x)]
results=[]
for i in range(0,len(names),16):
 group=names[i:i+16]
 row=run(['go','test','./internal/lower','-run','^('+'|'.join(group)+')$','-count=1','-json','-timeout','80s'],f'lower-{i//16:02}.jsonl')
 row['tests']=group; results.append(row)
 (p/'lower-results.json').write_text(json.dumps(results,indent=2)+'\n')
 print(row['log'],row['exit'],row['seconds'],flush=True)
print('tests',len(names),'bad shards',sum(r['exit']!=0 for r in results),flush=True)
