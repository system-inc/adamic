#!/usr/bin/env python3
"""Run the unchanged lint package and record its complete result and machine data."""
import json
from pathlib import Path
import subprocess
import time

root = Path(__file__).resolve().parents[5]
out = Path(__file__).resolve().parent / 'evidence'
command = ['go', 'test', './stage1/cohere/lint', '-count=1', '-v', '-json', '-timeout=90m']
load_before = Path('/proc/loadavg').read_text().strip()
nproc = int(subprocess.check_output(['nproc'], text=True))
started = time.monotonic()
with (out / 'whole.jsonl').open('w') as log:
    result = subprocess.run(command, cwd=root, stdout=log, stderr=subprocess.STDOUT)
wall = time.monotonic() - started
counts = {'pass': 0, 'fail': 0, 'skip': 0}
skips, failures = [], []
package_fail = 0
with (out / 'whole.log').open('w') as readable:
    for line in (out / 'whole.jsonl').read_text().splitlines():
        try:
            event = json.loads(line)
        except json.JSONDecodeError:
            readable.write(line + '\n')
            continue
        readable.write(event.get('Output', ''))
        action = event.get('Action')
        if event.get('Test') and action in counts:
            counts[action] += 1
            if action == 'skip':
                skips.append(event['Test'])
            if action == 'fail':
                failures.append(event['Test'])
        elif not event.get('Test') and action == 'fail':
            package_fail += 1
summary = {'command': command, 'exit': result.returncode, 'counts_include_subtests': True,
           **counts, 'package_fail': package_fail, 'skips': skips, 'failures': failures,
           'wall_seconds': round(wall, 3), 'nproc': nproc, 'load_before': load_before,
           'load_after': Path('/proc/loadavg').read_text().strip()}
(out / 'whole-summary.json').write_text(json.dumps(summary, indent=2) + '\n')
print(json.dumps(summary))
raise SystemExit(result.returncode)
