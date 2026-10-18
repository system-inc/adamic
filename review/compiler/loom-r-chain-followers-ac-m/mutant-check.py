import json,os,subprocess,time
from pathlib import Path
out=Path(__file__).resolve().parent
results=json.loads((out/'checks.json').read_text())
def run(label,command,expected=0,timeout=840):
 start=time.monotonic()
 with (out/(label+'.log')).open('w') as log:
  r=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=timeout)
 row=dict(label=label,command=command,exit=r.returncode,seconds=round(time.monotonic()-start,3));results.append(row)
 (out/'checks.json').write_text(json.dumps(results,indent=2));print(label,r.returncode,row['seconds'],flush=True)
 assert r.returncode==expected,(label,r.returncode)
 events=[]
 for line in (out/(label+'.log')).read_text().splitlines():
  try:e=json.loads(line)
  except ValueError:continue
  if isinstance(e,dict) and 'Test' in e and e.get('Action') in ('pass','fail'):events.append(e)
 oversized=[{'test':e['Test'],'seconds':e.get('Elapsed',0)} for e in events if e.get('Elapsed',0)>=60 and not any(x['Test'].startswith(e['Test']+'/') for x in events)]
 if oversized:
  (out/(label+'-over-limit.json')).write_text(json.dumps(oversized,indent=2));raise AssertionError((label,oversized))
standard=['-count=1','-timeout','90s','-parallel','4','-json']
for c in json.loads((out/'mutants/cases.json').read_text()):
 run('mutant-'+c['topic'],['go','test','-overlay',c['overlay'],*c['packages'],'-run',c['pattern'],*standard],expected=1,timeout=180)
 text=(out/('mutant-'+c['topic']+'.log')).read_text()
 assert '[build failed]' not in text and 'use of undeclared identifier' not in text,c['topic']
 events=[]
 for line in text.splitlines():
  try:e=json.loads(line)
  except ValueError:continue
  if e.get('Action')=='fail' and 'Test' in e:events.append(e)
 assert events,(c['topic'],'no test catcher')
