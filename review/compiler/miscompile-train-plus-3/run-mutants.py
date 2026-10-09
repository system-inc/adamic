from pathlib import Path
import subprocess, time, json, difflib
out=Path(__file__).parent
paths=[Path('internal/lower/readiness.go'),Path('internal/lower/view_member_read.go')]
saved={p:p.read_text() for p in paths}
results=[]
def run(name, package, regex):
 command=['go','test',package,'-run',regex,'-count=1','-v','-timeout','80s']
 start=time.monotonic()
 with (out/(name+'.log')).open('w') as log:
  result=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=90)
 log=(out/(name+'.log')).read_text()
 caught=result.returncode!=0 and '--- FAIL: Test' in log and '[build failed]' not in log
 results.append(dict(name=name,command=command,exit=result.returncode,caught=caught,seconds=round(time.monotonic()-start,3)))
 (out/'mutants.json').write_text(json.dumps(results,indent=2)+'\n')
 print(results[-1],flush=True)
 assert caught
try:
 for p in paths:
  base=subprocess.check_output(['git','show','51aa3a96:'+str(p)],text=True)
  p.write_text(base)
 (out/'revert-fix.patch').write_text(''.join(''.join(difflib.unified_diff(saved[p].splitlines(True),p.read_text().splitlines(True),fromfile='a/'+str(p),tofile='b/'+str(p))) for p in paths))
 run('mutant-revert-fix','./internal/native','^TestInheritanceMemoryPlans$')
 for p,text in saved.items(): p.write_text(text)
 p=paths[1]; original=saved[p]
 needle='func (l *lowering) checkedViewMembers(node, source *ast.Node, value ir.Expression, member *ast.Node) (ir.Expression, error) {'
 mutant=original.replace(needle,needle+'\n\treturn value, nil',1)
 assert mutant!=original
 p.write_text(mutant)
 (out/'drop-member-check.patch').write_text(''.join(difflib.unified_diff(original.splitlines(True),mutant.splitlines(True),fromfile='a/'+str(p),tofile='b/'+str(p))))
 run('mutant-drop-member-check','./internal/oracle','^TestCheckedView')
finally:
 for p,text in saved.items(): p.write_text(text)
