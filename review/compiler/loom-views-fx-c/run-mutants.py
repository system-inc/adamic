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
  (ev/'mutants.json').write_text(json.dumps(rows,indent=2)+'\n')
mutant('packed-snapshot-stale', [('internal/native/runtime/object.c','value.payload = (adamic_value){0};\n        if (boolean.present) { value.payload.boolean = boolean.boolean; }','value.payload.boolean = boolean.boolean;')], './internal/native','^TestViewSnapshotPayloadOwnership$')
mutant('native-direct-code', [('internal/native/view_unions_untagged.go','if !e.program.Functions[function].Closure || e.program.Functions[function].Receiver {','if false {')], './internal/oracle','^TestCallableProducerRegistryFiltersDirectFunctions$')
mutant('lower-direct-code', [('internal/lower/view_unions_untagged.go','if !function.Closure || function.Receiver {','if false {'),('internal/lower/view_callable_producers.go','!function.Closure || function.Receiver || ','')], './internal/lower','^TestCallableProducerCertificatesUseClosureThunks$')
mutant('exact-producer-type', [('internal/lower/view_unions_untagged.go','l.checker.IsTypeAssignableTo(producer.proven, target)','producer.proven == target')], './internal/oracle','^TestCheckedViewV2CallableProducerMutant$')
mutant('ambiguous-producer', [('internal/lower/view_callable_producers.go','if len(assignableMembers) > 1 {','if false {')], './internal/lower','^TestCallableProducerAmbiguousRefused$')
mutant('missing-abi-refusal', [('internal/lower/view_callable_producers.go','if len(assignableMembers) != 0 && !compatible {','if len(assignableMembers) != 0 && !compatible && false {')], './internal/lower','^TestCallableProducerMethodShorthandRefused$|^TestCallableProducerExtraOptionalRefused$')
mutant('unrecorded-result', [('internal/lower/view_callable_producers.go','if len(assignableMembers) != 0 && !compatible {','if len(assignableMembers) != 0 && !compatible && l.result.Functions[function].Returns != ir.Object {')], './internal/lower','^TestCallableProducerDiscardedObjectResultRefused$')
mutant('raw-union-result', [('internal/native/emit_functions.go','if expression.Returns == ir.Union {\n\t\treturn e.callableUnionResult(result, closure, method)\n\t}','')], './internal/oracle','^TestCallableUnionResultNumber$')
