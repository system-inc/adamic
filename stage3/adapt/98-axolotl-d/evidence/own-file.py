# Counts own-file checker diagnostics the way the daily meter does, by calling stage3/meter/report.py's report()
# unchanged on each census run, and compares every source file between a baseline and each other run.
# Usage: python3 own-file.py <output.json> <label>=<tree>:<census-directory> ...
# The first labelled run is the baseline. Each census directory holds census.jsonl written by stage3/census/tool.
import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[2] / 'meter'))
import report  # noqa: E402

output, *runs = sys.argv[1:]
tables = {}
for argument in runs:
    label, location = argument.split('=', 1)
    tree, census = location.split(':', 1)
    result = report.report(Path(tree), Path(census), 'own-file')
    rows = result['files'] if isinstance(result, dict) else result
    tables[label] = {row['file']: row['own_file_diagnostics'] for row in rows if row['source']}
labels = list(tables)
baseline = tables[labels[0]]
for label in labels[1:]:
    if set(tables[label]) != set(baseline):
        raise SystemExit(f'{label} measures different files from {labels[0]}')
comparison = {'baseline': labels[0], 'files': len(baseline), 'tables': tables, 'changes': {}}
for label in labels[1:]:
    changes = {name: [baseline[name], tables[label][name]] for name in sorted(baseline) if baseline[name] != tables[label][name]}
    comparison['changes'][label] = changes
    print(f'{label} against {labels[0]}: {len(changes)} files changed, own-file total {sum(baseline.values())} to {sum(tables[label].values())}')
    for name, (before, after) in changes.items():
        print(f'  {name}: {before} to {after}')
Path(output).write_text(json.dumps(comparison, indent=1, sort_keys=True) + '\n')
