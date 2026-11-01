#!/usr/bin/env python3
"""Replay source, idempotence, single-repair propagation, and undo controls.

Run after apply.sh has produced AFTER; BEFORE is a separate 95a-only tree.
The census binary is the real stage3/census/tool built by the caller.
"""
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys

before, after, census, output = map(Path, sys.argv[1:])
output.mkdir(parents=True, exist_ok=False)
here = Path(__file__).resolve().parent
sites = json.loads((here / 'sites.json').read_text())
env = dict(os.environ, NODE_PATH=str(here.parents[1] / 'api/node_modules'))
def run(label, command, expected=0):
    with (output / (label + '.log')).open('w') as log:
        result = subprocess.run(list(map(str,command)), stdout=log, stderr=subprocess.STDOUT, env=env)
    assert result.returncode == expected, (label,result.returncode,expected)
    return result.returncode

def observe(label, tree):
    destination = output / (label + '.jsonl')
    run(label,[census,tree / 'src/tsc/tsc.ts',destination])
    return destination

run('source',['node',here / 'verify.cjs',before,after])
run('idempotence',['node',here / 'adapt.cjs',after])
assert json.loads((output / 'idempotence.log').read_text())['declarations'] == 0
run('check',['node',here / 'adapt.cjs','--check',after])
before_census, after_census = observe('before',before), observe('after',after)
# Measure each widening alone, preserving exact bytes and restoring in finally.
original = {relative:(before / relative).read_bytes() for relative in {r['file'] for r in sites}}
final = {relative:(after / relative).read_bytes() for relative in original}
individual = []
try:
    for index,row in enumerate(sites):
        for relative,data in original.items():
            (after / relative).write_bytes(data)
        run('edit-individual-' + str(index+1),['node',here / 'edit-one.cjs',after,index,'forward'])
        individual.append(observe('individual-' + str(index+1),after))
finally:
    for relative,data in final.items():
        (after / relative).write_bytes(data)
# Keep the undo mutant isolated from the oracle input.
mutant = output / 'mutant'
shutil.copytree(after,mutant,ignore=shutil.ignore_patterns('.git','node_modules','built'))
row = sites[0]
run('edit-mutant',['node',here / 'edit-one.cjs',mutant,0,'undo'])
run('mutant-verifier',['node',here / 'verify.cjs',before,mutant],1)
run('mutant-check',['node',here / 'adapt.cjs','--check',mutant],1)
mutant_census = observe('mutant',mutant)
run('census',['python3',here / 'compare-census.py',before_census,after_census,mutant_census])
# Import only the shared meter normalization without invoking the comparison CLI.
namespace = {}
code = (here / 'compare-census.py').read_text().split('before, after, mutant =')[0]
exec(code,namespace)
normalize = namespace['observations']
base = set(normalize(before_census))
propagation = []
for row,record in zip(sites,individual):
    current = set(normalize(record))
    assert not current-base, 'single declaration introduced diagnostics'
    assert len(base-current) == 1, 'each declaration must remove one diagnostic'
    propagation.append(dict(owner=row['owner'],member=row['name'],removed=sorted(base-current),added=[]))
(output / 'propagation.json').write_text(json.dumps(propagation,indent=2)+'\n')
print(json.dumps(dict(status='pass',individualRepairs=len(individual),undoVerifierExit=1,undoCheckExit=1),indent=2))
