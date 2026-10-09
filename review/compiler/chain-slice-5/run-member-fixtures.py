import json,pathlib,subprocess,time,re
out=pathlib.Path('review/compiler/chain-slice-5');paths=json.loads((out/'member-oracle-paths.json').read_text());results=[]
for i in range(0,len(paths),8):
 group=paths[i:i+8];selector='^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^('+ '|'.join(re.escape(pathlib.Path(p).name) for p in group) +')$'
 cmd=['go','test','./internal/oracle','-run',selector,'-count=1','-json','-timeout','90s'];start=time.monotonic()
 with (out/'logs'/('member-fixtures-%02d.jsonl'%(i//8))).open('w') as log: p=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,timeout=105)
 results.append(dict(paths=group,command=cmd,exit=p.returncode,seconds=time.monotonic()-start));(out/'member-fixtures-results.json').write_text(json.dumps(results,indent=2)+'\n');print('Member fixture shard',i//8,'exit',p.returncode,flush=True)
 assert p.returncode==0
