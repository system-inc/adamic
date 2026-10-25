import pathlib,subprocess,time,json
r=pathlib.Path('/workspace/adamic');o=r/'review/compiler/loom-r-slice-5-ac2-m2';(o/'logs').mkdir(exist_ok=True)
for n in ['check-headers.py','gate-source.py']:(o/n).write_bytes((r/'review/compiler/loom-r-slice-5-ac2-m'/n).read_bytes())
paths=subprocess.check_output(['git','diff','HEAD','--name-only','--diff-filter=ACMR','--','*.a'],cwd=r,text=True);(o/'paths.txt').write_text(paths)
jobs=[('counts',['go','test','./internal/oracle','-run','^TestCountsAreRecorded$','-count=1','-parallel=4','-timeout','600s','-json','-args','-update-counts']),('headers',['python3',str(o/'check-headers.py'),'final']),('prepare-cli',['go','test','./cmd/adamic','-run','^TestProduct_CommandRuntime$','-count=1','-timeout','90s','-json'])]
for p in ['cmd/adamic','internal/flow','internal/fresh','internal/ir','internal/javascript','internal/lower','internal/native','internal/oracle','stage3/fixtures']:
 jobs.append((p.replace('/','-'),['go','test','./'+p+'/...','-count=1','-parallel=4','-timeout','1200s','-json']))
results=[]
for name,cmd in jobs:
 start=time.monotonic()
 with (o/'logs'/f'{name}.jsonl').open('w') as f:
  try:code=subprocess.run(cmd,cwd=r,stdout=f,stderr=subprocess.STDOUT,timeout=1250).returncode
  except subprocess.TimeoutExpired:code=124
 results.append(dict(name=name,cmd=cmd,exit=code,seconds=time.monotonic()-start));(o/'commands.json').write_text(json.dumps(results,indent=2)+'\n');print(name,code,round(results[-1]['seconds'],2),flush=True)
