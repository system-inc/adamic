import pathlib,subprocess,os
root=pathlib.Path(__file__).resolve().parents[3]
logs=pathlib.Path(os.environ.get('ADAMIC_LOOP_LOGS', '/tmp/adamic-loop-mutants'))
logs.mkdir(parents=True, exist_ok=True)
paths=['internal/native/element_borrow.go','internal/native/borrow.go','internal/native/emit_statements.go']
original={p:(root/p).read_text() for p in paths}
oracle=['go','test','./internal/oracle','-run','TestNativeAgreesWithNode/internal/oracle/testdata/borrow_(loop|global_call)','-count=1']
plan=['go','test','./internal/native','-run','TestLoopBorrowPlan|TestNbodyBorrowedLoopC','-count=1']
cases=[
 ('set-index',paths[0],'if _, isStore := statement.(ir.SetIndex); isStore {','if _, isStore := statement.(ir.SetIndex); false && isStore {',oracle),
 ('pop',paths[0],'case ir.ObjectLiteral, ir.ArrayPush,','case ir.ArrayPop, ir.ObjectLiteral, ir.ArrayPush,',oracle),
 ('splice',paths[0],'case ir.ObjectLiteral, ir.ArrayPush,','case ir.ArraySplice, ir.ObjectLiteral, ir.ArrayPush,',oracle),
 ('named-call',paths[0],'expression.Virtual == 0 && !changing[expression.Function]','expression.Virtual == 0',oracle),
 ('closure-call',paths[0],'case ir.ObjectLiteral, ir.ArrayPush,','case ir.CallClosure, ir.ObjectLiteral, ir.ArrayPush,',oracle),
 ('array-assigned',paths[0],'&& !assigned[array.Local]','',plan),
 ('binding-assigned',paths[0],'|| assigned[declare.Local]','',plan),
 ('binding-captured',paths[0],'local.Global || local.Captured ||','local.Global ||',plan),
 ('binding-owned',paths[0],'program.Locals[loop.Local].Borrowed = true','program.Locals[loop.Local].Borrowed = false',oracle),
 ('lending-fact',paths[0],'lending[loop.Iterable.(ir.Read).Local] = true','lending[loop.Iterable.(ir.Read).Local] = false',plan),
 ('borrow-disabled',paths[0],'ok && loopBorrowable(program, index, loop, assigned)','ok && false && loopBorrowable(program, index, loop, assigned)',plan),
 ('throw-owned',paths[2],'e.line("%s %s = %s;", cType(statement.Element), e.localName(statement.Local), element)','e.line("%s %s = %s;", cType(statement.Element), e.localName(statement.Local), element)\n\t\t\tif e.function.Name == "throwBody" { e.hold(e.localName(statement.Local)) }',oracle),
 ('iterator-hold',paths[2],'held, e.kept(iterable))','held, func() string { if e.function != nil && e.function.Name == "reassignBody" { return iterable }; return e.kept(iterable) }())',oracle),
 ('global-touches',paths[1],'touches(e.program, target, read.Local, map[int]bool{}) ||','false ||',oracle),
 ('later-argument',paths[1],'if !pure(argument) {','if false && !pure(argument) {',oracle),
]
env=os.environ.copy();env['ADAMIC_GATE_UNCACHED']='1'
cases += [
 ('virtual-call',paths[0],'expression.Virtual == 0 && !changing[expression.Function]','!changing[expression.Function]',oracle),
 ('global-virtual-targets',paths[1],'e.program.CallTargets(call)','[]int{call.Function}',oracle),
 ('iterator-transfer',paths[2],'held, e.kept(iterable))','held, func() string { for _, owned := range e.owned { if owned == iterable { return iterable } }; return e.kept(iterable) }())',oracle),
 ('error-construction',paths[0],'ir.MakeClosure, ir.MakeError, ir.Defined','ir.MakeClosure, ir.Defined',plan),
 ('global-disabled',paths[1],'parameters := e.program.Functions[call.Function].Parameters','if true { return "", false }; parameters := e.program.Functions[call.Function].Parameters',['go','test','./internal/native','-run','TestGlobalArgumentLending','-count=1']),
]
cases += [('element-virtual-call',paths[0],'expression.Virtual == 0 && !changing[expression.Function]','!changing[expression.Function]',['go','test','./internal/oracle','-run','TestNativeAgreesWithNode/internal/oracle/testdata/borrow_element_virtual_store','-count=1'])]
selected = os.environ.get('ADAMIC_LOOP_MUTANTS', '').split(',')
if selected != ['']:
 cases = [case for case in cases if case[0] in selected]
for name,p,old,new,cmd in cases:
 try:
  assert old in original[p],name
  (root/p).write_text(original[p].replace(old,new,1))
  with open(logs/(name+'.log'),'w') as log:
   result=subprocess.run(cmd,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
  text=pathlib.Path(logs/(name+'.log')).read_text()
  evidence='ASan' if 'AddressSanitizer' in text else ('assertion' if result.returncode else 'SURVIVED')
  print(name,result.returncode,evidence,flush=True)
  if result.returncode == 0 or ('AddressSanitizer' not in text and '--- FAIL: Test' not in text):
   raise RuntimeError('mutant was not caught by a runtime or behavior assertion: '+name)
 finally:
  for path,s in original.items(): (root/path).write_text(s)
