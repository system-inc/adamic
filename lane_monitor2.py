import pathlib,json,subprocess,os,time,hashlib,shutil,sys
out=pathlib.Path(__file__).parent;repo=pathlib.Path('/workspace/adamic');label,tree=sys.argv[1:3];root=pathlib.Path('/workspace/scratch/tracing2-lane-'+label);assert not root.exists();env=dict(os.environ,STAGE3_ADAPTED_TREE=tree,NODE_OPTIONS='--max-old-space-size=3072',PYTHONDONTWRITEBYTECODE='1')
if label=='control':env['TRACING_CATCH_EVENTS']=str(out/'control2-lane-catch-events.jsonl')
command=['python3','stage3/lane/measure_shards.py',str(root),'--concurrency','2'];runner=pathlib.Path(tree)/'built/local/run.js';wanted=hashlib.sha256(runner.read_bytes()).hexdigest();memory=lambda:{k:int(v) for k,v in (line.split() for line in pathlib.Path('/sys/fs/cgroup/memory.events').read_text().splitlines())};before=memory();started=time.monotonic();cleaned=set();evidence=[];last=0;finished=False
with (out/(label+'2-lane-measure.log')).open('w') as log:
 child=subprocess.Popen(command,cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT)
 while True:
  ledger=root/'ledger.jsonl'
  if ledger.exists():
   for line in ledger.read_text().splitlines():
    try:row=json.loads(line)
    except json.JSONDecodeError:continue
    name=row['name']
    if name in cleaned:continue
    view=root/name/'adapted-tree';report=root/name/'report.json';assert json.loads(report.read_text())['name']==name
    actual=hashlib.sha256((view/'built/local/run.js').read_bytes()).hexdigest() if (view/'built/local/run.js').exists() else None
    assert row['exit']!=0 or actual==wanted,(label,name,'runner bytes changed')
    evidence.append(dict(name=name,runnerSha256=actual,childExit=row['exit'],reportStatus=row['status'],completedViewRemoved=row['exit']==0));
    if row['exit']==0:shutil.rmtree(view)
    cleaned.add(name)
  if time.monotonic()-last>=60:
   print(label,'completed',len(cleaned),'/398','free GiB',round(shutil.disk_usage('/workspace').free/2**30,2),flush=True);last=time.monotonic()
  code=child.poll()
  if code is not None:
   if finished:break
   finished=True
   continue
  time.sleep(1)
after=memory();(out/(label+'2-lane-monitor.json')).write_text(json.dumps(dict(argv=command,childExit=code,seconds=time.monotonic()-started,tree=tree,runnerSha256=wanted,memoryBefore=before,memoryAfter=after,completedViews=evidence),indent=2)+'\n');print(label,'measure child exit='+str(code),flush=True)
if code:raise SystemExit(code)
ledger=json.loads((root/'ledger.json').read_text());assert len(ledger['units'])==398 and len(cleaned)==398;assert all(r['exit']==0 and r['status']=='pass' and r['seconds']<30 for r in ledger['units']);assert after['oom_kill']==before['oom_kill']
counts={k:sum(row['counts'][k] for row in ledger['units']) for k in ['pass','fail','skip']}
assert counts=={'pass':106366,'fail':1,'skip':0},counts
summary=dict(shards=398,allPass=True,counts=counts,maximumShardSeconds=ledger['maximum'],wallSeconds=ledger['wall_seconds'],memoryEventsUnchanged=before==after,runnerIdenticalInEveryShard=True)
if label=='control':
 events=[json.loads(line) for line in (out/'control2-lane-catch-events.jsonl').read_text().splitlines()];catches=[e for e in events if e['kind']=='catch'];loads=[e for e in events if e['kind']=='load'];assert not catches;assert len({e['pid'] for e in loads if e['bundle']=='run.js'})==398;summary.update(oldCatchExecutions=0,bundleLoads={name:sum(e['bundle']==name for e in loads) for name in ['run.js','_tsc.js','typescript.js']},harnessProcesses=398)
(out/(label+'2-lane-summary.json')).write_text(json.dumps(summary,indent=2)+'\n')
