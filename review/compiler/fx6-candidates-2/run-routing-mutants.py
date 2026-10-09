from pathlib import Path
import difflib
import subprocess

root = Path.cwd()
evidence = root / 'review/compiler/fx6-candidates-2'
mutants = [
 ('element-bypass', 'internal/lower/element_access_fields.go',
  'read = l.readViewMember(node, ir.Property{Object: heldObject, Name: name, Of: stored, Absent: fields[index].Flags&ast.SymbolFlagsOptional != 0}, fields[index], receiver)',
  'read = ir.Property{Object: heldObject, Name: name, Of: stored, Absent: fields[index].Flags&ast.SymbolFlagsOptional != 0}',
  'TestCheckedViewElementP05'),
 ('destructure-bypass', 'internal/lower/collections.go',
  'value = l.readViewMember(binding, property, fieldSymbol, destructured)',
  '_ = fieldSymbol; value = property', 'TestCheckedViewDestructuredUnion'),
]
for name, filename, before, after, test in mutants:
 path = root / filename
 original = path.read_text()
 assert original.count(before) == 1
 mutated = original.replace(before, after, 1)
 (evidence / (name+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True), mutated.splitlines(True), fromfile='a/'+filename, tofile='b/'+filename)))
 try:
  path.write_text(mutated)
  with (evidence / (name+'.log')).open('w') as log:
   result = subprocess.run(['go','test','./internal/oracle','-run','^'+test+'$','-count=1','-v','-timeout','90s'],stdout=log,stderr=subprocess.STDOUT,timeout=180)
  output=(evidence/(name+'.log')).read_text()
  print(name, 'exit', result.returncode)
  assert result.returncode != 0 and 'exit codes differ' in output and 'got exit 0' in output, output
  assert 'Sanitizer' not in output, output
 finally:
  path.write_text(original)
