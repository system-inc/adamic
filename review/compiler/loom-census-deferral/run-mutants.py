"""Execute each source mutant and require an assertion kill, not a build failure."""
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys

root = Path.cwd()
evidence = root / 'review/compiler/loom-census-deferral'

def command(arguments, log, environment=None, expected=0):
    with log.open('w') as output:
        result = subprocess.run(list(map(str, arguments)), env=environment,
                                stdout=output, stderr=subprocess.STDOUT, timeout=60)
    text = log.read_text()
    if expected == 0:
        assert result.returncode == 0, (arguments, result.returncode, text)
    else:
        assert result.returncode != 0 and 'AssertionError' in text, (arguments, result.returncode, text)
    return text

command(['python3', 'stage3/census/latent/audit_deferred.py', '/tmp/loom-census-no-caller', evidence/'no-caller-audit'],
        evidence/'no-caller-mutant.log', expected=1)
print('no-caller: caught by unused generic instantiated assertion')
command(['python3', 'stage3/census/speculative/audit.py', '/tmp/loom-census-plain-map', evidence/'plain-map-audit'],
        evidence/'plain-map-mutant.log', expected=1)
print('plain-map: caught by speculative depth/continuation assertion')
source = evidence/'compiler-controls'
source.mkdir(exist_ok=True)
for name in ['find', 'first', 'last', 'callback', 'array']:
    shutil.copyfile(root/'internal/oracle/testdata'/('instantiation_'+name+'.a'), source/(name+'.a'))
for name, binary in [('baseline', '/tmp/loom-census'), ('premature', '/tmp/loom-census-premature')]:
    output = evidence/(name+'-compiler-controls.jsonl')
    env = dict(os.environ, LATENT_FULL='1', LATENT_ASSERT_NO_OUTPUT='1')
    command([binary, source, output], evidence/(name+'-compiler-controls.log'), env)
    text = command(['python3', 'review/compiler/instantiation-proof/census-check.py', output],
                   evidence/(name+'-inventory-check.log'), expected=0 if name=='baseline' else 1)
    if name == 'baseline':
        rows = [json.loads(line) for line in output.read_text().splitlines()]
        mutant = json.loads(json.dumps(rows))
        for row in mutant:
            if 'units' in row:
                row['units'] = [unit for unit in row['units'] if unit.get('name') != 'firstOrUndefined']
        dropped = evidence/'missing-inventory.jsonl'
        dropped.write_text(''.join(json.dumps(row)+'\n' for row in mutant))
        command(['python3', 'review/compiler/instantiation-proof/census-check.py', dropped],
                evidence/'missing-inventory-mutant.log', expected=1)
        print('missing inventory: caught by original census-check.py')
    else:
        assert "'status': 'attempted'" in text, text
        print('premature: caught by original census-check.py deferred status assertion')
print('PASS: all four independent census mutants caught by assertions')
