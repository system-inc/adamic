"""Replay already-built owned cores only when their compiler/bridge/rule sources are unchanged."""
from pathlib import Path
import hashlib
import json
import subprocess
import sys
import time

repository, artifacts, destination = map(Path, sys.argv[1:])
destination.mkdir(parents=True, exist_ok=True)

def run(label, arguments):
    stdout = destination / (label + '.stdout')
    stderr = destination / (label + '.stderr')
    started = time.monotonic()
    with stdout.open('wb') as out, stderr.open('wb') as err:
        process = subprocess.run([str(a) for a in arguments], cwd=repository, stdout=out, stderr=err)
    assert process.returncode == 0, (label, process.returncode, stderr.read_bytes())
    return stdout.read_bytes(), stderr.read_bytes(), time.monotonic() - started

records = []
controls = {}
for mode, flags in [('default', []), ('element', ['--element']), ('globals', ['--allow-globals'])]:
    for manifest in sorted(artifacts.glob(mode + '-*.manifest')):
        name = manifest.stem
        expected, _, go_seconds = run(name + '-go', [artifacts / 'oracle', artifacts / 'config.json', manifest, *flags])
        for flavor in ['native', 'asan']:
            observed, errors, seconds = run(name + '-' + flavor, [artifacts / (name + '-' + flavor)])
            assert observed == expected and not errors, (name, flavor)
            records.append({'name': name, 'flavor': flavor, 'bytes': len(expected), 'sha256': hashlib.sha256(expected).hexdigest(), 'seconds': seconds, 'go_seconds': go_seconds})
        if mode == 'default':
            controls[name] = expected
for name in ['compiler', 'repository']:
    config = Path('/workspace/wave21-compiler/src/compiler/tsconfig.json') if name == 'compiler' else repository / 'tsconfig.json'
    manifest = Path('/workspace/wave21-' + name + '.manifest')
    expected, _, seconds = run(name + '-go', [artifacts / 'oracle', config, manifest])
    for flavor in ['native', 'asan']:
        observed, errors, native_seconds = run(name + '-' + flavor, [artifacts / (name + '-' + flavor)])
        assert observed == expected and not errors, (name, flavor)
        records.append({'name': name, 'flavor': flavor, 'bytes': len(expected), 'sha256': hashlib.sha256(expected).hexdigest(), 'seconds': native_seconds, 'go_seconds': seconds})
for name, message in [('fragment', 'preferFragment'), ('undef', 'jsxIdentifierNotDefined'), ('context', 'withIdentifierMsg')]:
    expected = next(data for data in controls.values() if ('\t' + message + '\t').encode() in data)
    observed, errors, _ = run(name + '-mutant', [artifacts / (name + '-mutant')])
    assert observed != expected and not errors, name
    first = next((i for i, (a, b) in enumerate(zip(observed, expected)) if a != b), min(len(observed), len(expected)))
    print(name, 'mutant exits 0, empty stderr, independent byte comparison catches byte', first, flush=True)
(destination / 'results.json').write_text(json.dumps(records, indent=2) + '\n')
print('66 batches and both corpora match fresh production Go, normal and sanitized', flush=True)
