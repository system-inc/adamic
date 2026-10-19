import json,os,subprocess,sys,time
from pathlib import Path
out=Path(__file__).resolve().parent
os.sched_setaffinity(0, sorted(os.sched_getaffinity(0))[:4])
packages=['internal/ir','internal/javascript','internal/lower','internal/flow','internal/native','internal/oracle','stage1/cohere/cssstrings','stage1/cohere/json','stage1/cohere/markdownblocks','stage1/cohere/selector','stage1/cohere/yaml','stage1/typescript/parser','stage1/typescript/scanner','stage3/fixtures']
def run(name,command,limit=840):
 start=time.monotonic()
 with (out/(name+'.log')).open('w') as log:r=subprocess.run(['timeout',str(limit),*command],stdout=log,stderr=subprocess.STDOUT,timeout=limit+10)
 events=[]
 for line in (out/(name+'.log')).read_text().splitlines():
  try:e=json.loads(line)
  except ValueError:continue
  if e.get('Test') and e.get('Action') in ['pass','fail','skip']:events.append(dict(package=e.get('Package'),test=e['Test'],action=e['Action'],seconds=e.get('Elapsed',0)))
 row=dict(name=name,command=command,exit=r.returncode,seconds=round(time.monotonic()-start,3),failed=[e for e in events if e['action']=='fail'],over_budget=[e for e in events if e['seconds']>=60])
 (out/(name+'-seconds.json.txt')).write_text(json.dumps(events,indent=2)+'\n')
 with (out/'results.jsonl').open('a') as f:f.write(json.dumps(row)+'\n')
 print(json.dumps(row),flush=True)
 return row
if sys.argv[1]=='all':
 if len(sys.argv)>2:packages=sys.argv[2:]
 for p in packages:
  label=p.replace('/','-')
  owner=Path(p)/'oracles'
  if owner.is_dir():run(label+'-owners',['go','test','./'+str(owner),'-run','^TestProduct_','-count=1','-json','-parallel=1','-timeout=13m','-p','1'])
  # Products settle ahead of ordinary consumers as Loom orders them.
  if True:
   run(label+'-products',['go','test','./'+p,'-run','^TestProduct_','-count=1','-json','-parallel=1','-timeout=13m','-p','1'])
  run(label,['go','test','./'+p+'/...','-skip','^TestProduct_','-count=1','-json','-parallel=4','-timeout=13m','-p','1'])
 run('reader-guard',['go','test','./internal/ir','-run','^TestCallTargetReaders$','-count=1','-timeout=90s','-json'],120)
 run('vet',['go','vet','-p','1',*['./'+p+'/...' for p in packages]],180)
