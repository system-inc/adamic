#!/usr/bin/env python3
"""Build Lower's own-file handoff from retained, unmodified meter streams."""
from collections import Counter, defaultdict
import gzip
from html import escape
import json
from pathlib import Path
import re
import sys

run = Path(__file__).resolve().parents[1]
summary = json.loads((run / 'main/report.json').read_text())
tree = Path(summary['adapted_tree'])
def records(name):
    path = run / 'main' / name
    if path.exists():
        return [json.loads(line) for line in path.read_text().splitlines()]
    with gzip.open(str(path) + '.gz', 'rt') as source:
        return [json.loads(line) for line in source]
ordinary = records('census.jsonl')
latent = records('latent.jsonl')
assert latent[0]['latent_mode'] == 'first-error'
rows = [row for row in summary['files'] if row['source']]
assert len(rows) == 79
own = defaultdict(set)
pattern = re.compile(r'^(.+?):([0-9]+):([0-9]+): error (TS[0-9]+): (.*)', re.S)
for record in ordinary:
    for diagnostic in record.get('diagnostics', []):
        match = pattern.match(diagnostic)
        if match:
            own[str(Path(match[1]).resolve())].add(diagnostic)
def diag_key(diagnostic):
    match = pattern.match(diagnostic)
    return (int(match[2]), int(match[3]), match[4], match[5])
def location(where):
    match = re.fullmatch(r'(.+):([0-9]+):([0-9]+)', where)
    if match:
        return str(Path(match[1]).resolve()), int(match[2]), int(match[3])
    return None, 0, 0
findings = defaultdict(dict)
all_sites = set()
for record in latent[1:]:
    for finding in record.get('findings') or []:
        all_sites.add(tuple(finding[key] for key in ('kind', 'where', 'reason', 'text')))
        name, line, column = location(finding['where'])
        if name:
            key = tuple(finding[k] for k in ('kind', 'where', 'reason', 'text'))
            findings[name][key] = finding
site_counts = Counter(site[0] for site in all_sites)
for kind, expected in summary['latent_lowering']['totals'].items():
    assert site_counts[kind] == expected, (kind, site_counts[kind], expected)
coverage = {str(Path(record['file']).resolve()): record for record in latent[1:]}
assert len(coverage) == 79
owners = {row['reason']: row['owner'] for row in summary['latent_lowering']['reason_rows']}
ledger = []
for row in rows:
    path = str((tree / row['file']).resolve())
    diagnostics = sorted(own[path], key=diag_key)
    assert row['checker_own_file'] == (not diagnostics), row['file']
    assert row['own_file_diagnostics'] == len(diagnostics), row['file']
    candidates = [f for f in findings[path].values() if f['kind'] in ('NotYet', 'Refused')]
    candidates.sort(key=lambda f: (location(f['where'])[1:], f['kind'], f['reason'], f['text']))
    first_diagnostic = None
    if diagnostics:
        match = pattern.match(diagnostics[0])
        first_diagnostic = dict(where=f"{row['file']}:{match[2]}:{match[3]}", code=match[4], message=match[5])
    first_finding = dict(candidates[0], owner=owners[candidates[0]['reason']]) if candidates else None
    if first_finding:
        first_finding['where'] = first_finding['where'].replace(str(tree) + '/', '')
        first_finding['unit'] = first_finding.get('unit', '').replace(str(tree) + '/', '')
    statuses = Counter(unit['status'] for unit in coverage[path].get('units', []))
    first_class = first_diagnostic['code'] if first_diagnostic else (
        first_finding['kind'] + ': ' + first_finding['reason'] if first_finding else ('accepted by ordinary lowering; no first stop' if row['lowering'] else 'no NotYet/Refused observed (partial census)'))
    ledger.append(dict(file=row['file'], checker_own_file=row['checker_own_file'], first_diagnostic=first_diagnostic,
                       ordinary_lowering_attempted=row['lowering_attempted'], ordinary_lowering_pass=row['lowering'], ordinary_outcome=row['outcome'],
                       latent_record_present=path in coverage, unit_statuses=dict(statuses), first_finding=first_finding,
                       first_stop_class=first_class))
counts = dict(Counter(row['first_stop_class'] for row in ledger))
assert sum(counts.values()) == 79
assert sum(row['checker_own_file'] for row in ledger) == summary['totals']['checker_own_file']
output = dict(definition='First primary diagnostic or NotYet/Refused by source line, UTF-16 column, then kind/reason. Findings are attributed to their actual file across all census attempts; both refusal_scan and lowering phases count.',
              first_stop_classes=counts, files=ledger)
(run / 'lower-ledger.json').write_text(json.dumps(output, indent=2) + '\n')
def cell(text):
    return escape(str(text)).replace('|', '&#124;').replace('\n', '<br>')
lines = ['## For Lower: every source file', '',
         output['definition'], '',
         'Own-file pass means no primary diagnostic located in that file across the ordinary compiler-root census. Imported diagnostics can still block ordinary lowering. The latent record reaches every file, but diagnosed bodies and deferred generics may remain unexamined. A recorded finding is an observed blocker, not a successful native compilation. First finding here is source order, not event order or a claim that production lowering reached the file. Full raw diagnostic elaborations remain in census.jsonl.gz and lower-ledger.json.', '',
         '| File | Own-file checker and first diagnostic | Ordinary lowering | Latent reach and first observed NotYet/Refused (for own-file passes) |',
         '| --- | --- | --- | --- |']
for row in ledger:
    diag = row['first_diagnostic']
    check = 'pass' if not diag else f"fail: {diag['where']} {diag['code']}: {diag['message']}"
    lower = row['ordinary_outcome'] if row['ordinary_lowering_attempted'] else 'blocked by checker'
    finding = row['first_finding']
    reach = 'own-file checker fails; latent record retained'
    if row['checker_own_file']:
        reach = 'yes; '
        if finding:
            reach += f"{finding['where']} {finding['kind']}: {finding['reason']} (phase {finding.get('phase', 'unknown')}; owner {finding['owner']})"
        else:
            reach += ('no NotYet/Refused observed; ordinary lowering accepted, no backend tested' if row['ordinary_lowering_pass'] else 'no NotYet/Refused observed; partial census, not a lowering pass')
        if row['unit_statuses']:
            reach += '; units ' + ', '.join(f'{key}={value}' for key, value in sorted(row['unit_statuses'].items()))
    lines.append('| ' + ' | '.join(map(cell, [row['file'], check, lower, reach])) + ' |')
lines += ['', '### Count by first-stop class', '',
          'Each of the 79 files contributes once: its first TS code if its own checker fails, otherwise its first observed lowering reason. No observed finding stays an explicit coverage limitation. There are 23 own-file checker stops, 49 files with a first latent NotYet/Refused, six blocked ordinary roots without a latent finding, and one ordinary lowering pass with no latent finding. The second ordinary lowering pass, hostErrors.ts, is among the 49 observed latent findings because that program context differs.', '',
          '| Class | Files |', '| --- | ---: |']
for key, value in sorted(counts.items(), key=lambda pair: (-pair[1], pair[0])):
    lines.append(f'| {cell(key)} | {value} |')
lines += ['| Total | 79 |', '']
(run / 'lower-section.md').write_text('\n'.join(lines))
print(json.dumps(dict(files=len(ledger), own_file_pass=sum(row['checker_own_file'] for row in ledger), first_stop_classes=counts), indent=2))
