#!/usr/bin/env python3
"""Remove each type view and plant a runtime edit to prove both guards fail."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess

p = argparse.ArgumentParser()
p.add_argument('tree', type=Path, help='driver slice with adaptations 60-63 applied')
p.add_argument('probe', type=Path, help='feature-integrated checker/lowering probe')
p.add_argument('compiler_cwd', type=Path, help='scratch compiler root containing official Node types')
p.add_argument('output', type=Path, help='new evidence directory')
a = p.parse_args()
a.output.mkdir(exist_ok=False)
repo = Path(__file__).resolve().parents[3]
env = dict(os.environ)
env['CENSUS_TYPESCRIPT'] = env.get('CENSUS_TYPESCRIPT', str(Path.home() / '.cache/adamic-stage3/api/node_modules/typescript/lib/typescript.js'))
entry = a.tree.resolve() / 'parser-host-control.a'

def probe(stem):
    result = (a.output / (stem + '.json')).resolve()
    with (a.output / (stem + '.log')).open('wb') as log:
        subprocess.run([str(a.probe.resolve()), str(entry), str(result)], cwd=a.compiler_cwd, stdout=log, stderr=subprocess.STDOUT, check=True)
    return json.loads(result.read_text())

control = probe('control')
assert control['outcome'] == 'Checker' and len(control['diagnostics']) == 1
assert '/src/compiler/core.ts:' in control['diagnostics'][0]
rows = []
for directory in sorted((repo / 'stage3/adapt').glob('6[0-3]-temporary-*')):
    script = (directory / 'adapt.cjs').read_text()
    fields = {key: json.loads(re.search(r'^const ' + key + r' = (.*);$', script, re.M)[1]) for key in ['file', 'owner', 'before', 'after']}
    target = a.tree / fields['file']
    original = target.read_bytes()
    before, after = (fields[k].encode() for k in ['before', 'after'])
    assert original.count(after) == 1
    try:
        target.write_bytes(original.replace(after, before))
        removed = probe(directory.name + '-removed')
        assert removed['outcome'] == 'Checker' and len(removed['diagnostics']) == 2
        assert any(d.startswith(str(a.tree.resolve() / fields['file']) + ':') for d in removed['diagnostics'])
        # On the same unadapted site, plant a runtime edit into the adapter.
        if fields['owner'] == 'createBaseDeclaration':
            mutated_after = fields['after'].replace('= undefined', '= null')
        elif fields['owner'] == 'createJSDocTypeLikeTagWorker':
            mutated_after = fields['after'].replace('= typeExpression', '= undefined')
        else:
            mutated_after = fields['after'].replace('JSON.stringify', 'String')
        assert mutated_after != fields['after']
        mutant = a.output / (directory.name + '-runtime-mutant.cjs')
        mutant.write_text(script.replace('const after = ' + json.dumps(fields['after']) + ';', 'const after = ' + json.dumps(mutated_after) + ';'))
        logpath = a.output / (directory.name + '-runtime-mutant.log')
        with logpath.open('wb') as log:
            code = subprocess.run(['node', str(mutant.resolve()), str(a.tree.resolve())], env=env, stdout=log, stderr=subprocess.STDOUT).returncode
        assert code != 0 and 'runtime JavaScript changed' in logpath.read_text()
        assert target.read_bytes() == original.replace(after, before), 'failed adapter wrote source'
        rows.append(dict(adaptation=directory.name, removed_view_diagnostics=removed['diagnostics'], runtime_mutant_exit=code, caught_by='adapter emitted-JavaScript equality assertion'))
    finally:
        target.write_bytes(original)
    assert hashlib.sha256(target.read_bytes()).digest() == hashlib.sha256(original).digest()
assert len(rows) == 4
assert probe('restored') == control
(a.output / 'report.json').write_text(json.dumps(dict(control=control, mutants=rows, restored=True), indent=2) + '\n')
print('Four removed-view mutants and four runtime-edit mutants caught; source restored.')
