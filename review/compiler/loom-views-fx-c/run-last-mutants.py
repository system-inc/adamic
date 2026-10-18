import pathlib,subprocess,json,difflib,time
root=pathlib.Path.cwd();ev=root/'review/compiler/loom-views-fx-c';rows=[]
def mutant(name,changes,pkg,test):
 originals={}
 try:
  patches=[]
  for path,old,new in changes:
   p=root/path;originals[p]=p.read_text();assert old in originals[p],(name,path,old)
   modified=originals[p].replace(old,new,1);p.write_text(modified)
   patches.extend(difflib.unified_diff(originals[p].splitlines(True),modified.splitlines(True),fromfile=path,tofile=path))
  (ev/(name+'.patch')).write_text(''.join(patches))
  cmd=['go','test',pkg,'-run',test,'-count=1','-v','-timeout','90s']
  start=time.monotonic()
  with (ev/(name+'.log')).open('w') as f:p=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,timeout=120)
  log=(ev/(name+'.log')).read_text();caught=p.returncode!=0 and '--- FAIL:' in log and '[build failed]' not in log
  rows.append(dict(name=name,command=cmd,exit=p.returncode,caught=caught,seconds=round(time.monotonic()-start,3)))
  print(name,caught,flush=True)
 finally:
  for p,s in originals.items():p.write_text(s)
  (ev/'last-mutants.json').write_text(json.dumps(rows,indent=2)+'\n')
mutant('unrecorded-unique-result', [('internal/lower/view_callable_producers.go','if len(assignableMembers) != 0 && !compatible {','if len(assignableMembers) != 0 && !compatible && false {')], './internal/lower','^TestCallableProducerUniqueObjectResultRefused$')
mutant('packed-number-stale', [('internal/native/runtime/object.c','number.present ? number.number : 0.0','number.number')], './internal/native','^TestViewSnapshotPayloadOwnership$')
