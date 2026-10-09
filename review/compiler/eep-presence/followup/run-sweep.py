import concurrent.futures,json,os,pathlib,signal,subprocess,time
ROOT=pathlib.Path('/workspace/adamic');OUT=ROOT/'review/compiler/eep-presence/followup'
paths=sorted([p for d in ['internal/oracle/testdata','stage3'] for p in (ROOT/d).rglob('*.a')])
(OUT/'fixtures.json').write_text(json.dumps([str(p.relative_to(ROOT)) for p in paths],indent=2)+'\n')
def check(path):
 row={'path':str(path.relative_to(ROOT))}
 for label,binary in [('897d0e79','/tmp/eep-sweep-897'),('branch','/tmp/eep-sweep-branch')]:
  start=time.monotonic(); process=subprocess.Popen([binary,str(path)],cwd=ROOT,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True,start_new_session=True)
  try:
   stdout,stderr=process.communicate(timeout=60)
   row[label]=json.loads(stdout) if process.returncode==0 else dict(admitted=False,stage='probe',error=stderr,exit=process.returncode)
  except subprocess.TimeoutExpired:
   os.killpg(process.pid,signal.SIGKILL);process.communicate();row[label]=dict(admitted=False,stage='timeout',error='60 second probe limit')
  row[label]['seconds']=round(time.monotonic()-start,3)
 return row
with (OUT/'admission-initial.jsonl').open('w') as results,concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
 for i,result in enumerate(pool.map(check,paths),1):
  results.write(json.dumps(result)+'\n');results.flush()
  if result['897d0e79']['admitted'] and not result['branch']['admitted']:print('REGRESSION',result['path'],result['branch']['error'],flush=True)
  if i%50==0: print(i,'/',len(paths),flush=True)
