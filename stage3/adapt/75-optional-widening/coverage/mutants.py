#!/usr/bin/env python3
"""Run actual counter, binding, map-schema and reproducer mutations in scratch."""
import argparse
import json
import os
from pathlib import Path
import subprocess
import tempfile

parser = argparse.ArgumentParser()
parser.add_argument('control', type=Path)
parser.add_argument('built_tree', type=Path)
parser.add_argument('driver_maps', type=Path)
parser.add_argument('checker', type=Path)
parser.add_argument('probe', type=Path)
parser.add_argument('output', type=Path)
args = parser.parse_args()
root = Path(__file__).resolve().parent
args.output.mkdir(parents=True, exist_ok=False)
manifest = json.loads((root / 'sites.json').read_text())
results = []

def run(name, command, expect_failure=False, cwd=None, env=None):
    with (args.output / (name + '.log')).open('w') as log:
        process = subprocess.run(command, stdout=log, stderr=subprocess.STDOUT, cwd=cwd, env=env)
    if expect_failure:
        assert process.returncode != 0, f'{name}: mutant survived'
    else:
        assert process.returncode == 0, f'{name}: control failed'
    results.append({'mutant': name, 'exit': process.returncode, 'caught': expect_failure})
    return process.returncode

with tempfile.TemporaryDirectory(prefix='optional-coverage-mutants-') as directory:
    scratch = Path(directory)
    for name, data in [('drop-site', manifest[:-1]), ('wrong-context', [dict(manifest[0], signature=dict(manifest[0]['signature'], contexts=['0'*64]*4)), *manifest[1:]])]:
        file = scratch / (name + '.json')
        file.write_text(json.dumps(data))
        run(name, ['node', str(root / 'instrument.cjs'), 'instrument', str(args.control), str(file)], True)
    maps = sorted(args.driver_maps.glob('*.json'))
    assert len(maps) == 301
    for name in ['missing-exit-map', 'missing-counter-id']:
        directory = scratch / name
        directory.mkdir()
        for file in maps[1:]:
            (directory / file.name).symlink_to(file.resolve())
        if name == 'missing-counter-id':
            data = json.loads(maps[0].read_text())
            del data[manifest[0]['id']]
            (directory / maps[0].name).write_text(json.dumps(data))
        run(name, ['python3', str(root / 'aggregate.py'), str(root / 'sites.json'), str(scratch / 'out.json'), f'driver:301:{directory}'], True)
    # Mutate emitted JavaScript only. TypeScript outputs must remain identical.
    built = args.built_tree / 'built/local'
    mutant = built / '_tsc-no-counter.js'
    script = '''const fs=require('fs'),ts=require('typescript');const [input,output]=process.argv.slice(1);const text=fs.readFileSync(input,'utf8');const f=ts.createSourceFile(input,text,99,true,ts.ScriptKind.JS);let found;function v(n){if(ts.isFunctionDeclaration(n)&&n.name?.text==='optionalWideningCoverageHit')found=n;ts.forEachChild(n,v);}v(f);if(!found)throw Error('missing counter');fs.writeFileSync(output,text.slice(0,found.body.getStart(f))+'{}'+text.slice(found.body.end));'''
    run('make-no-counter', ['node', '-e', script, str(built / '_tsc.js'), str(mutant)])
    try:
        control_maps = scratch / 'control-maps'; control_maps.mkdir()
        mutant_maps = scratch / 'mutant-maps'; mutant_maps.mkdir()
        project = args.built_tree / 'src/compiler/tsconfig.json'
        for name, cli, directory in [('counter-control', built / 'tsc.js', control_maps), ('counter-disabled', mutant, mutant_maps)]:
            env = dict(os.environ, ADAMIC_COVERAGE_DIR=str(directory))
            run(name, ['node', str(cli), '--project', str(project), '--noEmit', '--pretty', 'false', '--tsBuildInfoFile', str(scratch / (name + '.tsbuildinfo'))], env=env)
        def total(directory):
            files = list(directory.glob('*.json')); assert len(files) == 1
            data = json.loads(files[0].read_text()); assert set(data) == {r['id'] for r in manifest}
            return sum(data.values())
        positive, negative = total(control_maps), total(mutant_maps)
        assert positive > 0 and negative == 0
        assert (args.output / 'counter-control.log').read_bytes() == (args.output / 'counter-disabled.log').read_bytes()
        results.append({'mutant':'counter-increment-removed','caught':True,'control_total':positive,'mutant_total':negative,'same_tsc_output':True})
    finally:
        mutant.unlink()
    # Removing the contradictory arm eliminates the census's printed-never component.
    folder = scratch / 'probe'; (folder / 'src/compiler').mkdir(parents=True)
    original = (args.probe / 'src/compiler/probe.ts').read_text()
    (folder / 'src/compiler/probe.ts').write_text(original.replace('(A | { kind: 1 }) & A', 'A'))
    (folder / 'adamic.json').write_bytes((args.probe / 'adamic.json').read_bytes())
    (folder / 'sites.json').write_bytes((args.probe / 'sites.json').read_bytes())
    run('remove-contradictory-union-arm', [str(args.checker),str(folder/'adamic.json'),str(folder/'sites.json'),str(folder/'result.json')])
    original_result = json.loads((args.probe / 'adamic-result.json').read_text())[0]
    mutant_result = json.loads((folder / 'result.json').read_text())[0]
    assert original_result['relation']['source'] == 'never' and mutant_result['relation'] is None
    results.append({'mutant':'contradictory-arm-removed','caught':True,'control_relation':'never','mutant_relation':None})
report = {'checks':[r for r in results if r.get('caught')], 'commands':results}
assert len(report['checks']) == 6
(args.output / 'results.json').write_text(json.dumps(report, indent=2) + '\n')
print(json.dumps(report, indent=2))
