#!/usr/bin/env python3
"""Check an upstream observation against the deliberately sanctioned exception."""
import argparse
import difflib
import json
from pathlib import Path
import os
import re
import subprocess
import tempfile
import sys

LANE = Path(__file__).resolve().parent
sys.path.append(str(LANE.parent / 'oracle'))
from failure_details import read_failures, cause_counts, diff_previews, markdown

ANSI = re.compile(r'\x1b\[[0-?]*[ -/]*[@-~]')


def test_counts(log):
    log = ANSI.sub('', log)
    return {name: int((re.findall(r'^\s*(\d+) ' + name + r'\b', log, re.MULTILINE)
                       or ['0'])[-1]) for name in ('passing', 'failing', 'pending')}


def failed_tests(log):
    """Retain the reporter's complete title, joining its indented title lines."""
    log = ANSI.sub('', log)
    summary = list(re.finditer(r'^\s*\d+ failing\b', log, re.MULTILINE))
    if not summary:
        return []
    titles = []
    parts = None
    for line in log[summary[-1].end():].splitlines():
        start = re.fullmatch(r'\s*(\d+)\)\s*(.*)', line)
        if start:
            if parts is not None or int(start[1]) != len(titles) + 1:
                raise ValueError('malformed numbered failure title')
            parts = []
            line = start[2]
        if parts is not None:
            text = line.strip()
            if not text:
                continue
            if text.startswith(('Error:', 'AssertionError:', 'at ')):
                raise ValueError('incomplete failure title')
            parts.append(text.removesuffix(':') if text.endswith(':') else text)
            if text.endswith(':'):
                titles.append(' '.join(parts))
                parts = None
    if parts is not None:
        raise ValueError('unterminated failure title')
    return titles


def inspect(execution, oracle, log, expected, api_errors):
    errors = list(api_errors)
    platform = execution.get('platform', {})
    platform_key = '/'.join(str(platform.get(key, 'unknown')) for key in ('os', 'arch', 'node'))
    counts = expected['platforms'].get(platform_key)
    if counts is None:
        return dict(status='fail', errors=[f'unknown platform: {platform_key}; add measured expectations deliberately'])

    def equal(label, actual, wanted):
        if actual != wanted:
            errors.append(f'{label}: expected {wanted!r}, observed {actual!r}')

    equal('apply exit', execution.get('apply_exit'), expected['apply_exit'])
    equal('oracle exit', execution.get('oracle_exit'), expected['oracle_exit'])
    equal('oracle status', oracle.get('status'), expected['oracle_status'])
    for phase, code in expected['phase_exits'].items():
        equal(f'{phase} exit', oracle.get('phases', {}).get(phase, {}).get('exit'), code)
    equal('test counts', oracle.get('counts'), counts)
    equal('reporter counts', test_counts(log), counts)
    try:
        titles = failed_tests(log)
        equal('failed test names', titles, expected['failed_tests'])
        equal('number of reported failure titles', len(titles), counts['failing'])
    except ValueError as error:
        titles = []
        errors.append(f'failed test names: {error}')
    equal('baseline paths', oracle.get('baseline_diffs'), expected['baseline_diffs'])
    equal('node version', oracle.get('node'), platform.get('node'))
    equal('runners', oracle.get('runners'), 'all')
    equal('test filter', oracle.get('tests'), None)
    # Eight, fixed, whatever the box (stage3/oracle/run.py).
    equal('workers', oracle.get('workers'), 8)
    return dict(status='fail' if errors else 'pass', errors=errors,
                counts=oracle.get('counts'), failed_tests=titles,
                baseline_diffs=oracle.get('baseline_diffs'))


