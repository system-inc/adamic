import subprocess,time,json,os,signal,sys
from pathlib import Path
p=Path(__file__).resolve().parent;label=sys.argv[1]
cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/']+sys.argv[2:]
t=time.monotonic()
with (p/(label+'.log')).open('w') as f:
 proc=subprocess.Popen(cmd,stdout=f,stderr=subprocess.STDOUT,start_new_session=True)
 try:code=proc.wait(timeout=121)
 except subprocess.TimeoutExpired:os.killpg(proc.pid,signal.SIGKILL);proc.wait();code=124
r=dict(command=cmd,wall=time.monotonic()-t,exit=code,seconds=None,failed=[],passed=[],outputs={})
for line in (p/(label+'.log')).read_text().splitlines():
 try:e=json.loads(line)
 except:continue
 name=e.get('Test','')
 if e.get('Action') in ['pass','fail']:
  if name:r['failed' if e['Action']=='fail' else 'passed'].append(name)
  else:r['seconds']=e.get('Elapsed')
 if e.get('Output'):r['outputs'].setdefault(name,[]).append(e['Output'])
(p/(label+'.json')).write_text(json.dumps(r,indent=2));print(label,code,round(r['wall'],3),r['seconds'],r['failed'][:8],flush=True)
