import json,re,subprocess,time
from pathlib import Path
out=Path(__file__).resolve().parent
results=[]
for c in json.loads((out/'mutants/cases.json').read_text()):
 cmd=['go','test','-overlay',c['overlay'],*c['packages'],'-run',c['pattern'],'-count=1','-timeout','90s','-parallel','4','-p','1','-json'];start=time.monotonic();log=out/('mutant-'+c['topic']+'.log')
 with log.open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,timeout=180)
 text=log.read_text();events=[]
 for line in text.splitlines():
  try:e=json.loads(line)
  except ValueError:continue
  if isinstance(e,dict) and 'Test' in e and e.get('Action') in ['pass','fail']:events.append(e)
 failed=[e for e in events if e['Action']=='fail'];large=[e for e in events if e.get('Elapsed',0)>=60 and not any(x['Test'].startswith(e['Test']+'/') for x in events)]
 killed=r.returncode==1 and bool(failed) and '[build failed]' not in text and 'panic: test timed out' not in text and not re.search(r'\.c:[0-9]+:[0-9]+: error:',text)
 row=dict(topic=c['topic'],command=cmd,exit=r.returncode,seconds=round(time.monotonic()-start,3),killed=killed,catchers=failed);results.append(row);(out/'mutant-results.json').write_text(json.dumps(results,indent=2));print(c['topic'],r.returncode,row['seconds'],killed,flush=True)
 assert killed and not large,(c['topic'],large)
