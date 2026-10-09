"""Prove the fixture snapshot rejects a wrong exact NotYet reason."""
import json
from pathlib import Path
import subprocess
import tempfile

repo = Path(__file__).resolve().parents[3]
relative = 'internal/oracle/step20_iteration_test.go'
original = repo / relative
source = original.read_text()
needle = '\tfor index, probe := range probes {\n'
assert source.count(needle) == 1
source = source.replace(needle, needle + '\t\tif probe.File == "accessor_binding.a" { probe.Reason = "for...of over an object" }\n', 1)
with tempfile.TemporaryDirectory(prefix='step20-snapshot-') as scratch:
    scratch = Path(scratch)
    replacement = scratch / 'mutant.go'
    replacement.write_text(source)
    overlay = scratch / 'overlay.json'
    overlay.write_text(json.dumps({'Replace': {str(original): str(replacement)}}))
    log = repo / 'stage3/fixtures/iteration/snapshot-mutant.log'
    with log.open('w') as stream:
        result = subprocess.run(['go', 'test', '-overlay=' + str(overlay), './internal/oracle',
                                 '-run', '^TestStep20IterationOutcomes$/accessor_binding.a$',
                                 '-count=1', '-v'], cwd=repo, stdout=stream, stderr=subprocess.STDOUT)
    assert result.returncode != 0, 'snapshot mutant survived'
    assert 'outcome changed: want NotYet "for...of over an object", got NotYet "a for...of destructuring an object"' in log.read_text(), 'wrong failure caught mutant'
print('wrong exact baseline reason killed by snapshot contract')
