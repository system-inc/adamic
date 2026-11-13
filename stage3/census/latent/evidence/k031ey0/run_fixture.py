"""Run one bounded census child, retaining and returning its exact exit code."""
import json
import os
from pathlib import Path
import subprocess
import sys

binary, mode, label = sys.argv[1:]
root = Path(__file__).resolve().parent
env = {k: v for k, v in os.environ.items() if k not in ('LATENT_FULL', 'LATENT_SPECULATIVE')}
env['LATENT_ASSERT_NO_OUTPUT'] = '1'
env['LATENT_TAIL'] = '1'
if mode == 'full': env['LATENT_FULL'] = '1'
if mode == 'speculative': env['LATENT_SPECULATIVE'] = '1'
output = root / (label + '-' + mode + '.jsonl')
with (root / (label + '-' + mode + '.log')).open('w') as log:
    child = subprocess.run(['timeout', '60s', binary, str(root / 'fixture'), str(output)], env=env, stdout=log, stderr=subprocess.STDOUT)
print(f'{label} {mode}: child exit {child.returncode}')
if child.returncode == 0:
    rows = [json.loads(line) for line in output.read_text().splitlines()]
    findings = [f for row in rows for f in row.get('findings', []) if 'a view of a type that reaches' in f.get('reason', '')]
    print(f'new view NotYet findings: {len(findings)}')
    for finding in findings: print(json.dumps(finding, sort_keys=True))
sys.exit(child.returncode)
