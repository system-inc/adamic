from pathlib import Path
import subprocess,json,concurrent.futures,time
out=Path('review/compiler/lowering-gaps-3')
for mode,binary in [('before','/tmp/adamic-gaps3-main'),('after','/tmp/adamic-gaps3-tip')]:
 rows=json.loads((out/('admission-'+mode+'.json')).read_text());pending=[r for r in rows if r['exit']=='timeout' or '@types/node' in r.get('diagnostic','')]
 def probe(row):
  try:
   r=subprocess.run([binary,'c',row['path']],stdout=subprocess.DEVNULL,stderr=subprocess.PIPE,timeout=90)
   result=dict(path=row['path'],exit=r.returncode,diagnostic=r.stderr.decode(errors='replace')[:2000]);print(mode,row['path'],r.returncode,flush=True);return result
  except subprocess.TimeoutExpired:print(mode,row['path'],'timeout',flush=True);return row
 with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:changed={r['path']:r for r in pool.map(probe,pending)}
 rows=[changed.get(r['path'],r) for r in rows];(out/('admission-'+mode+'.json')).write_text(json.dumps(rows,indent=2)+'\n')
 print(mode,'admitted',sum(r['exit']==0 for r in rows),'timeouts',sum(r['exit']=='timeout' for r in rows),flush=True)
before={r['path']:r for r in json.loads((out/'admission-before.json').read_text())};after=json.loads((out/'admission-after.json').read_text())
delta=[r['path'] for r in after if r['exit']==0 and before[r['path']]['exit']!=0];regress=[r['path'] for r in after if r['exit']!=0 and before[r['path']]['exit']==0]
(out/'admission-new.json').write_text(json.dumps(delta,indent=2)+'\n');(out/'admission-regressions.json').write_text(json.dumps(regress,indent=2)+'\n');print('new',len(delta),'regressions',len(regress),flush=True)
