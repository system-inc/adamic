#!/usr/bin/env python3
"""Compare an in-process project-option report to the pinned checker ledger."""
import argparse
import csv
import io
import json
from pathlib import Path
import subprocess

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('report', type=Path)
parser.add_argument('tree', type=Path)
parser.add_argument('--ledger-ref', default='3f0926c0a55a7b5f64f037b1745e0e984e08c8be')
parser.add_argument('--ledger-file', default='rows.csv')
parser.add_argument('--include-json', action='store_true')
parser.add_argument('--save', type=Path)
parser.add_argument('--disposition', type=Path)
parser.add_argument('--allow-project-errors', action='store_true', help='Compare sites while retaining production prelude/soundness errors.')
args = parser.parse_args()
repository = Path(__file__).resolve().parents[2]
ledger = subprocess.check_output(['git', 'show', args.ledger_ref + ':stage3/ledger/checker-259/' + args.ledger_file], cwd=repository, text=True)
rows = list(csv.DictReader(io.StringIO(ledger)))
options = ['noUncheckedIndexedAccess', 'exactOptionalPropertyTypes', 'useUnknownInCatchVariables', 'strictBindCallApply']
if args.include_json:
    options.append('JSON.stringify')
expected = {}
for row in rows:
    json_site = args.include_json and row['option'] == 'other' and 'JSON.stringify overload' in row['cause']
    if row['option'] not in options and not json_site:
        continue
    key = (row['file'], int(row['line']), int(row['column']), int(row['code'][2:]))
    if key in expected:
        raise SystemExit(f'duplicate ledger site: {key}')
    expected[key] = {'JSON.stringify'} if json_site else {option for option in options if option in row['removed_by']}
report = json.loads(args.report.read_text())
if report['project_errors'] and not args.allow_project_errors:
    raise SystemExit(f"project errors: {report['project_errors']}")
actual = {}
for site in report['sites']:
    site['file'] = str(Path(site['file']).relative_to(args.tree.resolve()))
    site['message'] = site['message'].replace(str(args.tree.resolve()) + '/', '')
    key = (site['file'], site['line'], site['column'], site['code'])
    if key in actual:
        raise SystemExit(f'duplicate measured site: {key}')
    actual[key] = set(site['options'])
if expected != actual:
    missing = set(expected) - set(actual)
    extra = set(actual) - set(expected)
    wrong_options = {key: (expected[key], actual[key]) for key in expected.keys() & actual.keys() if expected[key] != actual[key]}
    raise SystemExit(f'missing={sorted(missing)}, extra={sorted(extra)}, wrong_options={wrong_options}')
counts = dict.fromkeys(options, 0)
for site in report['sites']:
    if not site['options']:
        raise SystemExit(f'unattributed stricter site: {site}')
    counts[site['options'][0]] += 1
print(f'project errors={len(report["project_errors"])}; sites={len(actual)}; missing=0; extra=0; wrong option attribution=0')
print('; '.join(f'{option}={counts[option]}' for option in options))
if args.save:
    args.save.write_text(json.dumps(report, indent=2) + '\n')

if args.disposition:
    remaining = set()
    for site in report.get('remaining_sites', []):
        file = str(Path(site['file']).relative_to(args.tree.resolve()))
        remaining.add((file, site['line'], site['column'], site['code']))
    if not remaining.issubset(actual):
        raise SystemExit('remaining error sites must be recorded sites')
    with args.disposition.open('w', newline='') as output:
        writer = csv.writer(output, lineterminator="\n")
        writer.writerow(['file', 'line', 'column', 'code', 'options', 'disposition'])
        for site in report['sites']:
            key = (site['file'], site['line'], site['column'], site['code'])
            writer.writerow([*key, ';'.join(site['options']), 'checker-error' if key in remaining else 'deferred-to-lowering'])
    print(f'checker errors={len(remaining)}; deferred to guarded lowering or explicit refusal={len(actual)-len(remaining)}')
