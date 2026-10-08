#!/usr/bin/env python3
"""Reconcile Callgrind self costs and prove reconciliation rejects a changed summary."""
import importlib.util
import json
from pathlib import Path
import re
import sys

scratch = Path(sys.argv[1]).resolve()
sys.dont_write_bytecode = True
spec = importlib.util.spec_from_file_location('scanner_profile', Path('stage1/typescript/scanner/profile.py'))
profile = importlib.util.module_from_spec(spec)
spec.loader.exec_module(profile)
results = {}
for workload in ['parse', 'service']:
    results[workload] = {}
    for mode in ['baseline', 'thin']:
        path = scratch / (mode + '-' + workload + '.callgrind')
        result = profile.summarize(path)
        results[workload][mode] = {'total': result['total'], 'top_self': result['self'][:15]}
        print(workload, mode, 'self reconciliation PASS', result['total'], flush=True)
path = scratch / 'baseline-parse.callgrind'
text = path.read_text()
mutant = scratch / 'accounting-mutant.callgrind'
mutant.write_text(re.sub(r'^summary: (\d+)', lambda match: 'summary: ' + str(int(match[1]) + 1), text, count=1, flags=re.MULTILINE))
try:
    profile.summarize(mutant)
except RuntimeError as error:
    if str(error) != 'callgrind self costs do not sum to summary':
        raise
    print('accounting mutant caught:', error, flush=True)
else:
    raise RuntimeError('accounting mutant survived')
(scratch / 'instruction-summary.json').write_text(json.dumps(results, indent=2) + '\n')
