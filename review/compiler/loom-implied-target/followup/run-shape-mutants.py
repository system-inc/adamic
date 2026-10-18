from pathlib import Path
import subprocess,difflib,json,time
p=Path('internal/lower/cycles.go');ev=Path('review/compiler/loom-implied-target/followup');original=p.read_text();rows=[]
mutants=[
 ('shape-ignores-readonly', ':%p:%d:%t;", value, field.Flags, f.l.checker.IsReadonlySymbol(field)', ':%p:%d;", value, field.Flags', '^TestCycleScalarShapeKeysPreserveDistinctFields$', 'shared a shape key'),
 ('shape-per-literal','if f.scalarShapes[key] {','if false && f.scalarShapes[key] {','^TestCycleScalarLiteralShapesOnce$','scalar literal shapes = 2'),
 ('shape-ignores-type',':%p:%d:%t;", value, field.Flags',':%d:%t;", field.Flags','^TestCycleScalarShapeKeysPreserveDistinctFields$','shared a shape key'),
 ('shape-ignores-name','key.WriteString(field.Name)','key.WriteString("")','^TestCycleScalarShapeKeysPreserveDistinctFields$','shared a shape key'),
 ('shape-includes-reference','if !boolean && flags & ^(','if false && !boolean && flags & ^(','^TestCycleScalarShapeKeysPreserveDistinctFields$','want 7/4'),
]
try:
 for name,old,new,test,marker in mutants:
  assert old in original,old
  changed=original.replace(old,new,1)
  (ev/(name+'.patch')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile=str(p),tofile=str(p))))
  p.write_text(changed);start=time.monotonic()
  with (ev/(name+'.log')).open('w') as log:
   r=subprocess.run(['timeout','120','go','test','./internal/lower','-run',test,'-count=1','-v','-timeout','90s'],stdout=log,stderr=subprocess.STDOUT,timeout=125)
  text=(ev/(name+'.log')).read_text()
  caught=r.returncode!=0 and marker in text and '--- FAIL:' in text and '[build failed]' not in text
  rows.append(dict(mutant=name,exit=r.returncode,caught=caught,seconds=round(time.monotonic()-start,3)))
  print(name,caught,flush=True)
  p.write_text(original)
finally:
 p.write_text(original)
 (ev/'shape-mutants.json').write_text(json.dumps(rows,indent=2)+'\n')
