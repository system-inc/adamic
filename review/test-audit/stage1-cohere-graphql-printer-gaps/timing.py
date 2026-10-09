from pathlib import Path
import subprocess,time,json
p=Path('review/test-audit/stage1-cohere-graphql-printer-gaps');data=[]
for row in ['TestPrinterConstructorGap','TestPrinterWhitespacePlantedDisagreement','TestPrinterWhitespaceShardGrowth']:
 for i in range(3):
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/graphql/printer/','-run','^'+row+'$'];s=time.monotonic()
  with (p/f'time-{row}-{i}.log').open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT)
  events=[]
  for l in open(p/f'time-{row}-{i}.log'):
   try:events.append(json.loads(l))
   except:pass
  end=[e for e in events if e.get('Action') in ['pass','fail'] and not e.get('Test')]
  data.append(dict(test=row,run=i,status=r.returncode,wall=time.monotonic()-s,seconds=end[-1].get('Elapsed') if end else None,command=cmd));(p/'timings.json').write_text(json.dumps(data,indent=2))
  if r.returncode:raise SystemExit('red isolated baseline '+row)
