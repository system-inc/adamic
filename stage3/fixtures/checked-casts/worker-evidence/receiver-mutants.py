"""Omit the new JavaScript receiver guards, independently, with source overlays."""
from pathlib import Path
import json
import subprocess
import sys
root = Path(__file__).resolve().parents[4]
scratch = Path(sys.argv[1]); scratch.mkdir(parents=True, exist_ok=True)
original = root / 'internal/javascript/view_arrays.go'
source = original.read_text()
mutants = [
 ('array-index-root', 'if (!Array.isArray(array)) panic("element read failed:', 'if (false && !Array.isArray(array)) panic("element read failed:', 'TestStep09NullableArrayReceiver', 'receiver check:'),
 ('scalar-index-root', 'if (!Array.isArray(a)) panic(%s); return a[i];', 'if (false && !Array.isArray(a)) panic(%s); return a[i];', 'TestStep09ScalarIndexRootLiar', 'snapshot receiver check:'),
]
for name, before, after, test, failure in mutants:
 assert source.count(before) == 1
 altered = scratch / (name + '.go'); altered.write_text(source.replace(before, after))
 overlay = scratch / (name + '.json'); overlay.write_text(json.dumps({'Replace': {str(original): str(altered)}}))
 log = scratch / (name + '.log')
 with log.open('w') as output:
  result = subprocess.run(['go', 'test', '-overlay=' + str(overlay), './internal/oracle', '-run', '^' + test + '$', '-count=1', '-v'], cwd=root, stdout=output, stderr=subprocess.STDOUT)
 text = log.read_text()
 assert result.returncode != 0 and failure in text and 'TypeError' in text, (name, text)
 assert 'build failed' not in text and 'compiler bug' not in text
 print(name + ': caught by exact checked receiver stop pin')
