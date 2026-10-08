#!/usr/bin/env python3
"""Reproduce coverage in an uncommitted tree with upstream's build and runners."""
import argparse
import json
import os
from pathlib import Path
import shutil
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument('output', type=Path, help='new scratch directory, outside the repository')
args = parser.parse_args()
unit = Path(__file__).resolve().parent
repository = unit.parents[3]
out = args.output.resolve()
assert not out.exists(), 'refusing to replace an existing output'
assert repository not in out.parents, 'TypeScript scratch source must stay outside the repository'
out.mkdir(parents=True)
env = dict(os.environ)
env.setdefault('NODE_PATH', str(Path.home() / '.cache/adamic-stage3/api/node_modules'))

def run(name, command, extra=None, cwd=repository):
    with (out / (name + '.log')).open('w') as log:
        process = subprocess.run(command, cwd=cwd, stdout=log, stderr=subprocess.STDOUT, env=env | (extra or {}))
    if process.returncode:
        raise RuntimeError(f'{name} exited {process.returncode}; see {out / (name + ".log")}')

scoreboard = repository / 'stage3/patch-set.md'
original = scoreboard.read_bytes()
try:
    run('apply', ['bash', 'stage3/apply.sh', str(out / 'tree')])
    shutil.copyfile(scoreboard, out / 'patch-set.md')
finally:
    scoreboard.write_bytes(original)
run('instrument', ['node', str(unit / 'instrument.cjs'), 'instrument', str(out / 'tree'), str(unit / 'sites.json')])
for name in ['upstream', 'driver', 'self']:
    (out / (name + '-counts')).mkdir()
run('oracle', ['bash', 'stage3/oracle/run.sh', str(out / 'tree'), str(out / 'oracle'), '--runners=compiler,conformance'], {'ADAMIC_COVERAGE_DIR':str(out / 'upstream-counts')})
report = json.loads((out / 'oracle/report.json').read_text())
assert report['status'] == 'pass' and report['counts']['passing'] > 0 and report['baseline_diffs'] == []
run('driver', ['bash', 'stage3/drivers/tsc/run.sh', '--', 'node', str(out / 'tree/built/local/tsc.js')], {'ADAMIC_COVERAGE_DIR':str(out / 'driver-counts'), 'TSC_RESULTS':str(out / 'driver-results')})
assert json.loads((out / 'driver-results/report.json').read_text())['passed'] == 301
run('self', ['node', str(out / 'tree/built/local/tsc.js'), '--project', str(out / 'tree/src/compiler/tsconfig.json'), '--noEmit', '--pretty', 'false', '--tsBuildInfoFile', str(out / 'self.tsbuildinfo')], {'ADAMIC_COVERAGE_DIR':str(out / 'self-counts')})
run('aggregate', ['python3', str(unit / 'aggregate.py'), str(unit / 'sites.json'), str(out / 'counts.json'), f'driver:301:{out / "driver-counts"}', f'upstream:5:{out / "upstream-counts"}', f'self:1:{out / "self-counts"}'])
assert json.loads((out / 'counts.json').read_text())['ran'] == 42
# Reconstruct a control source view for static, read-only checker queries.
shutil.copytree(out / 'tree/src', out / 'control/src')
(out / 'control/node_modules').symlink_to(out / 'tree/node_modules')
run('restore', ['node', str(unit / 'restore.cjs'), str(out / 'tree'), str(out / 'control'), str(unit / 'sites.json')])
config = {'extends':'./src/compiler/tsconfig.json','compilerOptions':{'strict':True,'exactOptionalPropertyTypes':True,'noUncheckedIndexedAccess':True,'verbatimModuleSyntax':True,'erasableSyntaxOnly':False,'strictBindCallApply':True,'useUnknownInCatchVariables':True,'allowImportingTsExtensions':True,'noEmit':True,'module':'ESNext','moduleDetection':'force','moduleResolution':'bundler','target':'ES2024','lib':['es2024'],'types':[]}}
(out / 'control/checker-config.json').write_text(json.dumps(config,indent=2)+'\n')
run('checker-build', ['go', 'build', '-o', str(out / 'checker'), str(unit / 'checker.go'), str(unit / 'relation.go')])
run('checker', [str(out / 'checker'), str(out / 'control/checker-config.json'), str(unit / 'sites.json'), str(out / 'checker.json')])
run('stock', ['node', str(unit / 'stock.cjs'), str(out / 'control/checker-config.json'), str(unit / 'sites.json'), str(out / 'stock.json')])
run('guards', ['node', str(unit / 'guards.cjs'), str(out / 'control'), str(out / 'guards.json')])
print(json.dumps({'output':str(out),'ran':42,'unobserved':5,'oracle':report['counts']}))