def api_check(results, sanctioned):
    tree = results / 'adapted-tree'
    relative = 'api/typescript.d.ts'
    pin = sanctioned['source_commit']
    head = subprocess.check_output(['git', '-C', str(tree), 'rev-parse', 'HEAD'], text=True).strip()
    if head != pin:
        return [f'API source pin: expected {pin}, observed {head}'], {}
    original = subprocess.check_output(['git', '-C', str(tree), 'show',
                                        f'{pin}:tests/baselines/reference/{relative}'])
    reference = tree / 'tests/baselines/reference' / relative
    local = tree / 'tests/baselines/local' / relative
    old, new = reference.read_bytes(), local.read_bytes()
    recorded_diff = (results / 'oracle/baseline.diff').read_bytes()
    actual_diff = ''.join(difflib.unified_diff(old.decode().splitlines(True), new.decode().splitlines(True),
                                              fromfile='reference/' + relative, tofile='local/' + relative)).encode()
    errors = []
    if recorded_diff != actual_diff:
        errors.append('baseline.diff does not exactly describe the API snapshots (extra, missing, or changed diff lines)')
    env = dict(os.environ)
    cache = Path(env.get('STAGE3_CACHE', str(Path.home() / '.cache/adamic-stage3')))
    env['NODE_PATH'] = str(cache / 'api/node_modules') + os.pathsep + env.get('NODE_PATH', '')
    with tempfile.TemporaryDirectory(prefix='stage3-lane-api-') as scratch:
        pristine = Path(scratch) / 'original.d.ts'
        pristine.write_bytes(original)
        normalized = json.loads(subprocess.check_output(
            ['node', str(LANE / 'normalize-api.cjs'), str(pristine), str(reference), str(local)],
            text=True, env=env, stderr=subprocess.PIPE))
    wanted = {row['declaration']: row for row in sanctioned['normalized_changes']}
    # Some adapters already update their proved reference lines (currently 40).
    # Require these accepted reference changes to be an exact subset of the sanction.
    for row in normalized['reference_changes']:
        if wanted.get(row['declaration']) != row:
            errors.append(f"unsanctioned API reference declaration: {row['declaration']}")
    observed = {row['declaration']: row for row in normalized['composed_changes']}
    for key in sorted(wanted.keys() | observed.keys()):
        if key not in wanted:
            errors.append(f'unsanctioned API declaration: {key}')
        elif key not in observed:
            errors.append(f'missing sanctioned API declaration: {key}')
        elif wanted[key] != observed[key]:
            errors.append(f'changed sanctioned API declaration: {key}')
    return errors, dict(reference_declarations=len(normalized['reference_changes']),
                        composed_declarations=len(normalized['composed_changes']),
                        sanctioned_declarations=len(wanted),
                        added_lines=sum(line.startswith(b'+') and not line.startswith(b'+++')
                                        for line in recorded_diff.splitlines()),
                        removed_lines=sum(line.startswith(b'-') and not line.startswith(b'---')
                                          for line in recorded_diff.splitlines()))


def check_results(results, expected_file=LANE / 'expected.json',
                  sanctioned_file=LANE / 'sanctioned-api.json'):
    try:
        expected = json.loads(expected_file.read_text())
        sanctioned = json.loads(sanctioned_file.read_text())
        execution = json.loads((results / 'execution.json').read_text())
        platform_key = '/'.join(str(execution.get('platform', {}).get(key, 'unknown'))
                                for key in ('os', 'arch', 'node'))
        if platform_key not in expected['platforms']:
            return dict(status='fail', errors=[f'unknown platform: {platform_key}; add measured expectations deliberately'])
        oracle = json.loads((results / 'oracle/report.json').read_text())
        api_errors, api = api_check(results, sanctioned)
        report = inspect(execution, oracle, (results / 'oracle/tests.log').read_text(),
                         expected, api_errors)
        report['api'] = api
        report['platform'] = execution['platform']
        report['execution'] = execution
        report['results'] = str(results.resolve())
    except (OSError, ValueError, KeyError, TypeError, subprocess.CalledProcessError) as error:
        report = dict(status='fail', errors=[f'missing or invalid lane evidence: {error}'])
    try:
        log = (results / 'oracle/tests.log').read_text(errors='replace')
        diff_file = results / 'oracle/baseline.diff'
        diff = diff_file.read_text() if diff_file.exists() else ''
        report['failures'] = read_failures(log, diff, results / 'oracle/mocha-errors')
        report['cause_counts'] = cause_counts(report['failures'])
        report['baseline_previews'] = list(diff_previews(diff).values())
    except (OSError, ValueError, KeyError, TypeError) as error:
        report['failure_details_error'] = str(error)
    return report


def write_verdict(results, report):
    message = 'PASS stage3 landing lane' if report['status'] == 'pass' else (
        'FAIL stage3 landing lane: ' + report['errors'][0].splitlines()[0])
    if report.get('cause_counts') is not None:
        message += ' [causes: ' + ', '.join(f'{name}={count}' for name, count in report['cause_counts'].items()) + ']'
    report['verdict'] = message
    try:
        (results / 'failures.md').write_text(markdown(report.get('failures', [])))
        (results / 'verdict.json').write_text(json.dumps(dict(verdict=message, status=report['status'],
            failures=report.get('failures', []), cause_counts=report.get('cause_counts', {}),
            baseline_previews=report.get('baseline_previews', [])), indent=2) + '\n')
    except OSError as error:
        report['failure_details_error'] = str(error)
    (results / 'report.json').write_text(json.dumps(report, indent=2) + '\n')
    (results / 'verdict.txt').write_text(message + '\n')
    for error in report['errors']:
        print(error, file=sys.stderr)
    print(message, flush=True)
    return 0 if report['status'] == 'pass' else 1


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('results', type=Path)
    parser.add_argument('--expected', type=Path, default=LANE / 'expected.json')
    parser.add_argument('--sanctioned', type=Path, default=LANE / 'sanctioned-api.json')
    args = parser.parse_args()
    raise SystemExit(write_verdict(args.results, check_results(args.results, args.expected, args.sanctioned)))
