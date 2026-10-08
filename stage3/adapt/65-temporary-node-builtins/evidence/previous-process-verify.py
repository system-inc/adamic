#!/usr/bin/env python3
"""Verify erased output, version/drift guards and the real Node-loader contract."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess

p = argparse.ArgumentParser()
p.add_argument('before', type=Path)
p.add_argument('after', type=Path)
p.add_argument('compiler', type=Path)
p.add_argument('compiler_root', type=Path)
p.add_argument('output', type=Path)
a = p.parse_args()
a.output.mkdir(exist_ok=False)
here = Path(__file__).resolve().parent
env = dict(os.environ)
assert env.get('CENSUS_TYPESCRIPT') and env.get('CENSUS_NODE_TYPES')
file = Path('src/compiler/core.ts')
original, adapted = (a.before / file).read_bytes(), (a.after / file).read_bytes()
report = {'before_sha256': hashlib.sha256(original).hexdigest(),
          'after_sha256': hashlib.sha256(adapted).hexdigest()}

def run(command, stem, cwd=None):
    with (a.output / (stem + '.stdout')).open('wb') as out, (a.output / (stem + '.stderr')).open('wb') as err:
        code = subprocess.run(command, cwd=cwd, env=env, stdout=out, stderr=err).returncode
    return code

report['idempotence_exit'] = run(['node', str(here / 'adapt.cjs'), str(a.after)], 'idempotence')
assert report['idempotence_exit'] == 0 and (a.after / file).read_bytes() == adapted
# Plant a changed runtime truth test into the adapter, not a type-only mutation.
mutant = a.output / 'runtime-mutant.cjs'
script = (here / 'adapt.cjs').read_text()
needle = 'changed = changed.replace(oldDeclaration, newDeclaration);'
assert script.count(needle) == 1
mutant.write_text(script.replace(needle, needle + '\nchanged = changed.replace("&& !process.browser", "&& !!process.browser");'))
copy = a.output / 'runtime-mutant-tree'
(copy / file.parent).mkdir(parents=True)
(copy / file).write_bytes(original)
report['runtime_mutant_exit'] = run(['node', str(mutant), str(copy)], 'runtime-mutant')
assert report['runtime_mutant_exit'] != 0
assert 'runtime JavaScript changed' in (a.output / 'runtime-mutant.stderr').read_text()
assert (copy / file).read_bytes() == original, 'failed adapter wrote source'
report['runtime_mutant_caught_by'] = 'emitted JavaScript equality before writing'
# The isolated checker probe activates the same official loader as the parser imports.
report['loader_types_exit'] = run([str(a.compiler), 'types', str(here / 'host-types.a')], 'loader-types', a.compiler_root)
assert report['loader_types_exit'] == 0
wrong = a.output / 'wrong-next-tick.a'
wrong.write_text((here / 'host-types.a').read_text() + '\nif (process !== undefined) { const wrong: number = process.nextTick; console.log(`${wrong}`); }\n')
report['wrong_node_type_exit'] = run([str(a.compiler), 'types', str(wrong)], 'wrong-node-type', a.compiler_root)
assert report['wrong_node_type_exit'] != 0
wrong_diagnostic = (a.output / 'wrong-node-type.stderr').read_text()
assert 'TS2322' in wrong_diagnostic and 'number' in wrong_diagnostic
report['wrong_node_type_diagnostic'] = wrong_diagnostic
# Wrong package metadata must fail before an otherwise idempotent source write.
wrong_package = a.output / 'wrong-version-node'
wrong_package.mkdir()
(wrong_package / 'package.json').write_text(json.dumps({'name': '@types/node', 'version': '24.0.0'}))
node_types = env['CENSUS_NODE_TYPES']
try:
    env['CENSUS_NODE_TYPES'] = str(wrong_package.resolve())
    report['wrong_node_version_exit'] = run(['node', str(here / 'adapt.cjs'), str(a.after)], 'wrong-version')
finally:
    env['CENSUS_NODE_TYPES'] = node_types
assert report['wrong_node_version_exit'] != 0
assert 'exact Node bindings' in (a.output / 'wrong-version.stderr').read_text()
assert (a.after / file).read_bytes() == adapted
report['wrong_node_version_caught_by'] = 'adapter exact-version guard, before source writes'
report['node_types_version'] = json.loads((Path(env['CENSUS_NODE_TYPES']) / 'package.json').read_text())['version']
assert report['node_types_version'] == '25.3.3'
(a.output / 'report.json').write_text(json.dumps(report, indent=2) + '\n')
print(json.dumps(report, indent=2))
