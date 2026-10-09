from pathlib import Path
import subprocess, json, time
out=Path(__file__).parent
results=[]
manifest={}
for package,count in [('native',16),('lower',4)]:
 names=[line for line in (out/(package+'-list.log')).read_text().splitlines() if line.startswith(('Test','Example','Fuzz'))]
 assert names and len(names)==len(set(names))
 shards=[names[i::count] for i in range(count)]
 assert sorted(sum(shards,[]))==sorted(names)
 manifest[package]=shards
 (out/'shards.json').write_text(json.dumps(manifest,indent=2)+'\n')
 for i,leaves in enumerate(shards):
  command=['go','test','./internal/'+package,'-run','^('+'|'.join(leaves)+')$','-count=1','-v','-timeout','80s']
  start=time.monotonic()
  with (out/(package+'-shard-'+str(i+1)+'.log')).open('w') as log:
   result=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=89)
  results.append(dict(package=package,shard=i+1,tests=len(leaves),exit=result.returncode,seconds=round(time.monotonic()-start,3),command=command))
  (out/'shard-results.json').write_text(json.dumps(results,indent=2)+'\n')
  print({k:v for k,v in results[-1].items() if k!='command'},flush=True)
  assert result.returncode==0
