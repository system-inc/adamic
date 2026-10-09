import concurrent.futures, json, os, pathlib, subprocess, time
root=pathlib.Path.cwd(); evidence=root/'review/compiler/views-v5'; outcomes=[]
def run(label, command, env=None):
 started=time.monotonic(); log=evidence/(label+'.log')
 with log.open('w') as output:
  try: code=subprocess.run(command,env=env,stdout=output,stderr=subprocess.STDOUT,timeout=88).returncode
  except subprocess.TimeoutExpired: code=124
 result=dict(label=label,command=command,exit=code,seconds=round(time.monotonic()-started,3),log=str(log.relative_to(root)))
 print(json.dumps(result),flush=True); return result
names=[n for n in pathlib.Path('/tmp/views-v5-lower-list.txt').read_text().splitlines() if n.startswith('Test')]
shards=[names[i:i+8] for i in range(0,len(names),8)]
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
 for result in pool.map(lambda pair:run('lower-shard-%02d'%pair[0],['go','test','./internal/lower','-run','^('+ '|'.join(pair[1])+')$','-count=1','-parallel=2','-timeout=80s']),enumerate(shards)): outcomes.append(result)
names=[n for n in pathlib.Path('/tmp/views-v5-intersection-list.txt').read_text().splitlines() if n.startswith('Test')]
for name in names:
 outcomes.append(run('lane-'+name,['go','test','./internal/oracle','-run','^'+name+'$','-count=1','-parallel=2','-timeout=80s','-v']))
for i,pattern in enumerate(['^TestCheckedView(Interfaces|Objects|Arrays)$','^TestCheckedView(ObjectPrimitiveSource|ArrayUnionMembership|ArrayUnionNarrowing)$','^TestV4Direct(Argument|Result|Compatible|OptionalReceiver|RecursiveRefusal)$']):
 outcomes.append(run('checked-view-selection-%d'%i,['go','test','./internal/oracle','-run',pattern,'-count=1','-parallel=2','-timeout=80s']))
(evidence/'intersection-check-results.json').write_text(json.dumps(outcomes,indent=2)+'\n')
