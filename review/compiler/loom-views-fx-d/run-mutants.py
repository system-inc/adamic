import pathlib,subprocess,json,difflib,time
root=pathlib.Path.cwd();ev=root/'review/compiler/loom-views-fx-d';rows=[]
def mutant(name,changes,pkg,test):
 originals={}
 try:
  patches=[]
  for path,old,new in changes:
   p=root/path;originals.setdefault(p,p.read_text());current=p.read_text();assert old in current,(name,path,old)
   modified=current.replace(old,new,1);p.write_text(modified)
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
mutant('cast-metadata', [('internal/native/emit_objects.go','needs.fieldTypes = needs.fieldTypes || expression.CheckedFields','needs.fieldTypes = needs.fieldTypes')], './internal/oracle','^TestLoomViewsDOriginal$')
mutant('clear-callable-contract', [('internal/lower/readiness.go','expression.Method && !certifiedUntaggedCallableRead(program, program.ViewContractTypes[expression.ViewTypeID])','expression.Method')], './internal/oracle','^TestLoomViewsDCallableUnion$')
mutant('source-argument-layout', [('internal/native/emit_functions.go','packed, count = e.checkedCallableArguments(expression, closure)','packed, count = e.closureArguments(expression)')], './internal/oracle','^TestLoomViewsDWideCallable$')
mutant('native-depth-128', [('internal/native/runtime/view_unions_untagged.c','ADAMIC_CHECK_STACK();','if (depth > 128) { return false; }')], './internal/oracle','^TestFX7Depth65$')
mutant('javascript-depth-128', [('internal/javascript/view_unions_untagged.go','const contract=contracts[id-1];','if(depth>128) return false;\n  const contract=contracts[id-1];')], './internal/oracle','^TestFX7Depth65$')
mutant('native-write-type', [('internal/native/emit_statements.go','expected := statement.DeclaredType','expected := ""')], './internal/oracle','^TestFX7TextP32$')
mutant('javascript-write-type', [('internal/javascript/javascript.go','quote(statement.DeclaredType)','quote("")')], './internal/oracle','^TestFX7TextP32$')
mutant('native-object-kind', [('internal/native/runtime/object.c','if (value->tuple) { found = "array"; }','if (false) { found = "array"; }')], './internal/oracle','^TestFX7TextScalarTuple$')
mutant('native-union-kind', [('internal/native/runtime/view_unions_mixed.c','if (object->tuple) { found = "array"; }','if (false) { found = "array"; }')], './internal/oracle','^TestFX7TextP72$')
mutant('tuple-unmarked', [('internal/javascript/readiness.go','adamicTupleObjects.add(values);','')], './internal/oracle','^TestLoomViewsDTupleUnion$')
mutant('tuple-rejected-object', [('internal/javascript/readiness.go','(!Array.isArray(value) || adamicTupleObjects.has(value))','!Array.isArray(value)')], './internal/oracle','^TestLoomViewsDTupleObject$')
mutant('ordinary-array-admitted', [('internal/javascript/view_unions_untagged.go','(Array.isArray(value) && !adamicTupleObjects.has(value))','false'),('internal/javascript/view_unions_untagged.go','!Array.isArray(value) || !adamicTupleObjects.has(value) || value.length','!Array.isArray(value) || value.length')], './internal/javascript','^TestFX7TupleMarkerRejectsOrdinaryArray$')
