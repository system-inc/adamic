import json,os,subprocess,time
from pathlib import Path
e=Path(__file__).resolve().parent
config={'GOMAXPROCS':'4','parallel':4,'memory_limit_bytes':17179869184,'base':'dbdf977f5b8b0979d414ebe6f20e7aadd312e309','v4':'82a5dffd7654f3eb2cb4652a12b99401210f442a','selected_environment':{k:os.environ.get(k) for k in ['ADAMIC_TEST_SHARD','ADAMIC_JSON_MEMORY_UNITS','ADAMIC_JSON_PRETTIER','ADAMIC_GATE_UNCACHED','ADAMIC_JSON_ARTIFACTS']}}
(e/'configuration.json').write_text(json.dumps(config,indent=2)+'\n')
with (e/'commands.jsonl').open('w') as out:
 for name,root in [('base','/tmp/adamic-v4-2-base'),('v4','/tmp/adamic-json-82')]:
  steps=[('corpus',['go','test','./stage1/cohere/json','-run','^TestProduct_JSONPinnedCorpus$','-count=1','-json','-parallel=4','-timeout=90s'],120),('owners',['go','test','./stage1/cohere/json/oracles','-run','^TestProduct_','-count=1','-json','-parallel=4','-timeout=10m'],660),('package',['go','test','./stage1/cohere/json','-count=1','-json','-parallel=4','-timeout=30m'],1860)]
  for step,command,limit in steps:
   start=time.monotonic();env={**os.environ,'GOMAXPROCS':'4'}
   with (e/(name+'-'+step+'.jsonl')).open('w') as log:
    try:r=subprocess.run(command,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT,timeout=limit);code=r.returncode
    except subprocess.TimeoutExpired:code=124
   row=dict(side=name,step=step,root=root,command=command,exit=code,seconds=round(time.monotonic()-start,3));out.write(json.dumps(row)+'\n');out.flush();print(json.dumps(row),flush=True)
   if step!='package' and code:raise RuntimeError('prerequisite failure: '+name+' '+step)
