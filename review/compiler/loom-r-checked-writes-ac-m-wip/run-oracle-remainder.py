import pathlib,subprocess,json,time,re,sys
out=pathlib.Path(__file__).parent;data=json.loads((out/'oracle-remainder.json').read_text());pending=data['remaining'];results=[]
for index in range(0,len(pending),24):
 roots=pending[index:index+24];selector='^('+'|'.join(re.escape(n)for n in roots)+')$';logname=f'oracle-remainder-{index//24:02d}.log'
 command=['timeout','300','go','test','./internal/oracle/...','-run',selector,'-count=1','-p=1','-parallel=4','-json','-timeout=270s'];start=time.monotonic()
 with(out/logname).open('w')as log:done=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT)
 events=[];packages=[]
 for line in(out/logname).read_text().splitlines():
  try:e=json.loads(line)
  except:continue
  if e.get('Action')in('pass','fail','skip'):
   if e.get('Test'):events.append(e)
   else:packages.append(e)
 names={e['Test']for e in events};leaves=[e for e in events if not any(n.startswith(e['Test']+'/')for n in names)];over=[dict(test=e['Test'],seconds=e.get('Elapsed'))for e in leaves if e.get('Elapsed',0)>=60]
 completed={e['Test']for e in events if e['Action']in('pass','skip')and'/'not in e['Test']};entry=dict(command=command,log=logname,roots=roots,exit=done.returncode,seconds=round(time.monotonic()-start,3),over_budget=over,missing=sorted(set(roots)-completed),test_events=len(events),max_leaf_seconds=max((e.get('Elapsed',0)for e in leaves),default=0));results.append(entry);(out/'oracle-remainder-results.json').write_text(json.dumps(results,indent=2)+'\n');print(logname,'exit',done.returncode,'wall',entry['seconds'],'events',len(events),'maxleaf',entry['max_leaf_seconds'],'missing',len(entry['missing']),flush=True)
 if done.returncode or over or entry['missing']:sys.exit(1)
all_completed=set(data['completed'])|{n for r in results for n in r['roots']};assert set(data['registered'])-all_completed==set(data['excluded'])
(out/'oracle-coverage.json').write_text(json.dumps(dict(registered=len(data['registered']),completed=len(all_completed),excluded=data['excluded'],all_other_roots_pass=True),indent=2)+'\n')
