"""Run semantic callback-flow mutants in valid Go binaries, restoring the source."""
from pathlib import Path
import subprocess
p=Path('internal/lower/shape_flow.go');source=p.read_text()
omission='flow.unknownParameters[param+1] = true\n\t\t\t\t\t\tchanged = merge(param+1, ir.FunctionTargets{Unknown: true}) || changed'
mutants={
 'callback-ignore-sibling-target':source.replace('return graph.program.ClosureTargetsWithFlow(call, graph.shapeFunctionTargets)','return graph.shapeFunctionTargets(ir.ClosureValue(call))'),
 'callback-drop-target':source.replace('for _, function := range incoming.Functions {','for _, function := range incoming.Functions[:min(1, len(incoming.Functions))] {'),
 'callback-ignore-unknown-arm':source.replace('return flow.facts[value.Local+1]','valueFact := flow.facts[value.Local+1]; valueFact.Unknown=false; return valueFact'),
 'callback-ignore-escape':source.replace('if !escaped[target] {','if false {'),
 'callback-ignore-omission':source.replace(omission,'changed = merge(param+1, ir.FunctionTargets{Unknown: true}) || changed'),
 'callback-ignore-returned-escape':source.replace('for _, returned := range flow.facts[graph.resultNode(target)].Functions {','for _, returned := range []int{} {'),
}
try:
 for name,changed in mutants.items():
  assert changed!=source,name+' did not mutate'
  p.write_text(changed)
  log=Path('stage3/shape-conformance/logs/'+name+'.log')
  with log.open('w') as stream:
   status=subprocess.run(['go','test','./internal/lower','-run','^TestShapeCallback','-count=1','-v'],stdout=stream,stderr=subprocess.STDOUT).returncode
  output=log.read_text()
  assert status!=0 and 'build failed' not in output and 'panic:' not in output,name+' escaped or did not produce a semantic witness'
  assert 'callback' in output and ('boundary disappeared' in output or 'callback' in output and 'FAIL' in output)
  print(name,'valid Go binary; semantic callback assertion caught mutant',flush=True)
finally:p.write_text(source)
