import json,os,subprocess,time
from pathlib import Path
e=Path(__file__).resolve().parent;art=e/'artifacts';rows=[]
for round_ in range(10):
 for side in (['base','v4'] if round_%2==0 else ['v4','base']):
  root='/tmp/adamic-v4-2-base' if side=='base' else '/tmp/adamic-json-82';start=time.monotonic()
  r=subprocess.run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(art/side/'program.mjs'),'--cases',str(art/'empty-cases.txt')],cwd=root,env={**os.environ,'GOMAXPROCS':'4'},stdout=subprocess.PIPE,stderr=subprocess.PIPE,timeout=30)
  rows.append(dict(side=side,round=round_,seconds=time.monotonic()-start,exit=r.returncode));assert r.returncode==0 and not r.stdout and not r.stderr
(e/'javascript-startup-precise.json').write_text(json.dumps(rows,indent=2)+'\n')
