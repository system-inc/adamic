import json, os, pathlib, subprocess
root = pathlib.Path('/workspace/adamic')
p = root/'internal/lower/phantom_array_presence.go'
original = p.read_text()
mutants = [
 ('in-refusal-dropped', 'return l.phantomArrayObservation(node, binary.Right, binary.Left, "in", false)', 'return nil', 'TestPhantomArrayPresenceNarrowing'),
 ('own-method-refusal-dropped', 'return l.phantomArrayObservation(node, receiver, key, name, false)', '_ = key; return nil', 'TestPhantomArrayPresenceRefusals'),
 ('object-reflection-refusal-dropped', 'return l.phantomArrayObservation(node, arguments[0], key, "Object."+name, false)', '_ = key; return nil', 'TestPhantomArrayPresenceRefusals'),
 ('spread-refusal-dropped', 'return l.phantomArrayObservation(node, node.AsSpreadAssignment().Expression, nil, "object spread", false)', 'return nil', 'TestPhantomArrayPresenceRefusals'),
 ('assign-source-refusal-dropped', 'if err := l.phantomArrayObservation(node, source, nil, "Object.assign", false); err != nil {', 'if err := l.phantomArrayObservation(node, source, nil, "Object.assign", true); false && err != nil {', 'TestPhantomArrayPresenceRefusals'),
 ('json-refusal-dropped', 'return l.phantomArrayObservation(node, arguments[0], nil, "JSON.stringify", false)', 'return nil', 'TestPhantomArrayPresenceRefusals'),
 ('write-refusal-dropped', 'if write {', 'if write { return nil }; if write {', 'TestPhantomArrayPresenceRefusals'),
 ('generic-constraint-forgotten', 'return l.phantomArrayFields(l.checker.GetBaseConstraintOfType(proven), seen)', 'return nil', 'TestPhantomArrayPresenceViews'),
 ('union-arm-forgotten', 'fields = append(fields, l.phantomArrayFields(part, seen)...)', '_ = part; fields = append(fields, l.phantomArrayFields(proven, seen)...)', 'TestPhantomArrayPresenceNarrowing'),
 ('member-name-forgotten', 'name, known = key.Text(), true', 'name, known = "first", true', 'TestPhantomArrayWritesNameTheMember'),
]
results=[]
try:
 for name, needle, replacement, test in mutants:
  assert original.count(needle)==1, name
  p.write_text(original.replace(needle,replacement))
  with open('/tmp/phantom-presence-mutant-'+name+'.log','w') as log:
   run=subprocess.run(['go','test','./internal/lower','-run','^'+test+'$','-count=1'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  results.append({'mutant':name,'test':test,'exit':run.returncode})
  p.write_text(original)
finally:
 p.write_text(original)
 pathlib.Path('/tmp/phantom-presence-mutants.json').write_text(json.dumps(results,indent=2)+'\n')
assert len(results)==len(mutants) and all(x['exit']==1 for x in results), results
