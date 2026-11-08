import subprocess,os,time,signal,json,pathlib
root=pathlib.Path('/workspace/scratch/b5w3ycg')
env=dict(os.environ,LATENT_FULL='1',LATENT_ONLY_FILE='/workspace/core-sites-adapted/src/compiler/core.ts',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(root/'unit-times.jsonl'))
for key in list(env):
 if key.startswith('LATENT_') and key not in ['LATENT_FULL','LATENT_ONLY_FILE','LATENT_ASSERT_NO_OUTPUT','LATENT_TIMES']: del env[key]
started=time.time(); timed_out=False
with (root/'census.log').open('wb') as log:
 p=subprocess.Popen([str(root/'latent'),'/workspace/core-sites-adapted/src/compiler',str(root/'raw.jsonl')],env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
 (root/'census-pid.json').write_text(json.dumps({'pid':p.pid,'started_epoch':started,'limit_seconds':3600})+'\n')
 try:
  rc=p.wait(timeout=3600)
 except subprocess.TimeoutExpired:
  timed_out=True
  os.killpg(p.pid,signal.SIGTERM)
  try: rc=p.wait(timeout=10)
  except subprocess.TimeoutExpired:
   os.killpg(p.pid,signal.SIGKILL);rc=p.wait()
 finally:
  try: os.killpg(p.pid,signal.SIGKILL)
  except ProcessLookupError: pass
 (root/'census-result.json').write_text(json.dumps({'exit_code':rc,'timed_out':timed_out,'elapsed_seconds':time.time()-started,'started_epoch':started,'ended_epoch':time.time(),'process_group_cleaned':True},indent=2)+'\n')
