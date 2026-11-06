from pathlib import Path
import json,os,signal,subprocess,time
repo=Path('/workspace/adamic'); out=repo/'stage3/census/next-stops-37f8314b'
scratch=Path('/workspace/scratch/next-stops-37f8314b'); work=scratch/'artifacts/tsc';work.mkdir(parents=True,exist_ok=True)
tree=Path('/workspace/next-stops-37f8314b-adapted')
commands=[('entry-census', [str(scratch/'census'),str(tree/'src/tsc/tsc.ts'),str(work/'census.jsonl')], 1800),('entry-latent',[str(scratch/'entry-census'),str(tree/'src/tsc/tsc.ts'),str(work/'latent.jsonl')],1800)]
for name,cmd,bound in commands:
 started=time.monotonic()
 with (out/(name+'.log')).open('wb') as log:
  env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1')
  process=subprocess.Popen(cmd,stdout=log,stderr=log,env=env,start_new_session=True)
  timedout=False
  try: result=process.wait(timeout=bound)
  except subprocess.TimeoutExpired:
   timedout=True;os.killpg(process.pid,signal.SIGTERM)
   try:process.wait(timeout=10)
   except subprocess.TimeoutExpired:os.killpg(process.pid,signal.SIGKILL);process.wait()
   result=124
 data=work/('census.jsonl' if name=='entry-census' else 'latent.jsonl')
 if data.exists(): (out/(name+'.jsonl')).write_bytes(data.read_bytes())
 status=dict(command=cmd,exit=result,timeout=timedout,seconds=time.monotonic()-started,records=len(data.read_bytes().splitlines()) if data.exists() else 0)
 (out/(name+'-status.json')).write_text(json.dumps(status,indent=2)+'\n')
 print(json.dumps(status),flush=True)
 if result and not timedout: break
