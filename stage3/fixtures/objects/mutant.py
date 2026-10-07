"""Plain-field rewrite must change the observed late-spread refusal into a native build."""
import difflib
import json
import subprocess
import tempfile
from pathlib import Path

bucket = Path(__file__).resolve().parent
repository = bucket.parents[2]
fixture = bucket / '03_resolution_cache_spreads.a'
original = fixture.read_text()
mutated = original.replace('        ...packageJsonInfoCache,\n        ...perDirectoryResolutionCache,\n        ...nonRelativeNameResolutionCache,', '        packageCount: packageJsonInfoCache.packageCount,\n        directoryCount: perDirectoryResolutionCache.directoryCount,\n        nameCount: nonRelativeNameResolutionCache.nameCount,')
assert mutated != original
status = next(row for row in json.loads((bucket / 'status.json').read_text()) if row['file'] == fixture.name)
assert status['stage0']['outcome'] == 'Refused'
assert 'a spread after the first field' in status['stage0']['what']
with tempfile.TemporaryDirectory(prefix='objects-mutant-') as directory:
    source = Path(directory) / fixture.name
    source.write_text(mutated)
    binary = Path(directory) / 'plain'
    node = subprocess.run(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(source)], cwd=repository, capture_output=True)
    build = subprocess.run(['go', 'run', './cmd/adamic', 'build', str(source), '-o', str(binary)], cwd=repository, capture_output=True)
    assert build.returncode == 0, build.stderr.decode()
    native = subprocess.run([str(binary)], cwd=repository, capture_output=True)
    assert node.stdout == native.stdout == status['node']['stdout'].encode()
    assert node.stderr == native.stderr == b''
    assert node.returncode == native.returncode == 0
    result = {'file': fixture.name, 'rewrite': 'three spreads to three explicit field reads', 'before': status['stage0'], 'after': {'outcome': 'Compiles', 'what': ''}, 'node': {'stdout': node.stdout.decode(), 'stderr': '', 'exit': 0}, 'native': {'stdout': native.stdout.decode(), 'stderr': '', 'exit': 0}, 'caught_by': 'the recorded Refused outcome changes to Compiles'}
(bucket / 'mutant.json').write_text(json.dumps(result, indent=2) + '\n')
(bucket / 'mutant.patch').write_text(''.join(difflib.unified_diff(original.splitlines(True), mutated.splitlines(True), fromfile=fixture.name, tofile=fixture.name)))
print('Refused -> Compiles; Node and native stdout: ' + repr(node.stdout.decode()))
