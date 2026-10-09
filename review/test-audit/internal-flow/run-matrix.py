import subprocess,time,json,os
from pathlib import Path
r=Path('/workspace/adamic/review/test-audit/internal-flow')
ids=['M'+str(i) for i in range(1,10)]+['PBuild','PConstruct','PLiveOut','PRanges']
for mid in ids:
 env=os.environ.copy();env['ADAMIC_MUTANT']=mid;env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u018/cache/'+mid
 began=time.monotonic()
 with (r/(mid+'.log')).open('w') as log:
  code=subprocess.call(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/flow/','-run','.'],stdout=log,stderr=subprocess.STDOUT,env=env)
 with (r/'run-times.jsonl').open('a') as f:f.write(json.dumps({'id':mid,'wall_seconds':time.monotonic()-began,'exit':code})+'\n')
 print(mid,code,round(time.monotonic()-began,3),flush=True)
