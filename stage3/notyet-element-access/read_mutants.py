"""Run semantic finite-read mutants independently and restore each edit."""
from pathlib import Path
import subprocess
import sys

tuple_mode = len(sys.argv) > 1 and sys.argv[1] == 'tuple'
presence_mode = len(sys.argv) > 1 and sys.argv[1] == 'presence'
source = Path('internal/lower/element_access_tuples.go' if tuple_mode else 'internal/lower/element_access_fields.go')
fixture = 'element_access_tuple' if tuple_mode else 'element_access_presence' if presence_mode else 'element_access_reads'
label = 'tuple' if tuple_mode else 'presence' if presence_mode else 'read'
original = source.read_text()
mutants = {
 'wrong-dispatch': ('Operator: ir.Equal', 'Operator: ir.NotEqual'),
 'missing-absence': ('Absent: fields[index].Flags&ast.SymbolFlagsOptional != 0', 'Absent: false'),
 'repeated-receiver': ('[]ir.Expression{object, key}', '[]ir.Expression{object, key, object}'),
}
if tuple_mode:
 mutants = {
  'fractional-index': ('Operator: ir.Equal', 'Operator: ir.LessOrEqual'),
  'missing-slot': ('Absent: absent', 'Absent: false'),
  'out-of-range': ('b.finish("element_access_tuple", missing)', 'b.finish("element_access_tuple", fit(ir.Property{Object: heldObject, Name: "0", Of: stored[0]}, result))'),
  'repeated-receiver': ('[]ir.Expression{object, key}', '[]ir.Expression{object, key, object}'),
 }
if presence_mode:
 mutants = {'missing-presence-check': ('return l.computedFieldPresence(node, value), nil', 'return value, nil')}
for name, (before, after) in mutants.items():
 assert before in original, name
 try:
  source.write_text(original.replace(before, after, 1))
  log = Path('/tmp/element-access-' + label + '-mutant-' + name + '.log')
  with log.open('w') as output:
   result = subprocess.run(['go', 'test', './internal/oracle', '-run', 'TestNativeAgreesWithNode/internal/oracle/testdata/' + fixture, '-count=1', '-timeout', '10m'], stdout=output, stderr=subprocess.STDOUT, env=__import__('os').environ | {'ADAMIC_GATE_UNCACHED':'1'})
  contents = log.read_text()
  assert result.returncode != 0 and '--- FAIL:' in contents, (name,contents)
  assert 'build failed' not in contents and 'C compile:' not in contents and 'clang failed' not in contents, (name,contents)
  print(name, 'caught', log)
 finally:
  source.write_text(original)
