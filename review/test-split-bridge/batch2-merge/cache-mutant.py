import os,json,subprocess,time,signal
from pathlib import Path
root=Path(__file__).resolve().parents[3];logs=Path(os.environ.get('ADAMIC_BRIDGE_MERGE_LOGS','/tmp/bridge-b2-validation'));p=root/'bridge/tsgo/products_test.go'
s=p.read_text();needle='directory, err := buildcache.Get(inputs, func(directory string) error {';assert s.count(needle)==1
s=s.replace(needle,'if target.output == "tsgo.a" { inputs.Flags = append(inputs.Flags, fmt.Sprint(os.Getpid())) }\n'+needle)
replacement=logs/'archive-rebuild.go';replacement.write_text(s);overlay=logs/'archive-rebuild-overlay.json';overlay.write_text(json.dumps({'Replace':{str(p):str(replacement)}}))
env={**os.environ,'ADAMIC_BUILD_LOG':str(logs/'cache-mutant.log')}
records=[]
for i in range(2):
 with (logs/f'cache-mutant-{i}.jsonl').open('w') as out:
  cmd=['go','test','./bridge/tsgo','-json','-run','^TestBridgeABI$','-count=1','-overlay='+str(overlay),'-timeout=90s'];begun=time.monotonic();p=subprocess.Popen(cmd,cwd=root,env=env,stdout=out,stderr=subprocess.STDOUT,start_new_session=True)
  try:p.wait(timeout=90)
  except subprocess.TimeoutExpired:
   os.killpg(p.pid,signal.SIGKILL);p.wait();raise SystemExit('P0: killed at 90 seconds: archive cache mutant '+str(i))
  assert time.monotonic()-begun<60,'over budget: split archive cache mutant smaller'
 assert p.returncode==0,(i,p.returncode)
 records.append(dict(test="TestBridgeABI",run=i,seconds=time.monotonic()-begun,cooked=False))
rows=[s.split() for s in (logs/'cache-mutant.log').read_text().splitlines() if s.startswith('build go_build_./bridge/tsgo/archive_tsgo.a ')]
misses=sum(r[3]=='miss' for r in rows);assert misses==2,rows
caught=False
try: assert misses==1,'archive built more than once'
except AssertionError:caught=True
assert caught
(logs/'cache-mutant.json').write_text(json.dumps(dict(mutant='add process ID to ordinary archive key',unit='TestBridgeABI',runs=2,selections=records,cooked=False,hard_limit_seconds=90,unit_exits=[0,0],misses=misses,expected_misses=1,caught_by='one-cache-miss assertion'),indent=2)+'\n')
print('archive rebuild mutant caught: two misses, expected one',flush=True)
