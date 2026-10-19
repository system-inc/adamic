import subprocess, json, time, os, signal
from pathlib import Path
root=Path.cwd(); evidence=root/'review/compiler/chain-slice-2'
def run(name,command,limit=110):
 start=time.monotonic()
 with (evidence/(name+'.log')).open('w') as log:
  p=subprocess.Popen(command,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  try: code=p.wait(timeout=limit)
  except subprocess.TimeoutExpired:
   os.killpg(p.pid,signal.SIGKILL);code=p.wait()
 row={'name':name,'command':command,'exit':code,'seconds':round(time.monotonic()-start,3)}
 with (evidence/'proof-runs.jsonl').open('a') as log:log.write(json.dumps(row)+'\n')
 print(json.dumps(row),flush=True)
 return code
names=[x for x in (evidence/'lower-list.log').read_text().splitlines() if x.startswith('Test')]
assert names
for start in range(0,len(names),25):
 pattern='^('+ '|'.join(names[start:start+25])+')$'
 run('lower-shard-'+str(start//25),['go','test','./internal/lower','-run',pattern,'-count=1','-json','-timeout=90s'])
for side in ['main','slice']:
 for start in range(0,38,8):
  pattern='^TestChainSliceAdmission('+ '|'.join(f'{x:02}' for x in range(start,min(start+8,38)))+')$'
  run('admission-'+side+'-'+str(start//8),['go','test','./internal/lower','-overlay='+str(evidence/('admission-'+side+'-overlay.json')),'-run',pattern,'-count=1','-json','-timeout=90s'])
run('reader-guard',['go','test','./internal/ir','-run','^TestCallTargetReaders$','-count=1','-json','-timeout=90s'])
run('fixture-directory',['go','test','./stage3/fixtures','-run','^(TestFixturesSelfCompare|TestFixtureDirectoriesHaveTopLevelTests)$','-count=1','-json','-timeout=90s'])
