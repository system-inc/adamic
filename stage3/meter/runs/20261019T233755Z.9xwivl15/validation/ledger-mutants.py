#!/usr/bin/env python3
"""Corrupt real run artifacts to prove Lower's accounting checks can fail."""
import gzip
import json
from pathlib import Path
import shutil
import subprocess
import tempfile

run = Path(__file__).resolve().parents[1]
for name in ['own-file-status', 'latent-coverage', 'drop-refused']:
    scratch = Path(tempfile.mkdtemp(prefix='meter-ledger-' + name + '-'))
    (scratch/'validation').mkdir()
    (scratch/'main').mkdir()
    shutil.copy2(run/'validation/lower-ledger.py', scratch/'validation/lower-ledger.py')
    report = json.loads((run/'main/report.json').read_text())
    latent_path = run/'main/latent.jsonl'
    if latent_path.exists():
        latent = [json.loads(line) for line in latent_path.read_text().splitlines()]
    else:
        with gzip.open(str(latent_path)+'.gz', 'rt') as source:
            latent = [json.loads(line) for line in source]
    if name == 'own-file-status':
        witness = next(row for row in report['files'] if row['file']=='src/compiler/binder.ts')
        assert witness['checker_own_file']
        witness['checker_own_file'] = False
    elif name == 'latent-coverage':
        removed = latent.pop()
        # Keep every site so the coverage assertion is the only possible catcher.
        latent[1]['findings'] = (latent[1].get('findings') or []) + (removed.get('findings') or [])
    else:
        count = sum(f['kind']=='Refused' for row in latent[1:] for f in row.get('findings') or [])
        assert count > 0
        for row in latent[1:]:
            row['findings'] = [f for f in row.get('findings') or [] if f['kind']!='Refused']
    (scratch/'main/report.json').write_text(json.dumps(report))
    census = run/'main/census.jsonl'
    if not census.exists(): census = Path(str(census)+'.gz')
    shutil.copy2(census,scratch/'main'/census.name)
    (scratch/'main/latent.jsonl').write_text(''.join(json.dumps(row)+'\n' for row in latent))
    log_path = run/'validation'/('ledger-mutant-'+name+'.log')
    with log_path.open('w') as log:
        result = subprocess.run(['python3',str(scratch/'validation/lower-ledger.py')], stdout=log, stderr=log, timeout=120)
    observed = log_path.read_text()
    assert result.returncode == 1 and 'AssertionError' in observed, (name,result.returncode,observed)
    print(name+': caught by real-artifact ledger assertion, exit 1')
