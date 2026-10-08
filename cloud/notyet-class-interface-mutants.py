#!/usr/bin/env python3
"""Run focused rule mutants, restoring every changed byte even on a failed run."""
from pathlib import Path
import subprocess
import sys

root = Path(__file__).resolve().parents[1]
logs = Path(sys.argv[1] if len(sys.argv) > 1 else '/tmp/class-interface-mutants')
logs.mkdir(parents=True, exist_ok=True)
helper = 'internal/lower/class_interface.go'
fixture = '^TestNativeAgreesWithNode/internal/oracle/testdata/class_interface_'
mutants = [
 ('erase-inferred-origin', helper,
  'return l.classInterfaceOrigin(declaration.Initializer(), seen)',
  'return nil', 'presence', 'stdout differs'),
 ('static-value-dispatch', helper,
  'if node.Kind == ast.KindCallExpression && l.classInterfaceMethod(node) {',
  'if false && node.Kind == ast.KindCallExpression && l.classInterfaceMethod(node) {', 'literal', 'stdout differs'),
 ('static-statement-dispatch', 'internal/lower/statements.go',
  'if l.classInterfaceMethod(expression) {', 'if false && l.classInterfaceMethod(expression) {', 'presence', 'stdout differs'),
 ('refuse-structural-literal', 'internal/lower/class_inheritance.go',
  ' || l.classInterfaceLiteral(node, target)', '', 'literal', 'nominal-class'),
 ('unchecked-contextual-class', helper,
  'return l.checkedClassCast(node, value, []*checker.Type{target}, "class/interface downcast: value lacks the class tag")',
  'return value, nil', 'checked', 'want the inserted check to fire'),
 ('unchecked-explicit-class', 'internal/lower/cast_proof.go',
  'return castProof{classes: []*checker.Type{target}}, nil',
  'return castProof{}, nil', 'cast', 'want the inserted check to fire'),
 ('own-only-presence', helper,
  'Prototype: true', 'Prototype: false', 'presence', 'stdout differs'),
 ('invoke-getter-on-presence', 'internal/native/runtime/class_interface_property.c',
  'if (adamic_property_name_equal(key, class->accessors[index].name)) return true;',
  'if (adamic_property_name_equal(key, class->accessors[index].name)) { (void)adamic_accessor_get((adamic_object *)object, class->accessors[index].name); return true; }',
  'presence', 'stdout differs'),
]
for name, filename, before, after, probe, caught in mutants:
 path = root / filename
 original = path.read_text()
 assert before in original, (name, filename)
 try:
  path.write_text(original.replace(before, after, 1))
  logfile = logs / (name + '.log')
  with logfile.open('w') as output:
   result = subprocess.run(['go', 'test', './internal/oracle', '-run', fixture + probe, '-count=1', '-v', '-timeout', '10m'], cwd=root, stdout=output, stderr=subprocess.STDOUT)
  evidence = logfile.read_text()
  if result.returncode == 0 or caught not in evidence:
   raise SystemExit(f'{name}: mutant survived or failed for the wrong reason; inspect {logfile}')
  print(f'{name}: killed (exit {result.returncode}, {caught}); {logfile}', flush=True)
 finally:
  path.write_text(original)
