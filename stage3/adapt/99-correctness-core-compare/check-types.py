import json,subprocess,time,os,signal,sys
from pathlib import Path
s=Path('/workspace/scratch/core-compare');side=sys.argv[1];roots=sorted((s/side/'src/compiler').rglob('*.ts'));assert len(roots)==79,len(roots);begin=time.monotonic()
with (s/('types-'+side+'.log')).open('w') as log:
 p=subprocess.Popen([str(s/'adamic'),'types',*[str(p) for p in roots]],stdout=log,stderr=log,start_new_session=True)
 try:code=p.wait(timeout=1800)
 finally:
  try:os.killpg(p.pid,signal.SIGKILL)
  except ProcessLookupError:pass
  p.wait()
result={'exit':code,'seconds':time.monotonic()-begin,'roots':[str(p) for p in roots]};(s/('types-'+side+'.json')).write_text(json.dumps(result,indent=2)+'\n');assert code==0,result
print(json.dumps(result),flush=True)
