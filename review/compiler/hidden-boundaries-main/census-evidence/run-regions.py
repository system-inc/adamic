import concurrent.futures,json,os,subprocess,sys,time
from pathlib import Path
root=Path(__file__).resolve().parent
queries=json.loads((root/'queries.json').read_text())
jobs=[(side,q) for q in queries for side in ['main','stack']]
# The first main interval was measured separately with a 90-second limit.
def run(job):
 side,q=job;name=side+'-'+q['label'];status=root/(name+'.status.json');output=root/(name+'.jsonl');log=root/(name+'.log')
 command=['/tmp/hidden-census-'+side,'-project','/tmp/any-returns-adapted/src/compiler','-region-file','/tmp/any-returns-adapted/src/compiler/'+q['file'],'-region-start',str(q['start']),'-region-end',str(q['end'])]
 started=time.monotonic()
 with output.open('w') as out,log.open('w') as err:
  try:code=subprocess.run(command,stdout=out,stderr=err,env=dict(os.environ,GOMAXPROCS='2',LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1'),timeout=85).returncode
  except subprocess.TimeoutExpired:code=124
 record=dict(side=side,query=q,command=command,exit=code,seconds=round(time.monotonic()-started,3))
 status.write_text(json.dumps(record,indent=2)+'\n')
 print(json.dumps(record),flush=True)
 return code
selected=[jobs[int(x)] for x in sys.argv[1:]]
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
 codes=list(pool.map(run,selected))
sys.exit(0 if all(x==0 for x in codes) else 1)
