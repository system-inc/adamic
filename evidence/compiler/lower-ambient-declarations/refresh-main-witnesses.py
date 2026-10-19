import json,subprocess
from pathlib import Path
e=Path('/tmp/adamic-ambient-evidence')
paths=['internal/oracle/testdata/ambient_refused/'+name+'.a' for name in ('global','module','reference_lib','reference_path','reference_types')]
updates={}
for path in paths:
 result=subprocess.run([str(e/'adamic-main'),'c',path],cwd='/workspace/adamic',stdout=subprocess.DEVNULL,stderr=subprocess.PIPE,timeout=120)
 updates[path]={'path':path,'main':{'exit':result.returncode,'stderr':result.stderr.decode()}}
(e/'sweep-main-refreshed-witnesses.jsonl').write_text(''.join(json.dumps(row)+'\n' for row in updates.values()))
rows=[json.loads(s) for s in (e/'sweep-main.jsonl').read_text().splitlines()]
(e/'sweep-main.jsonl').write_text(''.join(json.dumps(updates.get(r['path'],r))+'\n' for r in rows))
print('refreshed main observations for five strengthened witnesses')
