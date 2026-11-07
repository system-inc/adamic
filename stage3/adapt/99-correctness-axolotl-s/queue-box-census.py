import json,os,pathlib,signal,subprocess,sys,time
binary,tree,out=map(pathlib.Path,sys.argv[1:]);out.mkdir(parents=True,exist_ok=True)
env=dict(os.environ,LATENT_FULL='1',LATENT_ONLY_FILE=str((tree/'src/compiler/core.ts').resolve()),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str((out/'unit-times.jsonl').resolve()))
for key in list(env):
 if key.startswith('LATENT_') and key not in ['LATENT_FULL','LATENT_ONLY_FILE','LATENT_ASSERT_NO_OUTPUT','LATENT_TIMES']:del env[key]
start=time.time();timed_out=False
with (out/'census.log').open('w') as log:
 p=subprocess.Popen([str(binary.resolve()),str((tree/'src/compiler').resolve()),str((out/'raw.jsonl').resolve())],env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
 try:rc=p.wait(timeout=1800)
 except subprocess.TimeoutExpired:
  timed_out=True;os.killpg(p.pid,signal.SIGTERM)
  try:rc=p.wait(timeout=10)
  except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);rc=p.wait()
 finally:
  try:os.killpg(p.pid,signal.SIGKILL)
  except ProcessLookupError:pass
(out/'result.json').write_text(json.dumps({'exit':rc,'timedOut':timed_out,'seconds':time.time()-start,'processGroupCleaned':True},indent=2)+'\n')
raise SystemExit(rc)
