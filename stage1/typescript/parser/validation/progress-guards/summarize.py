"""Summarize complete test events, counts and active top-level durations."""
import csv
import datetime
import json
import pathlib
import re

root = pathlib.Path(__file__).resolve().parent
locations = {}
for path in root.parent.parent.glob('*_test.go'):
    for number, line in enumerate(path.read_text().splitlines(), 1):
        match = re.match(r'func (Test\w+)\(t \*testing.T\)', line)
        if match:
            locations[match[1]] = f'stage1/typescript/parser/{path.name}:{number}'
report = json.loads((root / 'gate-results.json').read_text())
for label in report:
    events = [json.loads(line) for line in (root/f'{label}.jsonl').read_text().splitlines() if line.startswith('{')]
    counts = dict(pass_=0, fail=0, skip=0)
    top_counts = counts.copy()
    states = {}
    for event in events:
        action, name = event['Action'], event.get('Test')
        if not name:
            if action in ('pass', 'fail', 'skip'):
                report[label]['package'] = dict(result=action, seconds=event.get('Elapsed'))
            continue
        if action in ('pass', 'fail', 'skip'):
            key = 'pass_' if action == 'pass' else action
            counts[key] += 1
            if '/' not in name:
                top_counts[key] += 1
        if label != 'parser' or '/' in name:
            continue
        state = states.setdefault(name, dict(test=name, seconds=0.0, source=locations.get(name, '')))
        timestamp = datetime.datetime.fromisoformat(event['Time'].replace('Z', '+00:00'))
        if action in ('run', 'cont'):
            state['started'] = timestamp
        elif action in ('pause', 'pass', 'fail', 'skip'):
            started = state.pop('started', None)
            if started is not None:
                state['seconds'] += (timestamp-started).total_seconds()
            if action != 'pause':
                state['result'] = action
    report[label]['all_tests'] = counts
    report[label]['top_level_tests'] = top_counts
    if label == 'parser':
        rows = sorted(states.values(), key=lambda row: row['seconds'], reverse=True)
        with (root/'after-times.csv').open('w', newline='') as stream:
            writer = csv.DictWriter(stream, fieldnames=['test', 'seconds', 'result', 'source'], lineterminator='\n')
            writer.writeheader()
            for row in rows:
                row.pop('started', None)
                row['seconds'] = round(row['seconds'], 6)
                writer.writerow(row)
        report[label]['largest'] = rows[:3]
        report[label]['completion'] = [e['Output'].strip() for e in events if 'Output' in e and ('22497' in e['Output'] or 'longest native build' in e['Output'])][-5:]
(root/'report.json').write_text(json.dumps(report, indent=2)+'\n')
print(json.dumps(report, indent=2))
