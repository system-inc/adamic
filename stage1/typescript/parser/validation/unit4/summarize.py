"""Rebuild active top-level times and counts from go test -json events."""
import csv
import datetime
import json
import pathlib
import re

root = pathlib.Path(__file__).resolve().parent
package = root.parent.parent
locations = {}
for path in package.glob('*_test.go'):
    for number, line in enumerate(path.read_text().splitlines(), 1):
        match = re.match(r'func (Test\w+)\(t \*testing.T\)', line)
        if match:
            locations[match[1]] = f'stage1/typescript/parser/{path.name}:{number}'

def summarize(filename):
    events = [json.loads(line) for line in (root / filename).read_text().splitlines() if line.startswith('{')]
    states = {}
    counts = {'top_level': dict(pass_=0, fail=0, skip=0), 'all_tests': dict(pass_=0, fail=0, skip=0)}
    package_result = None
    for event in events:
        action, name = event['Action'], event.get('Test')
        if not name:
            if action in ('pass', 'fail', 'skip'):
                package_result = {'result': action, 'seconds': event.get('Elapsed')}
            continue
        if action in ('pass', 'fail', 'skip'):
            key = 'pass_' if action == 'pass' else action
            counts['all_tests'][key] += 1
            if '/' not in name:
                counts['top_level'][key] += 1
        if '/' in name:
            continue
        state = states.setdefault(name, {'test': name, 'seconds': 0.0, 'source': locations.get(name, '')})
        timestamp = datetime.datetime.fromisoformat(event['Time'].replace('Z', '+00:00'))
        if action in ('run', 'cont'):
            state['started'] = timestamp
        elif action in ('pause', 'pass', 'fail', 'skip'):
            started = state.pop('started', None)
            if started is not None:
                state['seconds'] += (timestamp - started).total_seconds()
            if action != 'pause':
                state['result'] = action
                state['go_elapsed'] = event.get('Elapsed')
    rows = sorted(states.values(), key=lambda state: state['seconds'], reverse=True)
    for row in rows:
        row['seconds'] = round(row['seconds'], 6)
        row.pop('started', None)
    return rows, counts, package_result

rows, counts, result = summarize('parser-final.jsonl')
with (root / 'after-times.csv').open('w', newline='') as stream:
    writer = csv.DictWriter(stream, fieldnames=['test', 'seconds', 'go_elapsed', 'result', 'source'])
    writer.writeheader()
    writer.writerows(rows)
(root / 'counts.json').write_text(json.dumps({'parser': result, 'counts': counts, 'timing_method': 'run-to-pause plus cont-to-pass; excludes the paused queue; includes parallel child completion'}, indent=2) + '\n')
print(json.dumps({'parser': result, 'counts': counts, 'largest': rows[:3]}, indent=2))
