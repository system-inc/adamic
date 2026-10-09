import subprocess,pathlib,json,time
out=pathlib.Path('review/compiler/chain-slice-5');names=json.loads((out/'oracle-selected.json').read_text());results=[]
for i in range(0,len(names),8):
 group=names[i:i+8];command=['go','test','./internal/oracle','-run','^('+'|'.join(group)+')$','-count=1','-json','-timeout','90s']
 start=time.monotonic()
 with (out/'logs'/('oracle-shard-%02d.jsonl'%(i//8))).open('w') as log:p=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=105)
 results.append({'tests':group,'command':command,'exit':p.returncode,'seconds':round(time.monotonic()-start,3)})
 (out/'oracle-shards.json').write_text(json.dumps(results,indent=2)+'\n');print('Oracle shard',i//8,'exit',p.returncode,flush=True)
 assert p.returncode==0
