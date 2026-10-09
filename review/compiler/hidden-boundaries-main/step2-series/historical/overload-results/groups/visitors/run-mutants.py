"""Erase the retained input domain, then trust an unproved invocation."""
import json, os, subprocess, tempfile
from pathlib import Path
root=Path(__file__).resolve().parents[4]
out=Path(__file__).parent
mutants=[('erase-TIn-to-Node','overload_visitor_calls.go','l.localTypes[local] = bindings[i]','l.localTypes[local] = l.concrete(l.checker.GetTypeAtLocation(parameter))','helper'),('admit-unproven-invocation','overload_visitor_domain.go','declaration.Body().ForEachChild(scan)\n\treturn result','declaration.Body().ForEachChild(scan)\n\tresult.unproven = nil\n\treturn result','overloaded-helper')]
rows=[]
for name,file,needle,replacement,fixture in mutants:
 source=root/'internal/lower'/file; original=source.read_text(); assert original.count(needle)==1
 with tempfile.TemporaryDirectory(prefix='overload-visitors-mutant-') as directory:
  directory=Path(directory); altered=directory/file; altered.write_text(original.replace(needle,replacement))
  overlay=directory/'overlay.json'; overlay.write_text(json.dumps({'Replace':{str(source):str(altered)}}))
  command=['go','test','-overlay',str(overlay),'./internal/oracle','-run','^TestOverloadVisitors$/^'+fixture+'$/^liar$','-count=1','-v']
  with (out/(name+'.log.txt')).open('w') as log: result=subprocess.run(command,cwd=root,env=os.environ,stdout=log,stderr=subprocess.STDOUT)
  text=(out/(name+'.log.txt')).read_text()
  caught=result.returncode!=0 and 'unproven visitor input admitted' in text and '[build failed]' not in text
  rows.append({'mutant':name,'caught':caught,'exit':result.returncode});print(rows[-1],flush=True)
(out/'mutants.json').write_text(json.dumps(rows,indent=2)+'\n')
assert all(row['caught'] for row in rows)
