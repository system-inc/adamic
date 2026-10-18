import json,os,subprocess,time
from pathlib import Path
out=Path(__file__).parent
os.sched_setaffinity(0,sorted(os.sched_getaffinity(0))[:4])
env=dict(os.environ,GOMAXPROCS='4')
for index,each in enumerate(json.loads((out/'consumer-plan.json').read_text())):
 command=['timeout','180','go','test',each['package'],'-run',each['pattern'],'-json','-count=1','-timeout=90s','-parallel=4']
 start=time.monotonic();log=out/('consumer-%02d.jsonl'%index)
 with log.open('w') as f:run=subprocess.run(command,env=env,stdout=f,stderr=subprocess.STDOUT,timeout=190)
 events=[json.loads(x) for x in log.read_text().splitlines() if x.startswith('{')]
 leaves=[e for e in events if e.get('Test') and e['Action'] in ['pass','skip','fail']]
 row=dict(**each,command=command,exit=run.returncode,seconds=round(time.monotonic()-start,3),leaves=leaves,log=log.name)
 with (out/'consumers.jsonl').open('a') as f:f.write(json.dumps(row)+'\n')
 print(each['package'],run.returncode,row['seconds'],flush=True)
 if run.returncode or any(e.get('Elapsed',0)>=60 for e in leaves):raise SystemExit(1)
