import subprocess,time,json,os,threading
from pathlib import Path
start=time.monotonic();samples=[];stop=threading.Event()
def sample():
 while not stop.wait(5):samples.append(list(os.getloadavg()))
thread=threading.Thread(target=sample);thread.start()
command=['go','test','-json','-count=1','-timeout','60m','./stage1/cohere/lint']
with open('/tmp/unpark-lint.jsonl','wb') as out,open('/tmp/unpark-lint.stderr','wb') as err:result=subprocess.run(command,stdout=out,stderr=err)
stop.set();thread.join();events=[json.loads(s) for s in Path('/tmp/unpark-lint.jsonl').read_text().splitlines() if s.startswith('{')]
counts={k:sum(e.get('Action')==k and 'Test' in e for e in events) for k in ['pass','fail','skip']}
skips=[e['Test'] for e in events if e.get('Action')=='skip' and 'Test' in e]
summary=dict(command=command,exit=result.returncode,wall_seconds=time.monotonic()-start,counts_including_subtests=counts,skips=skips,nproc=int(subprocess.check_output(['nproc'])),load_samples=samples)
Path('/tmp/unpark-lint-summary.json').write_text(json.dumps(summary,indent=2)+'\n');print(json.dumps({k:v for k,v in summary.items() if k!='load_samples'}),flush=True)
