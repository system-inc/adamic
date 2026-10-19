import pathlib,subprocess,json,time,sys
out=pathlib.Path('review/compiler/loom-r-checked-writes-ac-m2');results=[]
commands=[('internal/ir',None),('internal/lower',None),('internal/javascript',None),('internal/native',None),('internal/flow',None),('internal/fresh',None),('internal/regexp',None),('cmd/adamic','^TestProduct_CommandRuntime'),('cmd/adamic',None),('stage1/cohere/values/oracles','^TestProduct_ValuesGoOracle$'),('stage1/cohere/values',None),('internal/oracle',None)]
for i,(pkg,select)in enumerate(commands):
 limit=1200 if pkg=='internal/oracle'else 540
 command=['timeout',str(limit+30),'go','test','./'+pkg,'-count=1','-p=1','-parallel=4','-json','-timeout='+str(limit)+'s']
 if select:command+=['-run',select]
 start=time.monotonic();name=f'package-{i:02d}.log'
 with(out/name).open('w')as log:r=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT)
 events=[];packages=[]
 for line in(out/name).read_text().splitlines():
  try:e=json.loads(line)
  except:continue
  if e.get('Action')in('pass','fail','skip'):
   if e.get('Test'):events.append(e)
   else:packages.append(e)
 names={e['Test']for e in events};leaves=[e for e in events if not any(n.startswith(e['Test']+'/')for n in names)];over=[dict(test=e['Test'],seconds=e.get('Elapsed'))for e in leaves if e.get('Elapsed',0)>=60]
 row=dict(package=pkg,selection=select,command=command,exit=r.returncode,wall_seconds=round(time.monotonic()-start,3),packages=packages,test_events=len(events),max_leaf_seconds=max((e.get('Elapsed',0)for e in leaves),default=0),over_budget=over,log=name);results.append(row);(out/'package-results.json').write_text(json.dumps(results,indent=2)+'\n');print(pkg,select,'exit',r.returncode,'wall',row['wall_seconds'],'maxleaf',row['max_leaf_seconds'],flush=True)
 if r.returncode or over:sys.exit(1)
