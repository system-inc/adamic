#!/usr/bin/env python3
"""Run each behavioral mutant against the source-on-Node oracle, restoring every edit."""
import os
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[2]
runtime = 'internal/native/runtime/library_object.c'
mutants = [
 ('keys-integer-last', runtime, 'return ai ? -1 : 1;', 'return ai ? 1 : -1;', 'keys'),
 ('keys-descending', runtime, 'return a < b ? -1 : a > b ? 1 : 0;', 'return a < b ? 1 : a > b ? -1 : 0;', 'keys'),
 ('values-first-slot', runtime, 'adamic_value value = object->slots[index];', 'adamic_value value = object->slots[indices[0]];', 'keys'),
 ('entries-wrong-key', runtime, 'pair->slots[0].reference = key_string(name);', 'pair->slots[0].reference = key_string("wrong");', 'keys'),
 ('assign-a-plus-one', runtime, '*slot = value;', 'if (strcmp(name, "a") == 0) value.number += 1; *slot = value;', 'assign'),
 ('same-value-nan-false', runtime, 'if (isnan(a) && isnan(b)) return true;', 'if (isnan(a) && isnan(b)) return false;', 'is'),
 ('same-value-zero', runtime, 'return a == b && (a != 0 || signbit(a) == signbit(b));', 'return a == b;', 'is'),
 ('has-own-always-false', runtime, 'memcmp(name, key->bytes, key->length) == 0) return true;', 'memcmp(name, key->bytes, key->length) == 0) return false;', 'has_own'),
 ('freeze-no-flag', runtime, 'object->frozen = true;', 'object->frozen = false;', 'freeze'),
 ('freeze-write-ignored', runtime, "if (!object->frozen || name[0] == '#') return;", "if (object->shape != NULL) return;", 'freeze_write'),
 ('spread-reuses-frozen', 'internal/native/reuse.go', 'held := fmt.Sprintf("(%s && !%s->frozen && !%s->nonextensible)", uniquelyHeld(source), source, source)', 'held := uniquelyHeld(source)', 'freeze'),
]
log_dir = Path(os.environ.get('OBJECT_MUTANT_LOGS', '/tmp/object-mutants'))
log_dir.mkdir(parents=True, exist_ok=True)
for name, relative, before, after, fixture in mutants:
 if os.environ.get('OBJECT_MUTANT_ONLY') and name != os.environ['OBJECT_MUTANT_ONLY']:
  continue
 path = root / relative
 original = path.read_text()
 assert original.count(before) == 1, name
 try:
  path.write_text(original.replace(before, after, 1))
  log = log_dir / (name + '.log')
  with log.open('w') as output:
   result = subprocess.run(['go', 'test', '-count=1', '-timeout', '5m', './internal/oracle', '-run', 'TestNativeAgreesWithNode/internal/oracle/testdata/library_object_' + fixture + '.a$'], cwd=root, stdout=output, stderr=subprocess.STDOUT)
  text = log.read_text()
  caught = result.returncode != 0 and ('stdout differs' in text or 'exit codes differ' in text) and 'error:' not in text and 'AddressSanitizer' not in text and 'runtime error:' not in text
  print(name + ': ' + ('caught by Node comparison' if caught else 'NOT PROVEN') + ', exit=' + str(result.returncode), flush=True)
  if not caught:
   print(text, flush=True)
   raise SystemExit(1)
 finally:
  path.write_text(original)
