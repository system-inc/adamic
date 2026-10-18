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
  (ev/'extra-mutants.json').write_text(json.dumps(rows,indent=2)+'\n')
mutant('lower-direct-code-fixed-mutant', [('internal/lower/view_unions_untagged.go','if !function.Closure || function.Receiver {','if (!function.Closure || function.Receiver) && false {'),('internal/lower/view_callable_producers.go','!function.Closure || function.Receiver || ','')], './internal/lower','^TestCallableProducerCertificatesUseClosureThunks$')
mutant('snapshot-retain-unmasked', [('internal/native/runtime/object.c','value.payload = (adamic_value){0};\n        if (boolean.present) { value.payload.boolean = boolean.boolean; }','value.payload.boolean = boolean.boolean;'),('internal/native/view_unions_read.go','e.line("else if (%s.kind >= adamic_view_union_string && %s.kind <= adamic_view_union_function) %s = adamic_retain(%s.payload.reference);", snapshot, snapshot, boxed, snapshot)\n\te.line("else %s = NULL;", boxed)','e.line("else %s = adamic_retain(%s.payload.reference);", boxed, snapshot)')], './internal/oracle','^TestLoomViewsCOriginal$')
mutant('tagged-lazy-check', [('internal/lower/view_callable_producers.go','if contract.Kind == ir.ViewObject {','if contract.Kind == ir.ViewObject && false {')], './internal/lower','^TestCallableProducerTaggedUnreadMethod$')
