import pathlib,json,subprocess,os,sys,time
p=pathlib.Path(__file__).resolve().parent
label=sys.argv[1]
groups=json.loads((p/'groups.json').read_text())
start=int(sys.argv[2]) if len(sys.argv)>2 else 0
results=json.loads((p/(label+'-runs.json')).read_text()) if start else []
for i,names in enumerate(groups):
 if i<start: continue
 env=os.environ.copy()
 env['ADAMIC_TYPESCRIPT_SOURCE']='/tmp/u156-typescript'
 if label!='clean': env['ADAMIC_BUILD_CACHE_DIR']='/tmp/parser-defense/cache/'+label
 log=p/(label+'-'+str(i)+'.log')
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/typescript/parser/','-run','^('+'|'.join(names)+')$']
 started=time.monotonic()
 with log.open('w') as f: rc=subprocess.call(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
 results.append({'group':i,'rows':names,'command':cmd,'exit':rc,'wall':time.monotonic()-started})
 (p/(label+'-runs.json')).write_text(json.dumps(results,indent=2)+'\n')
 print(label,i,rc,round(results[-1]['wall'],3),flush=True)
 if rc and label=='clean': break
