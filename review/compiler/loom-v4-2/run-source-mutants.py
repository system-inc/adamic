import ast,difflib,json,os,re,subprocess,time
from pathlib import Path
root=Path.cwd();e=root/'review/compiler/loom-v4-2'
cases=[]
for name in ['Argument','Result','Passed','Returned','MixedProducer']:
 cases.append(('escape-'+name,'internal/lower/view_callable_escape.go','if graph.viewCallableEscapeProven(read, reaches) {','if true {','TestV4Escape'+name+'Refusal'))
cases += [
 ('unsupported-domain','internal/lower/view_lazy.go','call.CallContract == 0 ||','call.CallContract == -1 ||','TestV4DirectRecursiveRefusal'),
 ('unsupported-producer','internal/lower/view_lazy.go','graph.viewCallableUncheckableProducer(property, reaches)','false','TestV4DirectProducerRefusal'),
 ('surface-borrow','internal/native/borrow.go',' || expression.ViewEscape && !expression.ViewEscapeAdamic','','TestV4EscapeAdapterName'),
 ('bind-name-borrow','internal/native/borrow.go',' || expression.Object.Type() == ir.Closure && expression.Name == "name"','','TestV4EscapeAdapterBindNameOwnership'),
 ('identity-root','internal/native/runtime/closure.c','return value->view != NULL ? value->view->underlying : value;','return value;','TestV4EscapeAdapterIdentity'),
]
cases.extend([
 ('metadata-inline','internal/native/view_arrays.go','e.line("adamic_array_view_storage(%s, %d);", value, kind)','e.line("%s->element_kind = %d;", value, kind)','TestPortElementMetadataEmission'),
 ('metadata-omitted','internal/native/view_arrays.go','e.line("adamic_array_view_storage(%s, %d);", value, kind)','e.line("(void)%s; /* omitted storage certificate %d */", value, kind)','TestPortElementMetadataEmission'),
])
for node in ast.parse((root/'review/compiler/views-v4/run-order-mutants.py').read_text()).body:
 if isinstance(node,ast.Assign) and any(isinstance(t,ast.Name) and t.id=='cases' for t in node.targets):
  cases.extend(tuple(row)+('TestV4DirectCalleeOrder',) for row in ast.literal_eval(node.value))
cases=[(n,f,b.replace('invocation := b.String() +','invocation := e.viewCallableInvoke(call, p) +'),a.replace('invocation := b.String() +','invocation := e.viewCallableInvoke(call, p) +'),t) for n,f,b,a,t in cases]
cases.extend([
 ('construct-witness','internal/oracle/testdata/views_v4_refused/construct-no-signature.a','new escaped(','escaped(','TestV4EscapeAdapterWithoutConstructSignatureRefusal'),
 ('boxed-witness','internal/oracle/testdata/views_v4_refused/view-result-wrong.a','const callback = viewed.run;','const callback = () => viewed.run();','TestV4Lane5BoxedAdamicRefusal'),
 ('factory-witness','internal/oracle/testdata/views_v4_refused/factory-165-wrong-arity.a','interface Target {onEmitNode:','interface Target {readonly onEmitNode:','TestV4Lane5FactoryWiderWriteRefusal'),
])
completed={__import__('json').loads(line)['mutant'] for line in (e/'source-mutants.jsonl').read_text().splitlines()} if (e/'source-mutants.jsonl').exists() else set()
cases=[case for case in cases if case[0] not in completed]
with (e/'source-mutants.jsonl').open('a') as result:
 for name,file,before,after,test in cases:
  p=root/file;original=p.read_text()
  if original.count(before)!=1:
   result.write(json.dumps(dict(mutant=name,blocked='mutation site changed'))+'\n');result.flush();continue
  changed=original.replace(before,after,1)
  (e/(name+'.patch')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
  try:
   p.write_text(changed)
   with (e/(name+'-mutant.log')).open('w') as log:
    start=time.monotonic()
    try:r=subprocess.run(['go','test','./stage1/cohere/json' if test=='TestPortElementMetadataEmission' else './internal/oracle','-run','^'+test+'$','-count=1','-timeout','90s','-v'],stdout=log,stderr=subprocess.STDOUT,timeout=120,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'});code=r.returncode
    except subprocess.TimeoutExpired:code=124
   output=(e/(name+'-mutant.log')).read_text()
   killed=code==1 and '--- FAIL: '+test in output and '[build failed]' not in output
   row=dict(mutant=name,test=test,seconds=round(time.monotonic()-start,3),exit=code,killed=killed)
   result.write(json.dumps(row)+'\n');result.flush();print(row,flush=True)
  finally:p.write_text(original)
