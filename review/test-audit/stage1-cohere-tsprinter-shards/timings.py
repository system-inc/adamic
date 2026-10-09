import json,os,subprocess,time
from pathlib import Path
p=Path('review/test-audit/stage1-cohere-tsprinter-shards')
for name,regex,kind in json.loads((p/'rows-plan.json').read_text()):
 for i in [1,2,3]:
  stem='timing-'+name.replace(' ','_')+'-'+str(i)
  cmd=['timeout','120','go','test','-count=1','-timeout','90s','./stage1/cohere/tsprinter/','-run',regex]
  start=time.monotonic()
  with (p/(stem+'.log')).open('w') as out:result=subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT)
  (p/(stem+'.json')).write_text(json.dumps({'command':cmd,'seconds_wall':time.monotonic()-start,'exit':result.returncode})+'\n')
  if result.returncode:break
