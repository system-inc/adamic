import pathlib,subprocess,json,time,sys
root=pathlib.Path.cwd();out=root/'review/compiler/loom-r-checked-writes-ac-m-wip';results=[]
commands=[['./cmd/adamic/...',None],['./stage1/cohere/values/...',None],['./stage3/fixtures/...',None],['./internal/oracle/...',None]]
for i,(pkg,selection) in enumerate(commands):
 if pkg=='./internal/oracle/...':
  while not (out/'native-final.exit').exists():time.sleep(1)
 command=['timeout','600','go','test',pkg,'-count=1','-p=1','-parallel=4','-json','-timeout=570s']
 if selection:command+=['-run',selection]
 label=f'remaining-final-{i:02d}';start=time.monotonic()
 with (out/(label+'.log')).open('w') as log:r=subprocess.run(command,cwd=root,stdout=log,stderr=subprocess.STDOUT)
 events=[];packages=[]
 for line in (out/(label+'.log')).read_text().splitlines():
  try:e=json.loads(line)
  except:continue
  if e.get('Action')in('pass','fail','skip'):
   if e.get('Test'):events.append(dict(package=e.get('Package'),test=e['Test'],action=e['Action'],seconds=e.get('Elapsed',0)))
   else:packages.append(dict(package=e.get('Package'),action=e['Action'],seconds=e.get('Elapsed',0)))
 names={(e['package'],e['test'])for e in events};leaves=[e for e in events if not any(p==e['package']and n.startswith(e['test']+'/')for p,n in names)]
 entry=dict(command=command,exit=r.returncode,seconds=round(time.monotonic()-start,3),packages=packages,tests=events,over_budget=[e for e in leaves if e['seconds']>=60]);results.append(entry);(out/'remaining-final-results.json').write_text(json.dumps(results,indent=2)+'\n')
 print(label,pkg,'exit',r.returncode,'wall',entry['seconds'],'test events',len(events),'max leaf',max((e['seconds']for e in leaves),default=0),flush=True)
 if r.returncode or entry['over_budget']:sys.exit(1)
