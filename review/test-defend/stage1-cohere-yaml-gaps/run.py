import pathlib,subprocess,json,time,os,sys
p=pathlib.Path('/tmp/defend-yaml-gaps');mid=sys.argv[1]
batches=json.loads((p/'batches.json').read_text());results=[]
env=os.environ.copy();env['ADAMIC_YAML_LIBRARY']=str(p/'library')
if mid!='baseline':env['ADAMIC_BUILD_CACHE_DIR']=str(p/'cache'/mid)
for i,rows in enumerate(batches):
 if len(sys.argv)>2 and i not in [int(v) for v in sys.argv[2].split(',')]:continue
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/yaml/','-run','^('+'|'.join(rows)+')$']
 start=time.monotonic()
 with (p/f'{mid}-batch{i}.log').open('w') as f: r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
 result={'id':mid,'batch':i,'rows':rows,'command':cmd,'exit':r.returncode,'wall_seconds':time.monotonic()-start};results.append(result)
 (p/f'{mid}-runs.json').write_text(json.dumps(results,indent=2)+'\n')
 print(mid,i,r.returncode,round(result['wall_seconds'],3),flush=True)
 if r.returncode not in (0,1):break
 if mid=='baseline' and r.returncode:break
