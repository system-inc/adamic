import pathlib,subprocess,json,difflib,time
root=pathlib.Path.cwd();ev=root/'review/compiler/loom-views-fx-d';rows=[]
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
  (ev/'extra-mutants.json').write_text(json.dumps(rows,indent=2)+'\n')
mutant('cast-metadata-isolated', [('internal/native/emit_objects.go','needs.fieldTypes = needs.fieldTypes || expression.CheckedFields','needs.fieldTypes = needs.fieldTypes')], './internal/native','^TestViewCheckedCastMetadata$')
mutant('clear-contract-bad-callee', [('internal/lower/readiness.go','expression.Method && !certifiedUntaggedCallableRead(program, program.ViewContractTypes[expression.ViewTypeID])','expression.Method')], './internal/oracle','^TestLoomViewsDBadCallable$')
mutant('native-static-kind', [('internal/native/runtime/view_unions_mixed.c','else if (object->class != NULL && object->class->is_static)','else if (false && object->class != NULL && object->class->is_static)')], './internal/oracle','^TestFX7TextP67$')
mutant('callable-spread', [('internal/lower/expression.go','property.ViewContract > 0 && len(spread) != 0','property.ViewContract > 0 && len(spread) != 0 && false')], './internal/lower','^TestCallableContractSpreadRefused$')
mutant('producer-mask', [('internal/native/view_unions_callable_producers.go','masks[i] = e.program.ViewContracts[child-1].RepresentationMask\n\t\t\t}','masks[i] = 0\n\t\t\t}')], './internal/oracle','^TestLoomViewsDWideCallable$')
