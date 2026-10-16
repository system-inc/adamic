import json,subprocess,time
from pathlib import Path
root=Path.cwd();out=root/'review/compiler/chain-slice-4'
files=[p for p in subprocess.check_output(['git','ls-tree','-r','--name-only','5e33a17b']).decode().splitlines() if p.endswith(('.a','.ts')) and p.startswith(('internal/oracle/testdata/','internal/lower/testdata/','stage3/fixtures/','internal/load/testdata/'))]
added=[p for p in subprocess.check_output(['git','diff','--name-only','--diff-filter=A','5e33a17b']).decode().splitlines() if p.endswith('.a') and p.startswith(('internal/','stage3/fixtures/','stage3/project-'))]
files=sorted(set(files+added));rows=[];runs=[]
for i in range(0,len(files),60):
 batch=files[i:i+60];pair=[]
 for version in ['main','slice']:
  saved=out/f'admission-{version}-{i//60}.jsonl'
  if saved.exists():
   cached=[json.loads(line) for line in saved.read_text().splitlines()]
   if [r['path'] for r in cached]==batch:
    pair.append(cached);runs.append({'version':version,'shard':i//60,'files':len(batch),'resumed':True});continue
  start=time.monotonic();result=subprocess.run(['/tmp/slice4/admission-'+version,*batch],capture_output=True,text=True,timeout=85)
  (out/f'admission-{version}-{i//60}.jsonl').write_text(result.stdout)
  if result.returncode:raise SystemExit(result.stderr)
  pair.append([json.loads(line) for line in result.stdout.splitlines()]);runs.append({'version':version,'shard':i//60,'seconds':round(time.monotonic()-start,3),'files':len(batch)})
 for before,after in zip(*pair):
  if before!=after: rows.append({'path':before['path'],'before':before,'after':after,'new_fixture':before['path'] in added})
 print(i//60,len(rows),flush=True)
(out/'admission-delta.json').write_text(json.dumps({'main':'5e33a17b','scope':'tracked main oracle/lower/load testdata and stage3 fixtures plus new slice witnesses; Load and Lower only','files':len(files),'deltas':rows,'shards':runs},indent=2)+'\n')
