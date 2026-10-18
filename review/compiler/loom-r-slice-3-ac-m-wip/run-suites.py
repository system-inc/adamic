import json,os,subprocess,time,sys,gzip,shutil
from pathlib import Path
out=Path(__file__).parent
os.sched_setaffinity(0,sorted(os.sched_getaffinity(0))[:4]);env=dict(os.environ,GOMAXPROCS='4',GOPROXY='https://proxy.golang.org|direct',ADAMIC_BUILD_OUTPUT_DIR='/tmp/adamic-gate/loom-r-slice-3-ac-m-work/products',GOCACHE='/tmp/adamic-gate/loom-r-slice-3-ac-m-work/go-cache',ADAMIC_GOCACHE_OFF='1');env.pop('GOCACHEPROG',None)
packages=['internal/ir','internal/javascript','internal/lower','internal/flow','internal/native','internal/oracle','stage1/cohere/cssstrings','stage1/cohere/json','stage1/cohere/markdownblocks','stage1/cohere/selector','stage1/cohere/yaml','stage1/typescript/parser','stage1/typescript/scanner','stage3/fixtures']
if len(sys.argv)>1:packages=sys.argv[1:]
for pkg in packages:
 # Main's compiler-free oracle owners must settle before the parent readers run.
 owner=Path(pkg)/'oracles'
 if owner.is_dir() and pkg not in ['stage1/cohere/markdownblocks']:
  pre=['timeout','120','go','test','-p','1','./'+str(owner),'-run','^TestProduct_','-json','-count=1','-timeout=110s','-parallel=4']
  with (out/(pkg.replace('/','-')+'-owners.log')).open('w') as log: result=subprocess.run(pre,env=env,stdout=log,stderr=subprocess.STDOUT,timeout=125)
  print(json.dumps(dict(owners=str(owner),exit=result.returncode)),flush=True)
 attempt=1+sum(__import__('json').loads(l)['package']==pkg for l in (out/'suites.jsonl').read_text().splitlines()) if (out/'suites.jsonl').exists() else 1
 label=pkg.replace('/','-')+(('-attempt-'+str(attempt)) if attempt>1 else '');command=['timeout','720','go','test','-p','1','./'+pkg+'/...','-count=1','-json','-timeout=11m','-parallel=4']
 start=time.monotonic(); raw=Path('/tmp/adamic-gate/loom-r-slice-3-ac-m-work/logs')/(label+'.jsonl')
 with raw.open('w') as log:ran=subprocess.run(command,env=env,stdout=log,stderr=subprocess.STDOUT,timeout=730)
 events=[]
 for s in raw.read_text().splitlines():
  if s.startswith('{'):
   try:events.append(json.loads(s))
   except ValueError:pass
 finished=[dict(package=e.get('Package'),test=e['Test'],seconds=e.get('Elapsed',0),action=e['Action']) for e in events if e.get('Test') and e['Action'] in ['pass','fail','skip']]
 leaves=[r for r in finished if not any(x['package']==r['package'] and x['test'].startswith(r['test']+'/') for x in finished)]
 overdue=[r for r in leaves if r['seconds']>=60]
 (out/(label+'-seconds.json.txt')).write_text(json.dumps(finished,indent=2)+'\n')
 row=dict(package=pkg,attempt=attempt,log=label+'.jsonl.gz',command=command,exit=ran.returncode,seconds=round(time.monotonic()-start,3),overdue=overdue,failed=[r for r in finished if r['action']=='fail'])
 with (out/'suites.jsonl').open('a') as f:f.write(json.dumps(row)+'\n')
 print(json.dumps({**row,'failed_count':len(row['failed']),'failed':row['failed'][:5]}),flush=True)
 with raw.open('rb') as source,gzip.open(out/(label+'.jsonl.gz'),'wb') as target:shutil.copyfileobj(source,target)
 raw.unlink()
 # A failed package still leaves the other requested packages to check. Never rerun an overdue leaf.
