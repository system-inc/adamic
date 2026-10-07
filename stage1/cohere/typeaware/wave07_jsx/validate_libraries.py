#!/usr/bin/env python3
"""External-byte oracle and a valid data mutant for owned bundled libraries."""
import hashlib
import json
import pathlib
import subprocess
UNIT = pathlib.Path(__file__).resolve().parent
ROOT = UNIT.parents[3]
OUT = pathlib.Path('/workspace/wave-07-jsx/library-proof')
OUT.mkdir(parents=True, exist_ok=True)
STAGE0 = pathlib.Path('/workspace/wave-07-next-rest/adamic')
pins = json.loads((UNIT/'library-pins.json').read_text())
expected = b''
for name, digest in pins['sha256'].items():
    data = (ROOT/'cohere/TypeScript/tsc/internal/bundled/libs'/name).read_bytes()
    assert hashlib.sha256(data).hexdigest() == digest
    expected += data+b'\n'
records = []
def run(name, command):
    with (OUT/(name+'.stdout')).open('wb') as stdout, (OUT/(name+'.stderr')).open('wb') as stderr:
        result = subprocess.run([str(value) for value in command], stdout=stdout, stderr=stderr)
    records.append(dict(name=name, command=[str(value) for value in command], exit=result.returncode))
    (OUT/'commands.json').write_text(json.dumps(records, indent=2)+'\n')
    assert result.returncode == 0
    return (OUT/(name+'.stdout')).read_bytes(), (OUT/(name+'.stderr')).read_bytes()
run('sanitized-build', [STAGE0, 'build', UNIT/'libraries_probe.a', '-o', OUT/'native', '--sanitize'])
actual, error = run('sanitized', [OUT/'native'])
assert actual == expected and not error
(OUT/'libraries_probe.a').write_bytes((UNIT/'libraries_probe.a').read_bytes())
source = (UNIT/'libraries.a').read_text()
assert 'Copyright (c)' in source
(OUT/'libraries.a').write_text(source.replace('Copyright (c)', 'CopyrigXt (c)', 1))
run('data-mutant-build', [STAGE0, 'build', OUT/'libraries_probe.a', '-o', OUT/'mutant'])
changed, error = run('data-mutant', [OUT/'mutant'])
assert changed != expected and not error
print('PASS: 114 libraries, '+str(len(expected))+' exact external bytes, sanitizers; valid data mutant exits 0 and is caught only by byte comparison.')
