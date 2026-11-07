import hashlib,json,os,signal,subprocess,time,sys
from pathlib import Path
s=Path('/workspace/scratch/levenshtein');repo=Path('/workspace/scratch/levenshtein/main-control');profile=s/'whole-tasks.jsonl';env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE='/workspace/typescript-source',STAGE3_TASK_PROFILE=str(profile),NODE_DISABLE_COMPILE_CACHE='1',NODE_OPTIONS='--require='+str(repo/'stage3/oracle/profile-tasks.cjs'))
start=time.monotonic()
with (s/'whole.log').open('w') as log:
 p=subprocess.Popen(['bash',str(repo/'stage3/lane/run.sh'),str(s/'whole')],cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
 try:code=p.wait(timeout=1750)
 except subprocess.TimeoutExpired:code=124
 finally:
  # Also kill descendants created in child process groups.
  rows=[r.split(None,2) for r in subprocess.check_output(['ps','-eo','pid,ppid,comm'],text=True).splitlines()[1:]];ids={p.pid}
  while True:
   expanded=ids|{int(pid) for pid,parent,_ in rows if int(parent) in ids}
   if expanded==ids:break
   ids=expanded
  for pid in sorted(ids,reverse=True):
   try:os.kill(pid,signal.SIGKILL)
   except ProcessLookupError:pass
  p.wait()
result={'exit':code,'seconds':time.monotonic()-start,'counted':False};(s/'control.json').write_text(json.dumps(result,indent=2)+'\n');print(result,flush=True);assert code==0,result
