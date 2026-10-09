import subprocess,json,time,os
from pathlib import Path
E=Path('review/test-audit/internal-unicodeproperties-alias'); scoped=['TestBinaryAliases','TestGeneralCategoryAliases','TestScriptAliasSet','TestRejectedNames','TestStringPropertyCensus','TestSetBoundaries','TestKnownMembership']; rows=scoped+['TestNodeAgrees','TestNodeStringProperties']; results={}
for item in json.loads((E/'menu.json').read_text()):
 mid=item['id']; env=os.environ.copy();env['ADAMIC_MUTANT']=mid
 log=E/(mid+'.log'); start=time.monotonic()
 with log.open('w') as out:run=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/unicodeproperties/','-run','^('+'|'.join(rows)+')$'],env=env,stdout=out,stderr=subprocess.STDOUT)
 data=[]
 for l in log.read_text().splitlines():
  try:data.append(json.loads(l))
  except:pass
 observed={d['Test']:d['Action'] for d in data if 'Test'in d and '/'not in d['Test'] and d['Action'] in ('pass','fail','skip')}
 cooked='panic: test timed out' in log.read_text() or run.returncode==124
 panic='panic: runtime error' in log.read_text()
 if cooked or panic:
  for row in scoped+['TestNodeStringProperties']:
   isolated=E/(mid+'-'+row+'.log')
   with isolated.open('w') as out:subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/unicodeproperties/','-run','^'+row+'$'],env=env,stdout=out,stderr=subprocess.STDOUT)
   entries=[]
   for l in isolated.read_text().splitlines():
    try:entries.append(json.loads(l))
    except:pass
   actions=[d['Action'] for d in entries if d.get('Test')==row and d['Action'] in ('pass','fail','skip')]
   if actions:observed[row]=actions[-1]
   elif 'panic: runtime error' in isolated.read_text():observed[row]='fail'
 results[mid]=dict(results=observed,seconds=round(time.monotonic()-start,3),exit=run.returncode,cooked=cooked,panic=panic)
 (E/'matrix.json').write_text(json.dumps(results,indent=2));print(mid,results[mid],flush=True)
