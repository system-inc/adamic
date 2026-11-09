import os,sys,subprocess,json,time,signal
from pathlib import Path
base=Path('/workspace/scratch/widened');out=base/'report/stage3/census/find-ancestor-widened'
for side in ['parent','head']:
 tree=base/side;log=out/(side+'-vet.log');start=time.monotonic();last=start;stamp=None;stop=None
 env=dict(os.environ,GOWORK=str(base/(side+'.work')))
 with log.open('wb') as f:
  p=subprocess.Popen(['go','vet','./internal/lower'],cwd=tree,env=env,stdout=f,stderr=subprocess.STDOUT,start_new_session=True)
  try:
   while p.poll() is None:
    st=log.stat();current=(st.st_size,st.st_mtime_ns)
    if current!=stamp:last=time.monotonic();stamp=current
    if time.monotonic()-last>1800:
     stop='silent for 1800 seconds';os.killpg(p.pid,signal.SIGTERM)
     try:p.wait(timeout=10)
     except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL)
     break
    time.sleep(1)
   rc=p.wait()
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   p.wait()
 (out/(side+'-vet.status.json')).write_text(json.dumps(dict(command=['go','vet','./internal/lower'],cwd=str(tree),sha=subprocess.check_output(['git','rev-parse','HEAD'],cwd=tree,text=True).strip(),GOWORK=env['GOWORK'],exit=rc,wall_seconds=time.monotonic()-start,stopped=stop),indent=2)+'\n')
 print(side+' vet child_exit_code='+str(rc),flush=True)
 if rc:sys.exit(rc)
