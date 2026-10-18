from pathlib import Path
import subprocess,difflib,json,time
root=Path.cwd();ev=root/'review/compiler/loom-implied-target';p=root/'internal/lower/invariance.go';original=p.read_text();rows=[]
for name,condition,test in [
 ('array-per-element-query','ok && parent.Kind != ast.KindArrayLiteralExpression','^TestImpliedTargetLargeArrayQueriesOnce$'),
 ('conditional-per-branch-query','ok && parent.Kind != ast.KindConditionalExpression','^TestImpliedTargetConditionalQueriesOnce$'),
 ('logical-per-operand-query','ok && parent.Kind != ast.KindBinaryExpression','^TestImpliedTarget(Or|And|Nullish)QueriesOnce$'),
 ('nil-cache-miss','ok && target != nil','^TestImpliedTargetTupleCachesNil$')]:
 changed=original.replace('if target, ok := l.impliedTargets[parent]; ok {','if target, ok := l.impliedTargets[parent]; '+condition+' {',1)
 assert changed!=original
 (ev/(name+'.patch')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='internal/lower/invariance.go',tofile='internal/lower/invariance.go')))
 try:
  p.write_text(changed)
  command=['timeout','120','go','test','./internal/lower','-run',test,'-count=1','-v','-timeout','90s'];start=time.monotonic()
  with (ev/(name+'.log')).open('w') as log:r=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=125)
  text=(ev/(name+'.log')).read_text();caught=r.returncode!=0 and 'more than once' in text and '--- FAIL:' in text and '[build failed]' not in text
  rows.append(dict(mutant=name,command=command,exit=r.returncode,caught=caught,seconds=round(time.monotonic()-start,3)))
  print(name,caught,flush=True)
 finally:
  p.write_text(original)
  (ev/'mutants.json').write_text(json.dumps(rows,indent=2)+'\n')
