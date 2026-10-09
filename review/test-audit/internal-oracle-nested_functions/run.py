import subprocess,time,json,os,signal,sys
from pathlib import Path
ROOT=Path(__file__).resolve().parent
label=sys.argv[1]; args=sys.argv[2:]
cmd=['go','test','-json','-count=1','-timeout','90s','./internal/oracle/']+args
t=time.monotonic()
with (ROOT/(label+'.log')).open('w') as f:
 proc=subprocess.Popen(cmd,stdout=f,stderr=subprocess.STDOUT,start_new_session=True)
 try: code=proc.wait(timeout=100)
 except subprocess.TimeoutExpired:
  os.killpg(proc.pid,signal.SIGKILL);code=124;proc.wait()
wall=time.monotonic()-t
failed=[];elapsed=None;outputs={};passes=[]
for line in (ROOT/(label+'.log')).read_text().splitlines():
 try:e=json.loads(line)
 except:continue
 test=e.get('Test','')
 if e.get('Action')=='fail' and test:failed.append(test)
 if e.get('Action')=='pass' and test:passes.append(test)
 if e.get('Action') in ['pass','fail'] and not test:elapsed=e.get('Elapsed')
 if e.get('Output'):outputs.setdefault(test,[]).append(e['Output'])
result=dict(command=cmd,wall_seconds=wall,exit=code,binary_seconds=elapsed,failed=failed,passed=passes,outputs=outputs)
(ROOT/(label+'.json')).write_text(json.dumps(result,indent=2))
print(label,code,round(wall,3),elapsed,failed[:15],flush=True)
