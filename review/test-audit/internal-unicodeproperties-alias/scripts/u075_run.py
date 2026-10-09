import subprocess,json,time,os
from pathlib import Path
E=Path('review/test-audit/internal-unicodeproperties-alias'); names=['TestBinaryAliases','TestGeneralCategoryAliases','TestScriptAliasSet','TestRejectedNames','TestStringPropertyCensus','TestSetBoundaries','TestKnownMembership']
results={}
for name in names:
 vals=[]
 for i in range(3):
  p=E/f'timing-{name}-{i+1}.log';t=time.monotonic()
  with p.open('w') as out:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/unicodeproperties/','-run','^'+name+'$'],stdout=out,stderr=subprocess.STDOUT)
  data=[json.loads(l) for l in p.read_text().splitlines() if l.startswith('{')]; event=[d for d in data if d.get('Action')=='pass' and 'Test' not in d];vals.append(event[-1]['Elapsed'] if event else None)
 results[name]=vals;(E/'timings.json').write_text(json.dumps(results,indent=2))
