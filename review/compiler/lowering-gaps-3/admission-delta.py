from pathlib import Path
import subprocess,json,concurrent.futures,sys,time
out=Path('review/compiler/lowering-gaps-3');mode=sys.argv[1];binary=sys.argv[2]
if mode=='before':
 names=subprocess.check_output(['git','ls-files','-co','--exclude-standard','-z']).decode().split('\0')
 names=sorted(set(n for n in names if n.endswith(('.a','.ts')) and not n.endswith('.d.ts') and not n.startswith(('cohere/','review/')) and (any(part in n for part in ('/testdata/','/fixtures/','/gaps/')) or n.startswith(('dedication/','examples/')))))
 (out/'admission-corpus.json').write_text(json.dumps(names,indent=2)+'\n')
else:names=json.loads((out/'admission-corpus.json').read_text())
def probe(name):
 try:
  p=subprocess.run([binary,'c',name],stdout=subprocess.DEVNULL,stderr=subprocess.PIPE,timeout=20)
  return {'path':name,'exit':p.returncode,'diagnostic':p.stderr.decode(errors='replace')[:2000]}
 except subprocess.TimeoutExpired:return {'path':name,'exit':'timeout'}
rows=[];start=time.monotonic()
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:
 for row in pool.map(probe,names):
  rows.append(row)
  if len(rows)%100==0:print(mode,len(rows),'of',len(names),flush=True)
(out/('admission-'+mode+'.json')).write_text(json.dumps(rows,indent=2)+'\n')
print(mode,'complete',len(rows),'admitted',sum(r['exit']==0 for r in rows),'timeouts',sum(r['exit']=='timeout' for r in rows),'seconds',round(time.monotonic()-start,2),flush=True)
if mode=='after':
 before={r['path']:r for r in json.loads((out/'admission-before.json').read_text())}
 delta=[r['path'] for r in rows if r['exit']==0 and before[r['path']]['exit']!=0]
 regression=[r['path'] for r in rows if r['exit']!=0 and before[r['path']]['exit']==0]
 (out/'admission-new.json').write_text(json.dumps(delta,indent=2)+'\n');(out/'admission-regressions.json').write_text(json.dumps(regression,indent=2)+'\n')
 print('new',len(delta),'regressions',len(regression),flush=True)
