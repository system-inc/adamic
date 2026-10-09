"""Summarize go test -json logs without conflating parent and leaf tests."""
import collections
import json
import sys
from pathlib import Path

result = {}
for filename in sys.argv[1:]:
    rows = []
    for line in Path(filename).read_text().splitlines():
        try:
            rows.append(json.loads(line))
        except json.JSONDecodeError:
            pass
    names = {row['Test'] for row in rows if 'Test' in row}
    leaves = {name for name in names if not any(other.startswith(name + '/') for other in names)}
    terminal = {row['Test']: row['Action'] for row in rows if row.get('Test') in leaves and row['Action'] in ('pass', 'fail', 'skip')}
    result[Path(filename).name] = {
        'leaf_counts': dict(collections.Counter(terminal.values())),
        'failures': sorted(name for name, action in terminal.items() if action == 'fail'),
        'incomplete': sorted(leaves - terminal.keys()),
        'concurrency': {name: action for name, action in terminal.items() if '/concurrency/' in name},
        'package_result': [row['Action'] for row in rows if 'Test' not in row and row['Action'] in ('pass', 'fail')],
    }
print(json.dumps(result, indent=2))
