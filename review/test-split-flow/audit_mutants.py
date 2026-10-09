#!/usr/bin/env python3
"""Prove the evidence audit rejects missing units, changed inputs and lost checks."""
import copy
import json
import re
from pathlib import Path
import subprocess
import sys

review = Path(__file__).resolve().parent
root = review.parents[1]
logs = Path('/tmp/test-split-flow/audit-mutants')
logs.mkdir(parents=True, exist_ok=True)
source = (review/'audit.py').read_text().replace("review = root / 'review/test-split-flow'", 'review = Path(sys.argv[1])')
checker = logs/'checker.py'
# Preserve the real repository root while giving the checker synthetic ledgers.
source = source.replace('root = Path(__file__).resolve().parents[2]', 'root = Path('+repr(str(root))+')')
checker.write_text(source)
original = {name:json.loads((review/(name+'.json')).read_text()) for name in ['before','after','units']}
results = []
for name in ['missing-measurement','changed-fixture-hash','changed-measured-test-hash','lost-write-checks']:
    directory = logs/name
    directory.mkdir(exist_ok=True)
    data = copy.deepcopy(original)
    if name == 'missing-measurement':
        data['after']['tests'].pop()
        catcher = 'len(new) == len(after'
    elif name == 'changed-fixture-hash':
        data['after']['inputs']['internal/flow/testdata/joins.a'] = '0'*64
        catcher = "before['inputs'][r['program']]"
    elif name == 'changed-measured-test-hash':
        data['after']['inputs']['internal/flow/flow_test.go'] = '0'*64
        catcher = 'measured input changed'
    else:
        owner = next(r for r in data['after']['tests'] if r['package']=='fresh' and re.search(r'[1-9][0-9]* writes,', Path(r['log']).read_text()))
        log = Path(owner['log']).read_text()
        changed, count = re.subn(r'(\d+) writes, (\d+) proven not to close a cycle', '0 writes, 0 proven not to close a cycle', log)
        assert count == 1 and changed != log
        forged = directory/'forged.jsonl'
        forged.write_text(changed)
        owner['log'] = str(forged)
        catcher = 'Writes'
    for ledger,value in data.items():
        (directory/(ledger+'.json')).write_text(json.dumps(value))
    command = [sys.executable,str(checker),str(directory)]
    output = directory/'audit.log'
    with output.open('w') as log:
        run = subprocess.run(command,stdout=log,stderr=subprocess.STDOUT)
    assert run.returncode == 1 and 'AssertionError' in output.read_text() and catcher in output.read_text(), name
    results.append({'name':name,'exit':run.returncode,'catcher':catcher,'log':str(output)})
    print(name+': caught')
(review/'audit-mutants.json').write_text(json.dumps(results,indent=2)+'\n')
