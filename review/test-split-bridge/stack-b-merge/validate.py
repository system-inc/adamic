import os,json,re,subprocess,time
from pathlib import Path
root=Path(__file__).resolve().parents[3]; logs=Path(os.environ.get('ADAMIC_BRIDGE_MERGE_LOGS','/tmp/bridge-stack-b-validation'));env=os.environ.copy()
roots=sorted(set(re.findall(r'func (Test\w+)\(', '\n'.join(p.read_text() for p in (root/'bridge/tsgo').glob('*_test.go')))))
roots=[t for t in roots if t != "TestMain"]
results=json.loads((logs/"units.json").read_text()) if (logs/"units.json").exists() else []
for mode in ('sample','corpus'):
 for test in roots:
  if any(r["mode"]==mode and r["test"]==test for r in results): continue
  e=env.copy()
  if mode=='corpus': e['ADAMIC_TSGO_CORPUS']='/tmp/test-split-bridge-typescript'
  else:e.pop('ADAMIC_TSGO_CORPUS',None)
  path=logs/f'{mode}-{test}.jsonl'; cmd=['go','test','./bridge/tsgo','-json','-run','^'+test+'$','-count=1','-timeout=5m'];beg=time.monotonic()
  with path.open('w') as out:p=subprocess.run(cmd,cwd=root,env=e,stdout=out,stderr=subprocess.STDOUT)
  events=[]
  for line in path.read_text().splitlines():
   try:events.append(json.loads(line))
   except ValueError:pass
  terminal=[x for x in events if x.get('Test')==test and x['Action'] in ('pass','fail','skip')]
  assert p.returncode==0 and len(terminal)==1,(mode,test,p.returncode,path)
  result=dict(mode=mode,test=test,action=terminal[0]['Action'],seconds=terminal[0].get('Elapsed',0),wall=time.monotonic()-beg)
  assert result['seconds']<30,result
  results.append(result);(logs/'units.json').write_text(json.dumps(results,indent=2)+'\n')
 print(mode,'completed',flush=True)
ledger=(logs/'cache.log').read_text().splitlines();counts={}
for name in ('go_build_./bridge/tsgo/archive_tsgo.a','go_build_./bridge/tsgo/archive_tsgo-asan.a'):
 rows=[s.split() for s in ledger if s.startswith('build '+name+' ')]; miss=sum(r[3]=='miss' for r in rows);hit=sum(r[3]=='hit' for r in rows)
 assert miss==1 and hit>=47,(name,miss,hit)
 counts[name]=dict(misses=miss,hits=hit,keys=sorted(set(r[2] for r in rows)))
(logs/'archive-counts.json').write_text(json.dumps(counts,indent=2)+'\n');print(counts,flush=True)
