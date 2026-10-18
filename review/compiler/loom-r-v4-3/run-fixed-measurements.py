import json,os,subprocess,time
from pathlib import Path
root=Path('/workspace/adamic');e=root/'review/compiler/loom-r-v4-3';fixed=e/'fixed'
start=time.monotonic()
while True:
 rows=[json.loads(x) for x in (fixed/'focused-results.jsonl').read_text().splitlines()]
 selected=json.loads((fixed/'selected-focused.json').read_text())
 done=len([r for r in rows if 'test' in r])==sum(map(len,selected.values()))
 mutants=[json.loads(x) for x in (fixed/'mutants/source-mutants.jsonl').read_text().splitlines()]
 if done and len(mutants)==19 and (fixed/"readers/done").exists():
  assert all(r['exit']==0 and r['grain'] for r in rows if 'test' in r)
  assert all(r.get('killed') for r in mutants)
  break
 assert time.monotonic()-start<900,'validation wait exceeded'
 time.sleep(5)
for script,limit in [('measure-fixed-javascript.py',840),('run-fixed-comparison.py',2400)]:
 print('Starting',script,flush=True)
 r=subprocess.run(['python3',str(e/script)],cwd=root,timeout=limit)
 assert r.returncode==0,(script,r.returncode)
 print('Finished',script,flush=True)
