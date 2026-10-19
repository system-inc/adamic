import json,os,subprocess,time
from pathlib import Path
root=Path.cwd();e=root/'review/compiler/loom-r-v4-2-ac-m';env={**os.environ,'GOMAXPROCS':'4'}
for key in ['ADAMIC_TEST_SHARD','ADAMIC_STAGE3_FIXTURE','ADAMIC_JSON_NODE_ONLY','ADAMIC_JSON_BENCH','ADAMIC_JSON_ARTIFACTS']:env.pop(key,None)
packages=[r['package'] for r in json.loads((e/'package-results.json').read_text()) if r['exit'] != 0];results=[]
for package in packages:
 command=['go','test','./'+package,'-count=1','-json','-parallel=4','-timeout=30m'];name=package.replace('/','-');start=time.monotonic();print('Starting',package,flush=True)
 with (e/(name+'-retry.jsonl')).open('w') as log:
  try:r=subprocess.run(command,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT,timeout=1860);code=r.returncode
  except subprocess.TimeoutExpired:code=124
 row=dict(package=package,command=command,exit=code,seconds=round(time.monotonic()-start,3));results.append(row);(e/'package-retry-results.json').write_text(json.dumps(results,indent=2)+'\n');print(json.dumps(row),flush=True)
print('Completed all touched packages',flush=True)
