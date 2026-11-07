import os,time,json,subprocess,signal,shutil
from pathlib import Path
s=Path('/workspace/scratch/core-extend')
while not (s/'whole-control-result.json').exists():
 print('waiting for whole-lane control before final observer count',flush=True);time.sleep(20)
assert json.loads((s/'whole-control-result.json').read_text())['exit']==0
pid=json.loads((s/'count-pause.json').read_text())['pid']
rows=[r.strip().split(None,2) for r in subprocess.check_output(['ps','-eo','pid,ppid,args'],text=True).splitlines()[1:]];ids={pid}
while True:
 expanded=ids|{int(p) for p,parent,_ in rows if int(parent) in ids}
 if expanded==ids:break
 ids=expanded
for task in sorted(ids,reverse=True):
 try:os.kill(task,signal.SIGKILL)
 except ProcessLookupError:pass
excluded=s/'excluded-prototype-prefix';excluded.mkdir()
for p in [s/'count-summary.json',s/'count-results',s/'count-pieces.log',*s.glob('piece-*')]:
 if p.exists():shutil.move(str(p),str(excluded/p.name))
(excluded/'REASON.txt').write_text('Excluded prefix: strengthen plain-object classification to distinguish Object.prototype from a custom null-prototype ancestor. Observer edge seeds passed. No production JavaScript changed. The accepted count reruns all 398 units from the beginning with the frozen final observer.\n')
with (s/'count-pieces.log').open('w') as log:
 result=subprocess.run(['python3',str(s/'count-pieces.py')],cwd='/workspace/adamic',stdout=log,stderr=subprocess.STDOUT)
print('final count exit',result.returncode,flush=True)
raise SystemExit(result.returncode)
