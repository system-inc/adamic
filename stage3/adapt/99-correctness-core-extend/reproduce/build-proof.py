import os,time,subprocess,signal,json,hashlib
from pathlib import Path
s=Path('/workspace/scratch/core-extend');phases={}
while not all((s/n/'patch-set.md').exists() for n in ['before','after']):
 print('waiting for both applies',flush=True);time.sleep(30)
for name in ['before','after']:
 tree=s/name
 if not (tree/'node_modules').exists():(tree/'node_modules').symlink_to(s/'before/node_modules',target_is_directory=True)
 begin=time.monotonic()
 with (s/('build-'+name+'.log')).open('w') as log:
  p=subprocess.Popen(['npm','run','build:compiler'],cwd=tree,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  try:code=p.wait(timeout=1400)
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   p.wait()
 phases[name]={'exit':code,'seconds':time.monotonic()-begin};(s/'compiler-builds.json').write_text(json.dumps(phases,indent=2)+'\n');print(name,phases[name],flush=True)
 if code:raise SystemExit(code)
def digest(p):return hashlib.sha256(p.read_bytes()).hexdigest()
files=['typescript.d.ts','tsserverlibrary.d.ts'];files+=sorted(p.name for p in (s/'before/built/local').glob('*.js'))
rows=[]
for name in files:
 a=s/'before/built/local'/name;b=s/'after/built/local'/name
 assert a.exists() and b.exists(),name
 row={'file':name,'before':digest(a),'after':digest(b),'identical':a.read_bytes()==b.read_bytes()};rows.append(row)
(s/'emitted-comparison.json').write_text(json.dumps(rows,indent=2)+'\n');assert all(r['identical'] for r in rows)
