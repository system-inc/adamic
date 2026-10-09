"""Pinned input declarations and test libraries, independent of the tested binary."""
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).parent / 'resources'


def prepare(tree, output):
    manifest = json.loads((ROOT / 'manifest.json').read_text())
    from run import PIN
    if manifest['pin'] != PIN:
        raise RuntimeError('namespace resource pin mismatch')
    for name, expected in manifest['test_libs'].items():
        if hashlib.sha256((tree / 'tests/lib' / name).read_bytes()).hexdigest() != expected:
            raise RuntimeError('namespace test library hash mismatch: ' + name)
    declarations = output / '.namespace-types'
    declarations.mkdir()
    for row in manifest['declarations']:
        raw = (ROOT / row['stored']).read_bytes()
        if hashlib.sha256(raw).hexdigest() != row['sha256']:
            raise RuntimeError('namespace declaration hash mismatch: ' + row['name'])
        (declarations / row['name']).write_bytes(raw)
    return declarations
