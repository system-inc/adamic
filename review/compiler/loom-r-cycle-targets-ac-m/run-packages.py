import pathlib,subprocess,json,time,difflib,sys
root=pathlib.Path.cwd();out=root/'review/compiler/loom-r-cycle-targets-ac-m'
deadline=time.monotonic()+180
while not(out/'admission-mutant.exit.log').exists():
 if time.monotonic()>deadline:raise SystemExit('admission mutant did not finish')
 time.sleep(1)
assert(out/'admission-mutant.exit.log').read_text().strip()=='0'
path=root/'internal/ir/call_targets.go';original=path.read_text();start=original.index('\tcase Effects:\n');end=original.index('\tcase Defined:',start);mutated=original[:start]+original[end:]
(out/'drop-effects-result.patch').write_text(''.join(difflib.unified_diff(original.splitlines(True),mutated.splitlines(True),fromfile='internal/ir/call_targets.go',tofile='internal/ir/call_targets.go')))
try:
 path.write_text(mutated)
 with(out/'drop-effects-result.log').open('w')as log:r=subprocess.run(['timeout','120','go','test','./internal/ir','-run','^TestClosureTargetsPreserveEffectsProducers$','-count=1','-parallel=4','-v','-timeout=90s'],stdout=log,stderr=subprocess.STDOUT)
 text=(out/'drop-effects-result.log').read_text();assert r.returncode==1 and 'effects lost adapter identity'in text and 'build failed'not in text
finally:path.write_text(original)
print('drop-effects-result: caught by TestClosureTargetsPreserveEffectsProducers',flush=True)
packages=['internal/ir','internal/lower','internal/native','internal/javascript','internal/flow','internal/fresh','internal/oracle'];results=[]
for index,pkg in enumerate(packages):
 limit=1200 if pkg=='internal/oracle'else 540;command=['timeout',str(limit+30),'go','test','./'+pkg,'-count=1','-p=1','-parallel=4','-json','-timeout='+str(limit)+'s'];logname=f'package-{index:02d}.log';start=time.monotonic()
 with(out/logname).open('w')as log:r=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT)
 events=[];packages_done=[]
 for line in(out/logname).read_text().splitlines():
  try:e=json.loads(line)
  except:continue
  if e.get('Action')in('pass','fail','skip'):
   if e.get('Test'):events.append(e)
   else:packages_done.append(e)
 names={e['Test']for e in events};leaves=[e for e in events if not any(n.startswith(e['Test']+'/')for n in names)];over=[dict(test=e['Test'],seconds=e.get('Elapsed'))for e in leaves if e.get('Elapsed',0)>=60]
 entry=dict(package=pkg,command=command,log=logname,exit=r.returncode,wall_seconds=round(time.monotonic()-start,3),packages=packages_done,test_events=len(events),max_leaf_seconds=max((e.get('Elapsed',0)for e in leaves),default=0),over_budget=over);results.append(entry);(out/'package-results.json').write_text(json.dumps(results,indent=2)+'\n');print(pkg,'exit',r.returncode,'wall',entry['wall_seconds'],'events',len(events),'maxleaf',entry['max_leaf_seconds'],flush=True)
 if r.returncode or over:sys.exit(1)
