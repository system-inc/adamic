"""Build an overlay mutation that drops the deferred read; run the unchanged pin."""
import json
from pathlib import Path
import sys

repository, original, output = map(lambda value: Path(value).resolve(), sys.argv[1:])
output.mkdir(parents=True, exist_ok=True)
metadata = json.loads((original / 'overlay.json').read_text())
source = repository / 'stage3/census/latent/deferred.go'
text = source.read_text()
assert text.count('return deferred') == 1
mutant = output / 'deferred.go'
mutant.write_text(text.replace('return deferred', 'return nil'))
metadata['Replace'][str(source)] = str(mutant)
(output / 'overlay.json').write_text(json.dumps(metadata, indent=2) + '\n')
