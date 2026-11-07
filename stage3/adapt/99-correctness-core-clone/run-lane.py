import os,subprocess,time,signal,json,pathlib
r=pathlib.Path('/workspace/scratch/core-wave');out=r/'lane';calls=r/'calls';calls.mkdir(exist_ok=True)
env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE='/workspace/typescript-source',STAGE3_TASK_PROFILE=str(r/'whole-tasks.jsonl'),NODE_DISABLE_COMPILE_CACHE='1',NODE_OPTIONS='--require=/workspace/adamic-baseline/stage3/oracle/profile-tasks.cjs')
begin=time.monotonic()
with (r/'lane.log').open('w') as log:
 p=subprocess.Popen(['bash','stage3/lane/run.sh',str(out)],cwd='/workspace/adamic-baseline',env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
 (r/'lane-pid.json').write_text(json.dumps({'pid':p.pid,'start':time.time()})+'\n')
 try: code=p.wait(timeout=1700)
 except subprocess.TimeoutExpired:
  code=124
 finally:
  rows=[x.split(None,2) for x in subprocess.check_output(['ps','-eo','pid,ppid,comm'],text=True).splitlines()[1:]];ids={p.pid}
  while True:
   expanded=ids|{int(pid) for pid,parent,_ in rows if int(parent) in ids}
   if expanded==ids:break
   ids=expanded
  for pid in sorted(ids,reverse=True):
   try:os.kill(pid,signal.SIGTERM)
   except ProcessLookupError:pass
  try:p.wait(timeout=5)
  except subprocess.TimeoutExpired:
   for pid in sorted(ids,reverse=True):
    try:os.kill(pid,signal.SIGKILL)
    except ProcessLookupError:pass
   p.wait()
(r/'lane-result.json').write_text(json.dumps({'exit':code,'seconds':time.monotonic()-begin,'complete':code==0,'processesCleaned':True},indent=2)+'\n')
print('lane',code,flush=True)
