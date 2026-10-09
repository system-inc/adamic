import json, os, subprocess, tempfile
from pathlib import Path
root=Path(__file__).resolve().parents[4]
out=Path(__file__).parent
source=root/'internal/lower/overload_results.go'
original=source.read_text()
needle='body = append(body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: test}, Then: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}}}})'
assert original.count(needle)==1
changed=original.replace(needle,'if field == "" { '+needle+' }')
rows=[]
for field in ['kind','value']:
 with tempfile.TemporaryDirectory(prefix='overload-hatch-mutant-') as scratch:
  scratch=Path(scratch);replacement=scratch/source.name;replacement.write_text(changed)
  overlay=scratch/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(source):str(replacement)}}))
  command=['go','test','-overlay',str(overlay),'./internal/oracle','-run','^TestOverloadFieldHatch$/^'+field+'$/^liar$','-count=1','-v']
  logpath=out/('drop-'+field+'-check.log.txt')
  with logpath.open('w') as log:r=subprocess.run(command,cwd=root,env=os.environ,stdout=log,stderr=subprocess.STDOUT)
  text=logpath.read_text();caught=r.returncode!=0 and 'unchecked wrong result' in text and '[build failed]' not in text
  rows.append({'mutant':'drop-'+field+'-check','caught':caught,'exit':r.returncode});print(rows[-1],flush=True)
(out/'mutants.json').write_text(json.dumps(rows,indent=2)+'\n')
assert all(r['caught'] for r in rows)
