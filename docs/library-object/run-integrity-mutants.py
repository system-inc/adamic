"""Behavioral mutants must fail only the Node comparison; restore the exact bytes after each."""
import os
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[2]
runtime = 'internal/native/runtime/object_integrity.c'
mutants = [
 ('own-names-misses-last-unit', 'internal/native/runtime/object_names.c', 'index < length', 'index + 1 < length', 'keys'),
 ('seal-keeps-configurable', runtime, 'if (sealed) object->sealed = true;', 'if (sealed) object->sealed = false;', 'freeze'),
 ('prevent-keeps-extensible', runtime, 'object->nonextensible = true;', 'object->nonextensible = false;', 'freeze'),
 ('integrity-skips-last-field', runtime, 'index < object->shape->count', 'index + 1 < object->shape->count', 'freeze'),
 ('spread-keeps-integrity', 'internal/native/reuse.go', 'held := fmt.Sprintf("(%s && !%s->frozen && !%s->nonextensible)", uniquelyHeld(source), source, source)', 'held := fmt.Sprintf("(%s && !%s->frozen)", uniquelyHeld(source), source)', 'freeze'),
 ('null-equals-undefined', 'internal/lower/library_object.go', 'result := ir.BooleanConstant{Value: null[0] && null[1]}', 'result := ir.BooleanConstant{Value: null[0] || null[1]}', 'same'),
]
logs = Path(os.environ.get('OBJECT_INTEGRITY_MUTANT_LOGS', '/tmp/object-integrity-mutants'))
logs.mkdir(parents=True, exist_ok=True)
for name, relative, before, after, fixture in mutants:
 path = root / relative
 original = path.read_bytes()
 assert original.decode().count(before) == 1, name
 try:
  path.write_text(original.decode().replace(before, after, 1))
  with (logs / (name + '.log')).open('w') as output:
   result = subprocess.run(['go', 'test', '-count=1', '-timeout', '5m', './internal/oracle', '-run', 'TestNativeAgreesWithNode/internal/oracle/testdata/library_object_' + fixture + '.a$'], cwd=root, env={**os.environ, 'ADAMIC_GATE_UNCACHED':'1'}, stdout=output, stderr=subprocess.STDOUT)
  text = (logs / (name + '.log')).read_text()
  caught = result.returncode != 0 and 'stdout differs' in text and 'Lower:' not in text and 'error:' not in text and 'AddressSanitizer' not in text and 'runtime error:' not in text
  print(name + ': ' + ('caught by Node stdout comparison' if caught else 'NOT PROVEN') + ', exit=' + str(result.returncode), flush=True)
  if not caught:
   print(text, flush=True)
   raise SystemExit(1)
 finally:
  path.write_bytes(original)
