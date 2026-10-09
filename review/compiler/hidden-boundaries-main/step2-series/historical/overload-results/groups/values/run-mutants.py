"""Remove an indirect boundary, then only boundaries through a single signature."""
import json, os, subprocess, tempfile
from pathlib import Path
root=Path(__file__).resolve().parents[4]
out=Path(__file__).parent
source=root/'internal/lower/overload_results.go'
original=source.read_text()
needle='body = append(body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: test}, Then: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}}}})'
assert original.count(needle)==1
rows=[]
for name,guard,fixture in [('drop-indirect-check','closure == nil','returned'),('skip-single-signature-check','resolved.Declaration() == overload','narrow')]:
 with tempfile.TemporaryDirectory(prefix='overload-values-mutant-') as directory:
  directory=Path(directory)
  replacement=directory/source.name
  replacement.write_text(original.replace(needle,'if '+guard+' { '+needle+' }'))
  overlay=directory/'overlay.json'
  overlay.write_text(json.dumps({'Replace':{str(source):str(replacement)}}))
  command=['go','test','-overlay',str(overlay),'./internal/oracle','-run','^TestOverloadValues$/^'+fixture+'$/^liar$','-count=1','-v']
  with (out/(name+'.log.txt')).open('w') as log:
   result=subprocess.run(command,cwd=root,env=os.environ,stdout=log,stderr=subprocess.STDOUT)
  text=(out/(name+'.log.txt')).read_text()
  caught=result.returncode!=0 and 'unchecked wrong result' in text and '[build failed]' not in text
  rows.append({'mutant':name,'caught':caught,'exit':result.returncode})
  print(rows[-1],flush=True)
(out/'mutants.json').write_text(json.dumps(rows,indent=2)+'\n')
assert all(row['caught'] for row in rows)
