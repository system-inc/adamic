import pathlib,json,subprocess,time
scope=json.load(open('review/test-audit/internal-oracle-oct6_mutant/scope.json'));all=[];start=time.monotonic()
for row in scope['rows']:
 samples=[]
 for rep in range(1,4):
  with open('/tmp/u066/time-'+row+'-'+str(rep)+'.log','w') as out:p=subprocess.run(['timeout','120','go','test','-count=1','-timeout','90s','./internal/oracle/','-run','^'+row+'$'],stdout=out,stderr=subprocess.STDOUT)
  samples.append(p.returncode);print(row,rep,p.returncode,flush=True)
  if p.returncode:break
 all.append(dict(test=row,codes=samples))
pathlib.Path('/tmp/u066/timing-codes.json').write_text(json.dumps(all,indent=2));pathlib.Path('/tmp/u066/timing-wall.txt').write_text(str(time.monotonic()-start))
