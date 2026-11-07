import json,os,subprocess,time,signal
from pathlib import Path
s=Path('/workspace/scratch/levenshtein');repo=Path('/workspace/adamic');env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE='/workspace/typescript-source',NODE_PATH=str(repo/'stage3/api/node_modules'));begin=time.monotonic()
with (s/'apply-after.log').open('w') as log:
 p=subprocess.Popen(['bash','stage3/apply.sh','--build-product',str(s/'after')],cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
 try:code=p.wait(timeout=1750)
 except subprocess.TimeoutExpired:code=124
 finally:
  try:os.killpg(p.pid,signal.SIGKILL)
  except ProcessLookupError:pass
  p.wait()
result={'exit':code,'seconds':time.monotonic()-begin};(s/'after-apply.json').write_text(json.dumps(result,indent=2)+'\n');print(result,flush=True);assert code==0
