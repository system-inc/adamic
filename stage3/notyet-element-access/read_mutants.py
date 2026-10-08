"""Run semantic finite-read mutants independently and restore each edit."""
from pathlib import Path
import subprocess

source = Path('internal/lower/element_access_fields.go')
original = source.read_text()
mutants = {
 'wrong-dispatch': ('Operator: ir.Equal', 'Operator: ir.NotEqual'),
 'missing-absence': ('Absent: fields[index].Flags&ast.SymbolFlagsOptional != 0', 'Absent: false'),
 'repeated-receiver': ('[]ir.Expression{object, key}', '[]ir.Expression{object, key, object}'),
}
for name, (before, after) in mutants.items():
 assert before in original, name
 try:
  source.write_text(original.replace(before, after, 1))
  log = Path('/tmp/element-access-read-mutant-' + name + '.log')
  with log.open('w') as output:
   result = subprocess.run(['go', 'test', './internal/oracle', '-run', 'TestNativeAgreesWithNode/internal/oracle/testdata/element_access_reads', '-count=1', '-timeout', '10m'], stdout=output, stderr=subprocess.STDOUT, env=__import__('os').environ | {'ADAMIC_GATE_UNCACHED':'1'})
  contents = log.read_text()
  assert result.returncode != 0 and '--- FAIL:' in contents, (name,contents)
  assert 'build failed' not in contents and 'C compile:' not in contents, (name,contents)
  print(name, 'caught', log)
 finally:
  source.write_text(original)
