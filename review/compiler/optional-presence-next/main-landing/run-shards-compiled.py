import concurrent.futures,json,os,pathlib,signal,subprocess,time
ROOT=pathlib.Path('/workspace/adamic'); OUT=ROOT/'review/compiler/optional-presence-next/main-landing'
plan=json.loads((OUT/'shards-remaining-plan.json').read_text())
def run(job):
 start=time.monotonic(); prefix=job['prefix']; log=OUT/('shard-'+prefix.replace('$','exact').replace('/','-')+'.log')
 args=['go','tool','test2json','-t','-p','github.com/system-inc/adamic/internal/lower','/workspace/scratch/landing-lower.test','-test.run=^'+prefix,'-test.count=1','-test.timeout=80s','-test.v=test2json']
 with log.open('w') as output:
  process=subprocess.Popen(args,cwd=ROOT/'internal/lower',stdout=output,stderr=subprocess.STDOUT,start_new_session=True)
  try: status=process.wait(timeout=85)
  except subprocess.TimeoutExpired:
   os.killpg(process.pid,signal.SIGKILL); process.wait(); status=124
 events=[]
 for line in log.read_text().splitlines():
  try: event=json.loads(line)
  except ValueError: continue
  if event.get('Test') in job['names'] and event.get('Action') in ['pass','fail','skip']: events.append(event)
 return dict(prefix=prefix,names=job['names'],command=args,status=status,seconds=round(time.monotonic()-start,3),events=events)
with (OUT/'shards-remaining-results.jsonl').open('w') as results, concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
 for i,result in enumerate(pool.map(run,plan),1):
  results.write(json.dumps(result)+'\n'); results.flush(); print(i,result['prefix'],result['status'],result['seconds'],flush=True)
