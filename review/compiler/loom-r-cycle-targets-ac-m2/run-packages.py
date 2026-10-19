import pathlib,subprocess,json,time,sys
root=pathlib.Path.cwd();out=root/"review/compiler/loom-r-cycle-targets-ac-m2"
packages=['internal/ir','internal/lower','internal/native','internal/javascript','internal/flow','internal/fresh','internal/oracle'];results=[]
for index,pkg in enumerate(packages, start=1):
 limit=1200 if pkg=='internal/oracle'else 540;command=['timeout',str(limit+30),'go','test','./'+pkg,'-count=1','-p=1','-parallel=4','-json','-timeout='+str(limit)+'s'];logname=f'final-package-{index:02d}.log';start=time.monotonic()
 with(out/logname).open('w')as log:r=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT)
 events=[];packages_done=[]
 for line in(out/logname).read_text().splitlines():
  try:e=json.loads(line)
  except:continue
  if e.get('Action')in('pass','fail','skip'):
   if e.get('Test'):events.append(e)
   else:packages_done.append(e)
 names={e['Test']for e in events};leaves=[e for e in events if not any(n.startswith(e['Test']+'/')for n in names)];over=[dict(test=e['Test'],seconds=e.get('Elapsed'))for e in leaves if e.get('Elapsed',0)>=60]
 entry=dict(package=pkg,command=command,log=logname,exit=r.returncode,wall_seconds=round(time.monotonic()-start,3),packages=packages_done,test_events=len(events),max_leaf_seconds=max((e.get('Elapsed',0)for e in leaves),default=0),over_budget=over);results.append(entry);(out/'final-package-results.json').write_text(json.dumps(results,indent=2)+'\n');print(pkg,'exit',r.returncode,'wall',entry['wall_seconds'],'events',len(events),'maxleaf',entry['max_leaf_seconds'],flush=True)
 if r.returncode or over:sys.exit(1)
