"""Exercise the original stock/full-report and rerun audits on the small control."""
import copy
import importlib
import json
import os
from pathlib import Path
import subprocess
import sys

root = Path.cwd()
evidence = root/'review/compiler/loom-census-deferral'
source = evidence/'deferred/source'
manifest = evidence/'deferred/manifest.json'
manifest.write_text(json.dumps([{'file': p.name} for p in sorted(source.glob('*.a'))]))
with (evidence/'stock-units.log').open('w') as log:
    subprocess.run(['node', 'stage3/census/latent/stock_units.cjs', str(source), str(manifest),
                    str(evidence/'stock-units.json')], stdout=log, stderr=subprocess.STDOUT,
                   timeout=30, check=True)
sys.path.insert(0, str(root/'stage3/census/latent'))
import full_report
import rerun2
from unit_status import deferred_summary
rows = [json.loads(line) for line in (evidence/'deferred/full.jsonl').read_text().splitlines()]
stock = json.loads((evidence/'stock-units.json').read_text())
full_report.audit_units(rows, stock, source)
summary = full_report.lowering_summary(rows, {row['file'] for row in rows[1:]}, full_report.MEASUREMENT)
summary.update(deferred_summary([u for row in rows[1:] for u in row['units']]))
owners = json.loads((root/'stage3/meter/owners.json').read_text())
full_report.audit_counts(summary, rows, owners)
mutant = copy.deepcopy(summary)
mutant.pop('never_reached')
try:
    full_report.audit_counts(mutant, rows, owners)
except AssertionError as error:
    assert str(error) == 'deferred coverage gap', error
    print('full-report missing no-caller gap ledger: caught')
else:
    raise AssertionError('missing gap ledger survived')
rerun = rerun2.summarize(rows, [{'file':row['file']} for row in rows[1:]], denominator=2)
assert rerun['deferred_generic_instantiation'] == summary['deferred_generic_instantiation']
assert rerun['never_reached'] == summary['never_reached']
(evidence/'full-report-summary.json').write_text(json.dumps(summary, indent=2)+'\n')
print('PASS: stock spans, full-report counts, no-caller ledger, rerun accounting')
