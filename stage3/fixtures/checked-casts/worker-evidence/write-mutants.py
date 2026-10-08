"""Omit each backend's real primitive-array write guard through Go overlays."""
from pathlib import Path
import json
import subprocess
import sys

root = Path(__file__).resolve().parents[4]
scratch = Path(sys.argv[1])
scratch.mkdir(parents=True, exist_ok=True)
mutants = [
 ('native-write-kind', 'internal/native/view_array_writes.go',
  'e.arrayPrimitiveUnionRead(ir.ViewContractID(index+1), value, "<array write>", source.Name)',
  'e.line("(void)%s;", value)'),
 ('javascript-write-kind', 'internal/javascript/view_array_writes.go',
  'if (!adamicArrayPrimitiveWrites[target](value)) panic(',
  'if (false && !adamicArrayPrimitiveWrites[target](value)) panic('),
]
for name, relative, before, after in mutants:
 original = root / relative
 source = original.read_text()
 assert source.count(before) == 1, (name, source.count(before))
 altered = scratch / (name + '.go')
 altered.write_text(source.replace(before, after))
 overlay = scratch / (name + '.json')
 overlay.write_text(json.dumps({'Replace': {str(original): str(altered)}}))
 log = scratch / (name + '.log')
 command = ['go', 'test', '-overlay=' + str(overlay), './internal/oracle', '-run', '^TestStep09PrimitiveArraySourceWriteLiar$', '-count=1', '-v']
 with log.open('w') as output:
  result = subprocess.run(command, cwd=root, stdout=output, stderr=subprocess.STDOUT)
 text = log.read_text()
 assert result.returncode != 0 and 'source-write contract:' in text and 'result[0]' in text, (name, result.returncode, text)
 assert 'build failed' not in text and 'compiler bug' not in text, text
 print(name + ': caught by exact source-write stop pin')
