#!/usr/bin/env python3
"""Resume a cloud-interrupted probe run, then compare the complete 301 outputs."""
import importlib.util
import json
import os
from pathlib import Path
import shutil
import sys

original = Path(__file__).resolve().parents[3] / 'drivers/tsc'
sys.path.insert(0, str(original))
import driver

results, logs, bundle, scratch = map(lambda p: Path(p).resolve(), sys.argv[1:])
selection = json.loads((original / 'selection.json').read_text())
complete = []
pending = []
for row in selection['cases']:
    if (results / row['id'] / 'actual.exit').exists() and (logs / (row['id'] + '.json')).exists():
        complete.append(row)
    else:
        pending.append(row)
assert not (results / 'tiny' / 'actual.exit').exists(), 'resume expects tiny still pending'
scratch.mkdir(parents=True)
source = scratch / 'source'
source.mkdir()
(source / 'corpus').symlink_to(original / 'corpus', target_is_directory=True)
(source / 'tiny').symlink_to(original / 'tiny', target_is_directory=True)
(source / 'selection.json').write_text(json.dumps({**selection, 'cases': pending}) + '\n')
driver.ROOT = source
os.environ['TSC_JOBS'] = '1'
os.environ['ADAMIC_NONNULL_LOG'] = str(logs)
command = driver.command_argv(['node', str(bundle)])
resumed = scratch / 'results'
resumed.mkdir()
assert not driver.run(command, resumed, golden_root=original)
resumed_report = json.loads((resumed / 'report.json').read_text())
for identifier in [row['id'] for row in pending] + ['tiny']:
    destination = results / identifier
    destination.mkdir(exist_ok=True)
    for file in (resumed / identifier).iterdir():
        shutil.copyfile(file, destination / file.name)
failures = []
for identifier in [row['id'] for row in selection['cases']] + ['tiny']:
    golden = original / ('tiny' if identifier == 'tiny' else 'corpus/' + identifier)
    if driver.compare(results / identifier, golden, identifier):
        failures.append(identifier)
report = {'command': command, 'cases': 301, 'passed': 301 - len(failures), 'failed': failures,
          'seconds': None, 'cloud_interruption': True, 'completed_before_interrupt': len(complete),
          'resumed_report': resumed_report, 'all_outputs_recompared_after_resume': True}
(results / 'report.json').write_text(json.dumps(report, indent=2) + '\n')
print(json.dumps(report))
assert not failures
