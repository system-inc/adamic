#!/usr/bin/env python3
"""Check repair serialization in the unmodified published next harness."""
import argparse
import json
from pathlib import Path
import subprocess

owned = Path(__file__).resolve().parent
repository = owned.parents[4]
p = argparse.ArgumentParser()
p.add_argument('--scratch', type=Path, required=True)
args = p.parse_args()
scratch = args.scratch.resolve()
scratch.mkdir(parents=True, exist_ok=True)
pin = '2650ad595b82220c368631ea13139fad4b306ed6'
source = subprocess.check_output(['git', 'show', pin + ':stage1/cohere/lint/testdata/oracle.go'], cwd=repository).decode()
assert source.count('func main() {') == 1
(scratch / 'serializer.go').write_text(source.replace('func main() {', 'func serializerMain() {'))
virtual = repository / 'cohere/wave15_next_serializer.go'
probe = repository / 'cohere/wave15_next_entry.go'
overlay = scratch / 'overlay.json'
overlay.write_text(json.dumps({'Replace': {str(virtual): str(scratch / 'serializer.go'), str(probe): str(owned / 'probe.go.txt')}}))
with (scratch / 'build.log').open('w') as log:
    subprocess.run(['go', 'build', '-overlay=' + str(overlay), '-o', str(scratch / 'oracle'), str(virtual), str(probe)], cwd=repository / 'cohere', stdout=log, stderr=subprocess.STDOUT, check=True)
for slug, name, expected in [
    ('typescript-no-unnecessary-type-constraint', '@typescript-eslint/no-unnecessary-type-constraint', 0),
    ('typescript-prefer-as-const', '@typescript-eslint/prefer-as-const', 2),
    ('typescript-prefer-enum-initializers', '@typescript-eslint/prefer-enum-initializers', 0),
]:
    with (scratch / (slug + '.log')).open('wb') as log:
        result = subprocess.run([str(scratch / 'oracle'), name, str(repository / 'stage1/cohere/lint/rules' / slug / 'testdata/blocked.ts.txt')], stdout=log, stderr=subprocess.PIPE)
    (scratch / (slug + '.stderr')).write_bytes(result.stderr)
    assert result.returncode == expected, (name, result.returncode, result.stderr)
    if expected == 2:
        assert b'panic: unexpected fix shape' in result.stderr
    else:
        assert not result.stderr
    print(name + ': next-harness serializer exit=' + str(result.returncode) + (' unexpected fix shape' if expected else ' complete proposals serialized'))
