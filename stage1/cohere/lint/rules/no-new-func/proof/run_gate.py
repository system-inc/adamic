"""Run the unchanged lint package with every public input and preserve evidence."""
import collections
import json
import os
from pathlib import Path
import statistics
import subprocess
import time

root = Path.cwd()
proof = root / 'stage1/cohere/lint/rules/no-new-func/proof'
profiles = Path('/workspace/wave-06-facts-profiles')
profiles.mkdir(exist_ok=True)
environment = dict(os.environ)
environment.update({
    'GOMAXPROCS': '4',
    'ADAMIC_LINT_BENCH': '1',
    'ADAMIC_TYPESCRIPT_SOURCE': '/workspace/wave-06-typescript',
    'ADAMIC_LINT_PROFILE_DIR': str(profiles),
    'ADAMIC_LINT_PROFILE_SNAPSHOTS': str(profiles),
})
command = ['go', 'test', '-json', '-count=1', '-timeout=3h', './stage1/cohere/lint']
(proof / 'command.json').write_text(json.dumps({'command': command, 'inputs': {k: environment[k] for k in environment if k.startswith('ADAMIC_LINT_') or k in ['GOMAXPROCS', 'ADAMIC_TYPESCRIPT_SOURCE']}, 'baseline': 'd845dccde413c89643293e808626344d12e3f023'}, indent=2) + '\n')
started = time.monotonic()
loads = []
with (proof / 'gate.jsonl').open('w') as log:
    process = subprocess.Popen(command, stdout=log, stderr=subprocess.STDOUT, env=environment)
    while process.poll() is None:
        loads.append({'elapsed': time.monotonic() - started, 'load': os.getloadavg()})
        time.sleep(2)
status = process.returncode
(proof / 'load.json').write_text(json.dumps(loads, indent=2) + '\n')
counts = collections.Counter()
top = collections.Counter()
skips = []
mutants = []
for number, line in enumerate((proof / 'gate.jsonl').read_text().splitlines(), 1):
    try:
        event = json.loads(line)
    except json.JSONDecodeError:
        continue
    action = event.get('Action')
    test = event.get('Test', '')
    if action in ['pass', 'fail', 'skip'] and test:
        counts[action] += 1
        if '/' not in test:
            top[action] += 1
        if action == 'skip':
            skips.append(test)
    if action == 'output' and ('caught' in event.get('Output', '') or 'mutant' in test):
        if any(name in test or name in event.get('Output', '') for name in ['no-new-func', 'prefer-arrow-callback']):
            mutants.append({'line': number, 'test': test, 'output': event.get('Output', '')})
summary = {'exit': status, 'wallSeconds': time.monotonic() - started, 'counts': {key: counts[key] for key in ['pass', 'fail', 'skip']}, 'topLevel': {key: top[key] for key in ['pass', 'fail', 'skip']}, 'skips': skips, 'nproc': int(subprocess.check_output(['nproc'], text=True)), 'loadMinMedianMax': [min(x['load'][0] for x in loads), statistics.median(x['load'][0] for x in loads), max(x['load'][0] for x in loads)]}
(proof / 'summary.json').write_text(json.dumps(summary, indent=2) + '\n')
(proof / 'mutant-lines.json').write_text(json.dumps(mutants, indent=2) + '\n')
print(json.dumps(summary), flush=True)
raise SystemExit(status)
