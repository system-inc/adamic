"""Revert each reported fix independently, run its witness, and always restore."""
from pathlib import Path
import os
import subprocess

root = Path(__file__).resolve().parents[4]
logs = Path('/tmp/lane4b-fix-mutants')
logs.mkdir(exist_ok=True)
cases = [
 ('graph-size', 'internal/native/graph_regions.go', 'fmt.Sprintf("adamic_object_size(%s->shape->count)", value)', 'fmt.Sprintf("sizeof *%s + %s->shape->count * (sizeof(adamic_value) + 2)", value, value)', 'graph', 'AddressSanitizer: heap-buffer-overflow'),
 ('message-size', 'internal/native/runtime/object.c', 'strlen(expression) + 2 * strlen(type) + strlen(found) + 100', 'strlen(expression) + strlen(type) + strlen(found) + 100', 'long-message', 'AddressSanitizer: heap-buffer-overflow'),
 ('tuple-member', 'internal/lower/view_lazy.go', 'if l.objectPrimitiveTupleMember(target) {', 'if false {', 'tuple', 'expected named tuple union refusal'),
]
for name, relative, before, after, witness, catcher in cases:
 path = root / relative
 original = path.read_bytes()
 text = original.decode()
 assert text.count(before) == 1, (name, 'mutation anchor drift')
 try:
  path.write_text(text.replace(before, after, 1))
  env = dict(os.environ, ADAMIC_GATE_UNCACHED='1')
  with (logs / (name+'.log')).open('w') as output:
   result = subprocess.run(['go', 'test', './internal/oracle', '-run', '^TestCheckedViewObjectPrimitiveFixes/'+witness+'$', '-count=1', '-v', '-timeout', '10m'], cwd=root, env=env, stdout=output, stderr=subprocess.STDOUT)
  observed = (logs / (name+'.log')).read_text()
  assert result.returncode != 0 and catcher in observed, (name, result.returncode, observed)
  assert 'error:' not in observed or 'AddressSanitizer' in observed, (name, 'not an executable witness')
  print(name+': caught by '+catcher, flush=True)
 finally:
  path.write_bytes(original)
