import subprocess,sys,time,json,os
from pathlib import Path
label=sys.argv[1]; cmd=sys.argv[2:]; out=Path('/tmp/tracing-8777346a')
env=dict(os.environ)
if label=='census':
 env={k:v for k,v in env.items() if not k.startswith('LATENT_')}
 env.update(LATENT_FULL='1',LATENT_ONLY_FILE='/tmp/tracing-8777346a/adapted/src/compiler/tracing.ts',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(out/'unit-times.jsonl'))
t=time.time()
with (out/(label+'.log')).open('w') as f:
 p=subprocess.Popen(cmd,stdout=f,stderr=subprocess.STDOUT,env=env,start_new_session=True)
 try: rc=p.wait(timeout=1740)
 except subprocess.TimeoutExpired:
  import signal
  os.killpg(p.pid,signal.SIGTERM)
  try: rc=p.wait(timeout=5)
  except subprocess.TimeoutExpired: os.killpg(p.pid,signal.SIGKILL); rc=p.wait()
r={'command':cmd,'exit_code':rc,'seconds':time.time()-t}
(out/(label+'-result.json')).write_text(json.dumps(r,indent=2)+'\n');print(json.dumps(r),flush=True)
sys.exit(rc if rc>=0 else 128-rc)
