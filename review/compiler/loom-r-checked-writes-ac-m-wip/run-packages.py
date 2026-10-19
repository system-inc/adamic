import pathlib,subprocess,json,time,sys
root=pathlib.Path.cwd();out=root/'review/compiler/loom-r-checked-writes-ac-m-wip';results=[]
commands=[['./internal/native','^TestProduct_WASICompiler$'],['./cmd/adamic','^TestProduct_CommandRuntime']]
for pkg in ['internal/ir','internal/fresh','internal/javascript','internal/flow','internal/regexp','internal/lower','internal/native','cmd/adamic','stage1/cohere/values','stage3/fixtures','internal/oracle']:
 commands.append(['./'+pkg+'/...',None])
for i,(pkg,selection) in enumerate(commands):
 command=['timeout','600','go','test',pkg,'-count=1','-p=1','-parallel=4','-json','-timeout=570s']
 if selection:command+=['-run',selection]
 label=f'package-{i:02d}';start=time.monotonic()
 with (out/(label+'.log')).open('w') as log:r=subprocess.run(command,cwd=root,stdout=log,stderr=subprocess.STDOUT)
 events=[];packages=[]
 for line in (out/(label+'.log')).read_text().splitlines():
  try:e=json.loads(line)
  except:continue
  if e.get('Action')in('pass','fail','skip'):
   if e.get('Test'):events.append(dict(package=e.get('Package'),test=e['Test'],action=e['Action'],seconds=e.get('Elapsed',0)))
   else:packages.append(dict(package=e.get('Package'),action=e['Action'],seconds=e.get('Elapsed',0)))
 names={(e['package'],e['test'])for e in events};leaves=[e for e in events if not any(p==e['package']and n.startswith(e['test']+'/')for p,n in names)]
 entry=dict(command=command,exit=r.returncode,seconds=round(time.monotonic()-start,3),packages=packages,tests=events,over_budget=[e for e in leaves if e['seconds']>=60]);results.append(entry);(out/'package-results.json').write_text(json.dumps(results,indent=2)+'\n')
 print(label,pkg,'exit',r.returncode,'wall',entry['seconds'],'test events',len(events),'max leaf',max((e['seconds']for e in leaves),default=0),flush=True)
 if r.returncode or entry['over_budget']:sys.exit(1)
