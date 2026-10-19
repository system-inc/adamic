import concurrent.futures,json,os,subprocess,time
from pathlib import Path
root=Path.cwd(); out=root/'review/compiler/chain-slice-6'
paths=subprocess.check_output(['rg','--files','internal/oracle/testdata','-g','*.a'],text=True).splitlines()
env=dict(os.environ,GOMAXPROCS='1')
def check(path):
 row={'path':path}
 for name,binary in [('main','/tmp/chain-slice-6-main'),('slice','/tmp/chain-slice-6-current')]:
  try:
   p=subprocess.run([binary,'c',path],stdout=subprocess.DEVNULL,stderr=subprocess.PIPE,env=env,timeout=20)
   row[name]={'exit':p.returncode,'diagnostic':p.stderr.decode()[:1000]}
  except subprocess.TimeoutExpired:row[name]={'timeout':True}
 return row
rows=[]
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:
 for row in pool.map(check,paths):
  rows.append(row)
  if len(rows)%50==0:
   (out/'admission-census.json').write_text(json.dumps(rows,indent=2)+'\n');print('observed',len(rows),flush=True)
(out/'admission-census.json').write_text(json.dumps(rows,indent=2)+'\n')
delta=[r['path'] for r in rows if r['main'].get('exit',0)!=0 and r['slice'].get('exit')==0]
timeouts=[r['path'] for r in rows if any('timeout' in r[x] for x in ['main','slice'])]
(out/'admission-delta.json').write_text(json.dumps({'corpus':len(paths),'observed':len(rows),'newly_admitted':delta,'timeouts':timeouts},indent=2)+'\n')
print('delta',len(delta),'timeouts',len(timeouts),flush=True)
assert not timeouts
