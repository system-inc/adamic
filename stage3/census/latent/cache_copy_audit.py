"""Exercise rollback facts in the guarded overlay and kill two copier mutants."""
import json
import os
from pathlib import Path
import subprocess
import sys

repository, overlay, output = map(lambda p: Path(p).resolve(), sys.argv[1:])
output.mkdir(parents=True, exist_ok=True)
config = json.loads((overlay / 'overlay.json').read_text())
config['Replace'][str(repository / 'internal/lower/latent_cache_test.go')] = str(Path(__file__).with_name('cache_copy_test.go.txt'))
full = Path(config['Replace'][str(repository / 'internal/lower/latent_full.go')])
text = full.read_text()
entry = '{reflect.TypeFor[ir.Program](), "argumentFacts"}: true,'
assert text.count(entry) == 1, 'one explicitly named derived cache'
for name, source, expected in [
    ('baseline', text, None),
    ('removed-cache-entry', text.replace(entry, ''), 'latent state copy: unexported IR field argumentFacts'),
    ('skip-all-private', text.replace('panic("latent state copy: unexported IR field " + v.Type().Field(i).Name)', 'continue'), 'unreviewed private field accepted'),
]:
    selected = dict(config, Replace=dict(config['Replace']))
    copied = output / (name + '.go')
    copied.write_text(source)
    selected['Replace'][str(repository / 'internal/lower/latent_full.go')] = str(copied)
    manifest = output / (name + '-overlay.json')
    manifest.write_text(json.dumps(selected))
    log = output / (name + '.log.txt')
    command = ['go', 'test', '-buildvcs=false', '-overlay=' + str(manifest), './internal/lower',
               '-run', '^TestLatent(CacheRollbackFacts|UnknownPrivateField)$', '-count=1', '-v']
    with log.open('w') as stream:
        run = subprocess.run(command, cwd=repository, env=dict(os.environ, TMPDIR=str(output)), stdout=stream, stderr=stream)
    observed = log.read_text()
    if expected is None:
        assert run.returncode == 0, observed
        print('PASS: actual failed-initializer rollback, subsequent lowering, packed facts equal fresh program, unknown private fields refused')
    else:
        assert run.returncode != 0 and expected in observed and '--- FAIL: TestLatent' in observed, observed
        print('CAUGHT:', name, 'by', expected)
