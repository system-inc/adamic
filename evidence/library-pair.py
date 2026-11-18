"""Verify the historical 12 array-option refusals against the paired parser."""
import json
from pathlib import Path

from pair import location, rows, semantics

historical = json.loads(Path('evidence/library-sites.json').read_text())
filename = 'commandLineParser.ts-full.jsonl'
base, head = (rows(Path('evidence')/side/filename)[2] for side in ('base', 'head'))
entries = []
for row in historical:
    found = []
    for findings in (base, head):
        matches = [finding for finding in findings if location(finding) == row['Site']
                   and finding['reason'] == row['Words']
                   and finding['kind']=='Refused' and finding['phase']=='refusal_scan']
        assert len(matches) == 1, (row, matches)
        found.extend(matches)
    assert semantics(found[0]) == semantics(found[1])
    entries.append(dict(historical=row, base=found[0], head=found[1], disposition='Remains stopped at the same refusal'))
Path('evidence/library-pair.json').write_text(json.dumps(entries, indent=2)+'\n')
with Path('library-table.md').open('w') as out:
    print('| Site in `commandLineParser.ts` | Refusal in both compilers | Disposition |', file=out)
    print('| --- | --- | --- |', file=out)
    for entry in entries:
        row = entry['historical']
        site = row['Site'].split(':',1)[1]
        message = row['Words'].replace('|',r'\|')
        print(f'| {site} | {message} | {entry["disposition"]} |', file=out)
