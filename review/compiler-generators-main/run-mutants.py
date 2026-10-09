#!/usr/bin/env python3
"""Apply source overlays without changing the delivery tree; save non-Go evidence."""
import json
import os
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[2]
evidence = Path(__file__).resolve().parent
rules = [
 ('skip-finally', 'TestGeneratorCancellation', [('internal/lower/generator_machine.go', 'for i := depth; i < len(ctx.finals); i++ {', 'for i := len(ctx.finals); i < len(ctx.finals); i++ {')], 'stdout differs'),
 ('delegate-return', 'TestGeneratorDelegate', [('internal/lower/generator_delegate.go', 'abruptReturn, abruptThrow = invoke("return"), invoke("throw")', 'abruptReturn, abruptThrow = ordinary, invoke("throw")')], 'stdout differs'),
 ('defer-default', 'TestGeneratorDefaults', [('internal/lower/generator_body.go', 'prefix, body := function.Body[:types.Prologue], function.Body[types.Prologue:]', 'prefix, body := function.Body[:0], function.Body[:]')], 'stdout differs'),
 ('first-next-argument', 'TestGeneratorBasic', [
 ('internal/lower/generator_body.go', 'ir.Assign{Local: input, Value: ir.Undefined{Of: ir.Union}}', 'ir.Assign{Local: input, Value: supplied}'),
 ('internal/lower/generator_expressions.go', 'g.store("input", ir.Undefined{Of: ir.Union}), ir.Return{Value: g.packet(g.slot(held), false)}', 'ir.Return{Value: g.packet(g.field("input", ir.Union), false)}')], 'stdout differs'),
 ('forget-pattern-parameter', 'TestGeneratorDestructuredParameterCycle', [('internal/lower/functions.go', 'l.noteLocal(incoming, l.checker.GetTypeAtLocation(parameter), parameter)', '// Mutant forgets the whole retained parameter type.')], "want the retained whole parameter's cycle refusal"),
 ('admit-cycle', 'TestGeneratorCycleFrame', [('internal/lower/generator_cycles.go', 'if f.reaches(proven, cycleNode{proven: target}) {', 'if false && f.reaches(proven, cycleNode{proven: target}) {')], 'LeakSanitizer'),
]
results=[]
for name, test, edits, witness in rules:
 directory=evidence/name
 directory.mkdir(exist_ok=True)
 mapping={}
 for index,(relative, before, after) in enumerate(edits):
  original=root/relative
  contents=original.read_text()
  if before not in contents:
   raise RuntimeError(f'{name}: mutation absent from {relative}')
  mutant=directory/f'source-{index}.go.txt'
  mutant.write_text(contents.replace(before,after,1))
  mapping[str(original)]=str(mutant)
 overlay=directory/'overlay.json'
 overlay.write_text(json.dumps({'Replace':mapping},indent=2)+'\n')
 package = './internal/lower' if name == 'forget-pattern-parameter' else './internal/oracle'
 command=['go','test','-overlay',str(overlay),package,'-run','^'+test+'$','-timeout','60s','-count=1','-v']
 with (directory/'test.log').open('w') as log:
  result=subprocess.run(command,cwd=root,stdout=log,stderr=subprocess.STDOUT,env=os.environ)
 text=(directory/'test.log').read_text()
 caught=result.returncode!=0 and witness in text and '[build failed]' not in text
 results.append({'mutant':name,'command':command,'exit':result.returncode,'caught':caught,'witness':witness})
 print(name,result.returncode,'caught' if caught else 'NOT CAUGHT',flush=True)
(evidence/'mutants.json').write_text(json.dumps(results,indent=2)+'\n')
if not all(row['caught'] for row in results): raise SystemExit(1)
