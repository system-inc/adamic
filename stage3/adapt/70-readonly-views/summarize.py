#!/usr/bin/env python3
"""Preserve the controlled latent comparison and complete remaining-site list."""
import collections
import gzip
import hashlib
import json
import pathlib
import sys

if len(sys.argv) not in (6, 7):
    raise SystemExit('usage: summarize.py BEFORE_TREE BEFORE_JSONL AFTER_TREE AFTER_JSONL OUTPUT [REMAINING_PATH]')
before_tree, before_raw, after_tree, after_raw, out = map(pathlib.Path, sys.argv[1:6])
out.mkdir(parents=True, exist_ok=True)
label = 'measured on a checker-rejected program'

def normalize(value, tree):
    if isinstance(value, str):
        return value.replace(str(tree.resolve()) + '/', '')
    if isinstance(value, list):
        return [normalize(item, tree) for item in value]
    if isinstance(value, dict):
        return {key: normalize(item, tree) for key, item in value.items()}
    return value

def read(tree, raw):
    rows = normalize([json.loads(line) for line in raw.read_text().splitlines()], tree)
    assert rows[0]['checker_rejected'] and rows[0]['measurement'] == label
    sites = {tuple(f[key] for key in ['kind', 'where', 'reason', 'text']) for row in rows[1:]
             for f in row['findings'] if ' seen as ' in f['reason']}
    units = {unit['where']: unit['status'] for row in rows[1:] for unit in row['units']}
    files = {p.relative_to(tree).as_posix(): hashlib.sha256(p.read_bytes()).hexdigest()
             for p in (tree / 'src/compiler').rglob('*.ts')}
    measured_files = [row['file'] for row in rows[1:]]
    assert len(measured_files) == len(files) and set(measured_files) == set(files), \
        'incomplete or duplicate compiler-file measurement'
    return rows, sites, units, files

before, first, before_units, before_files = read(before_tree, before_raw)
after, last, after_units, after_files = read(after_tree, after_raw)
removed = first - last
added = last - first
fields = ['kind', 'where', 'reason', 'text']
def records(sites):
    return [dict(zip(fields, site), measurement=label) for site in sorted(sites)]

summary = {
    'measurement': label,
    'count_definition': 'unique (kind, where, reason, text), reason contains " seen as "',
    'before': len(first), 'after': len(last), 'removed': records(removed), 'added': records(added),
    'before_per_reason': dict(sorted(collections.Counter(s[2] for s in first).items())),
    'after_per_reason': dict(sorted(collections.Counter(s[2] for s in last).items())),
    'checker_diagnostics_before': len(before[0]['diagnostics']),
    'checker_diagnostics_after': len(after[0]['diagnostics']),
    'checker_diagnostics_equal': before[0]['diagnostics'] == after[0]['diagnostics'],
    'files_before': len(before)-1, 'files_after': len(after)-1,
    'eligible_units_before': sum(status != 'skipped_checker_body' for status in before_units.values()),
    'eligible_units_after': sum(status != 'skipped_checker_body' for status in after_units.values()),
    'newly_eligible_units': sorted(key for key, status in after_units.items()
                                  if status != 'skipped_checker_body' and before_units.get(key) == 'skipped_checker_body'),
    'no_longer_eligible_units': sorted(key for key, status in before_units.items()
                                     if status != 'skipped_checker_body' and after_units.get(key) == 'skipped_checker_body'),
    'source_hashes_before': before_files, 'source_hashes_after': after_files,
    'changed_compiler_files': sorted(key for key in before_files.keys() | after_files.keys()
                                    if before_files.get(key) != after_files.get(key)),
    'remaining': records(last),
}
(out / 'census.json').write_text(json.dumps(summary, indent=2) + '\n')
for name, rows in [('before', before), ('after', after)]:
    with (out / (name + '.jsonl.gz')).open('wb') as target:
        with gzip.GzipFile(filename='', mode='wb', fileobj=target, mtime=0) as compressed:
            compressed.write(('\n'.join(json.dumps(row) for row in rows) + '\n').encode())
text = '# Remaining variance refusals\n\n'
text += f'{len(last)} unique sites. Every reason below is the exact latent reason, measured on a checker-rejected program. '
text += 'An entry is a retained refusal, not a claim of a reachable type violation. See README.md for declines and PUBLIC_VIEWS.md for public targets.\n\n'
for site in sorted(last):
    text += f'- `{site[1]}`: {site[2]}\n'
remaining_path = pathlib.Path(sys.argv[6]) if len(sys.argv) == 7 else out.parent / 'REMAINING.md'
remaining_path.write_text(text)
print(json.dumps({key: value for key, value in summary.items() if key in [
    'before', 'after', 'files_before', 'files_after', 'checker_diagnostics_before', 'checker_diagnostics_after',
    'checker_diagnostics_equal', 'changed_compiler_files', 'newly_eligible_units', 'no_longer_eligible_units']}, indent=2))
