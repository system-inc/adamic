#!/usr/bin/env python3
"""Prove the executed union and apply the original verdict to the same-tree whole run."""
import argparse
import json
import os
from pathlib import Path
from check import check_results
from shard_plan import LANE, load_plan, names_digest


def merge(results, whole, plan=None):
    plan = plan or load_plan()
    errors, records, reports = [], [], []
    for unit in plan['units']:
        report = json.loads((results / unit['name'] / 'report.json').read_text())
        reports.append(report)
        if report['name'] != unit['name'] or report['inputs_hash'] != unit['inputs_hash'] or report['artifact'] != plan['artifact']:
            errors.append('unit input identity mismatch: ' + unit['name'])
        if report['status'] != 'pass' or report['counts'] != unit['expect'] or report['seconds'] >= 30:
            errors.append('unit verdict rejected: ' + unit['name'])
        observed = json.loads((results / unit['name'] / 'observations.json').read_text())
        if names_digest(observed) != unit['names_digest']:
            errors.append('unit observation identity mismatch: ' + unit['name'])
        counted = {'pass': sum(row['status'] == 'pass' for row in observed),
                   'fail': sum(row['status'] == 'fail' for row in observed),
                   'skip': sum(row['status'] == 'pending' for row in observed)}
        if counted != report['counts']:
            errors.append('unit evidence counts mismatch: ' + unit['name'])
        records.extend(observed)
    execution = json.loads((whole / 'execution.json').read_text())
    if execution.get('artifact') != plan['artifact']:
        errors.append('unsharded lane did not run on the same hashed artifact')
    profile = [json.loads(line) for line in (whole / 'whole-tasks.jsonl').read_text().splitlines()]
    whole_records = [dict(record, task=row['task'], status=status) for row in profile
                     for status, key in [('pass', 'passes'), ('fail', 'errors')] for record in row[key]]
    if names_digest(records) != names_digest(whole_records):
        errors.append('executed test union differs from the unsharded lane')
    # Identity counts retain multiplicity, so duplicate titles cannot hide missing tests.
    counts = {status: sum(row['counts'][status] for row in reports) for status in ('pass', 'fail', 'skip')}
    oracle = json.loads((whole / 'oracle/report.json').read_text())
    whole_counts = dict(zip(('pass', 'fail', 'skip'), (oracle['counts'][key] for key in ('passing', 'failing', 'pending'))))
    if counts != whole_counts or counts != plan['counts']:
        errors.append(f'counts differ: units {counts}, whole {whole_counts}, plan {plan["counts"]}')
    expected = json.loads((LANE / 'expected.json').read_text())
    failures = [name for row in reports for name in row['failed_tests']]
    paths = [name for row in reports for name in row['baseline_diffs']]
    if failures != expected['failed_tests'] or paths != expected['baseline_diffs']:
        errors.append('merged failure titles or baseline paths differ from the unchanged sanction')
    # Recheck raw shard diffs against the whole-run diff, not just reported path sets.
    diff = b''.join((results / row['name'] / 'oracle/baseline.diff').read_bytes() for row in reports)
    if diff != (whole / 'oracle/baseline.diff').read_bytes():
        errors.append('merged baseline bytes differ from the unsharded lane')
    whole_verdict = check_results(whole)
    if whole_verdict['status'] != 'pass':
        errors.extend('unsharded lane: ' + message for message in whole_verdict['errors'])
    return dict(status='fail' if errors else 'pass', errors=errors, counts=counts,
                whole_counts=whole_counts, units=len(reports), tests=len(records),
                artifact=plan['artifact'], failed_tests=failures, baseline_diffs=paths,
                whole_verdict=whole_verdict['status'])


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('results', type=Path)
    parser.add_argument('--whole', type=Path, required=True, help='same-artifact unsharded run.sh results')
    parser.add_argument('--output', type=Path)
    args = parser.parse_args()
    plan = load_plan()
    artifact = Path(os.environ.get('STAGE3_ARTIFACT', str(Path.home() / '.cache/adamic-stage3-artifacts' / plan['artifact'])))
    os.environ['TSC_ADAPT_TYPESCRIPT'] = str(artifact / 'api')
    try:
        report = merge(args.results.resolve(), args.whole.resolve(), plan)
    except (OSError, ValueError, KeyError) as error:
        report = dict(status='fail', errors=['missing or invalid shard evidence: ' + str(error)])
    if args.output:
        args.output.write_text(json.dumps(report, indent=2) + '\n')
    print(json.dumps(report, indent=2))
    raise SystemExit(0 if report['status'] == 'pass' else 1)
