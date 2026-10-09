import json,os,subprocess,time,statistics
from pathlib import Path
p=Path('/tmp/u098'); rows=json.loads((p/'scope.json').read_text())['rows'];env=os.environ.copy();env.update(ADAMIC_GRAPHQL_PRINTER_BENCH='1',ADAMIC_GRAPHQL_PRETTIER='/tmp/u082/library')
results=[]
for ri,r in enumerate(rows):
 samples=[]
 for n in range(3):
  name=f'time-{ri}-{n+1}'; cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/graphql/printer/','-run',r['pattern']]
  t=time.monotonic()
  with (p/(name+'.log')).open('w') as f:run=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
  events=[]
  for l in (p/(name+'.log')).read_text().splitlines():
   try:events.append(json.loads(l))
   except:pass
  sec=next((x['Elapsed'] for x in reversed(events) if x.get('Action')=='pass' and not x.get('Test')),None)
  # The printed ok line is the required timer, rather than JSON package Elapsed.
  import re
  printed=[re.search(r'\t([0-9.]+)s',x.get('Output','')) for x in events]
  sec=next((float(x.group(1)) for x in reversed(printed) if x),sec)
  meta=dict(command=cmd,exit=run.returncode,wall=time.monotonic()-t,seconds=sec)
  (p/(name+'.meta')).write_text(json.dumps(meta,indent=2));assert run.returncode==0,(name,meta)
  samples.append(sec)
 results.append(dict(test=r['test'],samples=samples,median=statistics.median(samples)))
 (p/'timings.json').write_text(json.dumps(results,indent=2))
(p/'timing-done').write_text('done')
