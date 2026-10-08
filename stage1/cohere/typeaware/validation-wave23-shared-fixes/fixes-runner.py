import subprocess,time,json,os
from pathlib import Path
root=Path('/workspace/wave-23')
env=os.environ.copy()
env.update(GOMAXPROCS='4',ADAMIC_TYPESCRIPT_SOURCE=str(root/'typescript'),ADAMIC_LINT_BENCH='1',ADAMIC_LINT_PROFILE_DIR=str(root/'fixes-final-profiles'),ADAMIC_LINT_PROFILE_SNAPSHOTS=str(root/'fixes-final-profiles'),ADAMIC_WAVE23_REPOSITORY_MANIFEST=str(root/'landing-inputs/repository.manifest'),ADAMIC_WAVE23_COMPILER_MANIFEST=str(root/'landing-inputs/compiler.manifest'))
(root/'fixes-final-profiles').mkdir(exist_ok=True)
for name,args in [('lint',['go','test','-json','./stage1/cohere/lint','-count=1','-timeout=90m'])]:
 start=time.monotonic(); loads=[]
 with (root/('fixes-'+name+'-complete.jsonl')).open('w') as log:
  proc=subprocess.Popen(args,cwd='/workspace/adamic',env=env,stdout=log,stderr=subprocess.STDOUT)
  while proc.poll() is None:
   loads.append(os.getloadavg());time.sleep(10)
 result={'command':args,'exit':proc.returncode,'wall_seconds':time.monotonic()-start,'nproc':len(os.sched_getaffinity(0)),'load_start':loads[0] if loads else [],'load_end':os.getloadavg(),'load_peak_1m':max((x[0] for x in loads),default=0)}
 (root/('fixes-'+name+'-result.json')).write_text(json.dumps(result,indent=2)+'\n')
