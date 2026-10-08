"""Run assignment-value semantic and analysis mutants, restoring every source."""
import os
import pathlib
import subprocess

scratch = pathlib.Path('/tmp/assignment-mutants')
scratch.mkdir(exist_ok=True)
native = 'internal/native/assignment_value.go'
mutants = [
 ('double-right', native, 'value := e.snapshot(raw.Type(), e.value(raw))', 'e.value(raw)\n value := e.snapshot(raw.Type(), e.value(raw))', 'oracle'),
 ('double-field-receiver', native, 'object := e.value(store.Object)', 'e.value(store.Object)\n object := e.value(store.Object)', 'oracle'),
 ('double-element-receiver', native, 'array := e.value(store.Array)', 'e.value(store.Array)\n array := e.value(store.Array)', 'oracle'),
 ('double-element-index', native, 'index := e.snapshot(ir.Number, e.value(store.Index))', 'e.value(store.Index)\n index := e.snapshot(ir.Number, e.value(store.Index))', 'oracle'),
 ('old-local', native, 'case ir.Assign:\n\t\tvalue, stored :=', 'case ir.Assign:\n old := ""\n if expression.Type() == ir.Number && store.Value.Type() == ir.Number { old = e.snapshot(ir.Number, e.read(ir.Read{Local: store.Local, Of: ir.Number})) }\n\t\tvalue, stored :=', 'old-local'),
 ('old-field', native, 'object := e.value(store.Object)', 'object := e.value(store.Object)\n old := ""\n if expression.Type() == ir.Number && store.Value.Type() == ir.Number { old = e.snapshot(ir.Number, e.fieldSlot(object, store.Name, store.Class)+"->number") }', 'old-field'),
 ('old-element', native, 'index := e.snapshot(ir.Number, e.value(store.Index))', 'index := e.snapshot(ir.Number, e.value(store.Index))\n old := ""\n if expression.Type() == ir.Number && store.Value.Type() == ir.Number && !store.Array.Type().IsTypedArray() { old = e.snapshot(ir.Number, array+"->elements[(size_t)("+index+")].number") }', 'old-element'),
 ('late-local-read','internal/native/emit_locals.go', ' && !e.program.Locals[read.Local].ExpressionAssigned', '', 'oracle'),
 ('lose-result-alias','internal/fresh/fresh.go','a.write(WriteField, store.Site, store.Name, holder, stored, store.Name)\n\t\t\treturn held','a.write(WriteField, store.Site, store.Name, holder, stored, store.Name)\n\t\t\t_ = held\n\t\t\treturn value{}','cycle'),
]
for name,file,before,after,kind in mutants:
 if os.environ.get("ASSIGNMENT_MUTANT_ONLY") and name != os.environ["ASSIGNMENT_MUTANT_ONLY"]: continue
 path=pathlib.Path(file); original=path.read_text(); assert original.count(before)==1,(name,original.count(before))
 try:
  changed=original.replace(before,after)
  if kind.startswith('old-'):
   section={'old-local':'case ir.Assign:', 'old-field':'case ir.SetProperty:', 'old-element':'case ir.SetIndex:'}[kind]
   start=changed.index(section); end=changed.index('\n\tcase ',start+1) if '\n\tcase ' in changed[start+1:] else changed.index('\n\t}',start)
   part=changed[start:end];assert part.count('return value')==1
   changed=changed[:start]+part.replace('return value','if old != "" { return old }; return value')+changed[end:]
  path.write_text(changed)
  log=scratch/(name+'.log')
  args=['go','test','./internal/oracle','-run','TestNativeAgreesWithNode/internal/oracle/testdata/assignment_value','-count=1','-timeout','15m']
  if kind=='cycle':args=['go','test','./internal/lower','-run','TestAssignmentValuePreservesCycleChecks','-count=1']
  with log.open('w') as output:
   result=subprocess.run(args,stdout=output,stderr=output,env=dict(os.environ,ADAMIC_GATE_UNCACHED='1'))
  text=log.read_text()
  expected='want assignment-result cycle refused' if kind=='cycle' else 'stdout differs'
  assert result.returncode!=0 and expected in text,(name,text)
  assert '[build failed]' not in text and 'error:' not in text,(name,text)
  print(name,'caught, exit',result.returncode,'log',log,flush=True)
 finally:path.write_text(original)
