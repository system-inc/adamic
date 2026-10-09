#!/usr/bin/env python3
"""Refresh the committed counts and observations from a completed measurement."""
import csv
import json
from pathlib import Path
import shutil
import sys

source = Path(sys.argv[1]).resolve()
here = Path(__file__).resolve().parent
summary = json.loads((source / 'summary.json').read_text())
observations = {'summary': summary, 'runs': []}
rows = []
for run in summary['runs']:
    data = json.loads((source / (run['name'] + '.json')).read_text())
    observations['runs'].append({'name': run['name'], 'examples': data['examples']})
    rows.extend({'run': run['name'], **row} for row in data['rows'])
(here / 'observations.json').write_text(json.dumps(observations, indent=2) + '\n')
with (here / 'inputs.csv').open('w', newline='') as out:
    writer = csv.DictWriter(out, lineterminator='\n', fieldnames=['run', 'site', 'input', 'assignments', 'violations', 'reads', 'outOfTypeReads'])
    writer.writeheader()
    writer.writerows(rows)
lines = ['# Runtime counts', '', 'Pinned stock TypeScript 6.0.3; Node ' + summary['node'] + '.', '',
         'Reads count getter executions after the site write and before a later store.',
         'Violations use the declared domain, not a comparison with the old value.', '',
         '| Run | Site | Assignments | Violations | Later reads | Out-of-type reads |',
         '| --- | --- | ---: | ---: | ---: | ---: |']
for run in summary['runs']:
    if run['name'].startswith('mutant-') or run['name'] in ['static', 'typed']:
        continue
    for site, counts in run['totals'].items():
        lines.append('| ' + ' | '.join([run['name'], site] + [str(counts[key]) for key in
            ['assignments', 'violations', 'reads', 'outOfTypeReads']]) + ' |')
lines += ['', 'The unreached fixture asserts an empty observation before the compatible controls:',
          'both sites have zero assignments, violations and reads there.', '',
          'The `fixtures` rows then test reached compatible stores. The parent compatible',
          'store calls the observation helper directly because the upstream parent site',
          'always stores undefined. It is a predicate control, not an upstream execution.', '',
          '`inputs.csv` preserves one row per process and source filename; repeated virtual',
          'filenames can belong to different compiler tests and are not unique case IDs.',
          'Processes with no assignment have no CSV row. `observations.json` retains',
          'process counts, commands, bundle hashes and sampled write/read stacks.', '']
(here / 'counts.md').write_text('\n'.join(lines))
logs = here / 'logs'
logs.mkdir(exist_ok=True)
for file in source.glob('*.log'):
    shutil.copyfile(file, logs / file.name)
shutil.copyfile(source / 'acceptance-projects/report.json', here / 'acceptance-report.json')
print('refreshed counts.md, inputs.csv, observations.json, acceptance-report.json and logs')
