#!/usr/bin/env python3
"""Focused field-storage and parameter-storage mutations through Go overlays."""
import json, os, subprocess, tempfile
from pathlib import Path
root=Path(__file__).resolve().parents[4]
out=Path(__file__).parent
field=root/'internal/lower/overload_result_fields.go'
constraint=root/'internal/lower/generic_constraint_storage.go'
expression=root/'internal/lower/expression.go'
f=field.read_text(); c=constraint.read_text(); e=expression.read_text()
mutants=[
 ('drop-field-storage',field,f.replace('if !compatible {','if false && !compatible {'),'TestOverloadStructuralRefuses/mixed_field_storage','got <nil>'),
 ('drop-constraint-storage',constraint,c.replace('return ir.Union, true','return 0, false').replace('return held, true','return 0, false'),'TestHiddenTNodeConstraintRepresentation','constraint representation'),
 ('object-layout',constraint,c.replace('return ir.Union, true','return ir.Object, true'),'TestHiddenTNodeConstraintMutationRemainsNotYet','got <nil>'),
 ('array-layout',constraint,c.replace('case ir.Object:','case ir.Array:\n return ir.Array, true\n case ir.Object:'),'TestHiddenTNodeConstraintRepresentation/array_layout','constraint representation'),
 ('unconstrained-layout',constraint,c.replace('return 0, false','return ir.Number, true',1),'TestHiddenTNodeConstraintRepresentation','constraint representation'),
 ('ignore-substitution',expression,e.replace('return substituted, true','_ = substituted; return l.constraintStorage(proven)',1),'TestHiddenTNodeConstraintRepresentation','lost concrete substitution'),
 ('ignore-mapper',expression,e.replace('proven = l.concrete(proven)\n','// mutant: ignore concrete checker mapper\n',1),'TestHiddenTNodeConstraintRepresentation','lost concrete checker mapping'),
]
rows=[]
for name, source, changed, selector, marker in mutants:
 assert changed != source.read_text(),name
 with tempfile.TemporaryDirectory(prefix='overload-fields-mutant-') as tmp:
  tmp=Path(tmp); replacement=tmp/source.name; replacement.write_text(changed)
  overlay=tmp/'overlay.json'; overlay.write_text(json.dumps({'Replace':{str(source):str(replacement)}}))
  selection='/'.join('^'+part+'$' for part in selector.split('/'))
  command=['go','test','-overlay',str(overlay),'./internal/lower','-run',selection,'-count=1','-v']
  logpath=out/(name+'.log.txt')
  with logpath.open('w') as log: result=subprocess.run(command,cwd=root,env=os.environ,stdout=log,stderr=subprocess.STDOUT)
  text=logpath.read_text()
  caught=result.returncode!=0 and '--- FAIL:' in text and '[build failed]' not in text and marker in text
  rows.append(dict(mutant=name,exit=result.returncode,caught=caught,command=command))
  print(name,'caught' if caught else 'NOT CAUGHT',flush=True)
(out/'mutants.json').write_text(json.dumps(rows,indent=2)+'\n')
assert all(row['caught'] for row in rows)
