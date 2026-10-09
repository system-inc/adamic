from pathlib import Path
import subprocess,time,json,os
r=Path('review/test-audit/internal-lower-check_pragmas'); rows=json.loads((r/'scope.json').read_text()); regex='^('+'|'.join(rows)+')$'
results={}
def run(mid,name,pattern,cap=120):
 env=os.environ.copy();env['ADAMIC_MUTANT']=mid;env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u028/cache/'+(mid or 'clean-switch')
 start=time.monotonic()
 with (r/(name+'.log')).open('w') as f:
  c=subprocess.run(['timeout',str(cap),'go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run',pattern],stdout=f,stderr=subprocess.STDOUT,env=env)
 return dict(exit=c.returncode,wall_seconds=round(time.monotonic()-start,3),log=name+'.log',run=pattern)
results['clean-switch']=run('','clean-switch',regex); assert results['clean-switch']['exit']==0
for m in json.loads((r/'menu.json').read_text()):
 mid=m['id']; result=run(mid,mid,'.'); content=(r/(mid+'.log')).read_text()
 if 'panic:' in content or 'fatal error:' in content:
  result['panic']=True;result['isolated']={row:run(mid,mid+'-'+row,'^'+row+'$') for row in rows}
 elif result['exit']==124 or 'test timed out' in content:
  result['cooked']=True;result['bounded']=run(mid,mid+'-bounded',regex)
 results[mid]=result
 (r/'run-results.json').write_text(json.dumps(results,indent=2)+'\n')
 print(mid,result['exit'],result['wall_seconds'],flush=True)
