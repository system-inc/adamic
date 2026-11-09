import json,os,signal,subprocess,time
from pathlib import Path
log=Path('/tmp/record-select-run.log')
with log.open('w') as output:
 p=subprocess.Popen(['/tmp/record-select.test','-test.run','^TestRecordSelect(Census|ObserverFixture)$','-test.v','-test.timeout','100m'],stdout=output,stderr=subprocess.STDOUT,start_new_session=True,cwd='/workspace/record-select-run')
 last=time.monotonic();size=-1;stopped=False
 try:
  while p.poll() is None:
   n=log.stat().st_size
   if n!=size:size=n;last=time.monotonic()
   if time.monotonic()-last>1800:
    stopped=True;os.killpg(p.pid,signal.SIGTERM)
    try:p.wait(timeout=10)
    except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL)
    break
   time.sleep(2)
  rc=p.wait()
 finally:
  if p.poll() is None:os.killpg(p.pid,signal.SIGKILL);p.wait()
Path('/tmp/record-select-run-exit.json').write_text(json.dumps({'child_exit_code':rc,'silent_30_minutes':stopped})+'\n')
print('child_exit='+str(rc),flush=True)
raise SystemExit(rc if rc>=0 else 128-rc)
