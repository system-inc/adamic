"""Restore the optional receiver stop and remove required-read throw paths separately."""
import json, os, subprocess, tempfile
from pathlib import Path
root=Path.cwd()
out=Path(__file__).parent/'optional-chain';out.mkdir(exist_ok=True)
specs=[
('check-optional-receiver','internal/lower/narrowed.go','l.acceptsUndefined(node) || optionalReceiver(node)','(l.acceptsUndefined(node) && !optionalReceiver(node))','TestOptionalAfterCall79'),
('javascript-required-panic','internal/javascript/javascript.go','if expression.Throws() {','if false && expression.Throws() {','TestOptionalAfterCallRequired'),
('native-required-panic','internal/native/reuse.go','if (ir.Defined{Message: message}).Throws() {','if false && (ir.Defined{Message: message}).Throws() {','TestOptionalAfterCallRequired'),
('drop-flow-throw-edge','internal/flow/build.go','if value.Interface().(ir.Defined).Throws() {','if false && value.Interface().(ir.Defined).Throws() {','TestOptionalAfterCallFlowEdges'),
('drop-required-propagation','internal/lower/exceptions.go','found = node.Throws()','found = false','TestOptionalAfterCallPropagation'),
]
rows=[]
for name,file,needle,replacement,test in specs:
 source=root/file;original=source.read_text();assert original.count(needle)==1,(name,original.count(needle))
 with tempfile.TemporaryDirectory(prefix='optional-call-mutant-') as directory:
  directory=Path(directory);changed=directory/(source.name+'.txt');changed.write_text(original.replace(needle,replacement))
  overlay=directory/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(source):str(changed)}}))
  package = './internal/flow' if test == 'TestOptionalAfterCallFlowEdges' else './internal/oracle'
  command=['go','test','-overlay',str(overlay),package,'-run','^'+test+'$','-count=1','-timeout','60s','-v']
  logpath=out/(name+'.log')
  with logpath.open('w') as log:r=subprocess.run(command,cwd=root,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'},stdout=log,stderr=subprocess.STDOUT,timeout=120)
  text=logpath.read_text();caught=r.returncode!=0 and '--- FAIL:' in text and ('backend exit=' in text or 'catchable read lost its flow exception edge' in text) and '[build failed]' not in text
  rows.append(dict(mutant=name,caught=caught,exit=r.returncode,command=command));print(name,caught,flush=True)
(out/'mutants.json').write_text(json.dumps(rows,indent=2)+'\n')
assert all(row['caught'] for row in rows)
