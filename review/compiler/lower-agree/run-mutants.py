"""Run one mutation at a time and restore the exact input file even on failure."""
import difflib
import json
from pathlib import Path
import subprocess
import time

root = Path('review/compiler/lower-agree')
helper = Path('internal/lower/agree_test.go')
mutants = [
 ('skip-stdout', 'TestAgreementRejectsWrongLoweredOutput', helper, None, ('if !bytes.Equal(got.stdout, want.stdout) {', 'if false {')),
 ('accept-empty', 'TestAgreementRejectsEmptyAnswer', helper, None, ('if len(want.stdout) == 0 {', 'if false {')),
 ('enum-plus-one', 'TestAgreementEnumConstantWitness', Path('internal/lower/enums.go'), ('enum_flags','M01'), None),
 ('drop-base-fields', 'TestAgreementInheritedFieldWitness', Path('internal/lower/class.go'), ('class_inheritance','M02'), None),
 ('drop-parameter-store', 'TestAgreementParameterPropertyWitness', Path('internal/lower/parameter_properties.go'), ('parameter_properties','M03'), None),
 ('drop-static-initializer', 'TestAgreementStaticInitializerWitness', Path('internal/lower/class_static.go'), None, ('statements = append(statements, ir.Evaluate{Value: ir.Call{Function: index, Arguments: []ir.Expression{object}}})', '// Mutant drops the static initializer call.')),
 ('swap-parseInt-radix', 'TestAgreementParseIntRadixWitness', Path('internal/lower/library_method_values.go'), ('library_language','M20'), None),
]
results = []
for name, test, path, evidence, replacement in mutants:
 original = path.read_bytes()
 try:
  if evidence:
   area, identifier = evidence
   ref = 'origin/test-audit/internal-lower-' + area
   location = 'review/test-audit/internal-lower-' + area + '/' + identifier + '.diff'
   patch = subprocess.check_output(['git', 'show', ref + ':' + location])
   (root / (name + '.diff')).write_bytes(patch)
   subprocess.run(['git', 'apply', str(root / (name + '.diff'))], check=True, timeout=10)
  else:
   old, new = replacement
   text = original.decode()
   assert text.count(old) == 1, (name, text.count(old))
   mutated = text.replace(old, new)
   path.write_text(mutated)
   patch = ''.join(difflib.unified_diff(text.splitlines(True), mutated.splitlines(True), fromfile='a/'+str(path), tofile='b/'+str(path)))
   (root / (name + '.diff')).write_text(patch)
  started = time.monotonic()
  command = ['go', 'test', './internal/lower', '-run', '^'+test+'$', '-count=1', '-v', '-timeout', '90s']
  with (root / (name + '.log')).open('w') as log:
   ran = subprocess.run(command, stdout=log, stderr=subprocess.STDOUT, timeout=120)
  output = (root / (name + '.log')).read_text()
  assert ran.returncode == 1 and ('--- FAIL: '+test) in output, (name, output)
  results.append({'mutant':name, 'test':test, 'exit':ran.returncode, 'wall_seconds':round(time.monotonic()-started,3), 'command':command})
  print(name, 'caught by', test, flush=True)
 finally:
  path.write_bytes(original)
(root / 'mutants.json').write_text(json.dumps(results, indent=2)+'\n')
