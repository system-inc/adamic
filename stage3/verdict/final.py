#!/usr/bin/env python3
"""Require the complete Node verdict before measuring the supplied compiler."""
import argparse
import hashlib
import json
import math
import os
from pathlib import Path
import statistics
import subprocess
import sys
from run import ROOT, write_json

PERFORMANCE = ROOT.parent / 'performance/run.sh'
SUITES = ('acceptance', 'tiny', 'baselines')
MODES = ('native', 'node', 'go-single', 'go-default')
BAR_SECONDS = 1.78


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def verdict_passes(summary, code, manifest):
    if summary.get('harness_errors'):
        raise RuntimeError('verdict harness errors; see correctness/summary.json')
    suites = summary.get('suites', {})
    if set(suites) != set(SUITES):
        raise RuntimeError('incomplete verdict suite population')
    counts = {'acceptance': 301, 'tiny': 1, 'baselines': manifest['configurations']}
    for name, total in counts.items():
        row = suites[name]
        if (row.get('status') == 'harness_error' or row.get('total') != total
                or row.get('deferred', 0) != 0
                or type(row.get('passed')) is not int or type(row.get('failed')) is not int
                or min(row['passed'], row['failed']) < 0
                or row['passed'] + row['failed'] != total
                or len(row.get('failures', [])) != row['failed']):
            raise RuntimeError('incomplete or inconsistent correctness counts: ' + name)
    passed = all(suites[name]['failed'] == 0 for name in SUITES)
    if summary.get('success') is not passed or code != (0 if passed else 1):
        raise RuntimeError('verdict exit and summary disagree')
    return passed


def measured_rows(folder):
    preflight = json.loads((folder / 'preflight.json').read_text())
    if preflight['native_mismatches']:
        raise RuntimeError('native performance preflight differs from Node')
    accepted = preflight['accepted']
    if len(accepted) != len(set(accepted)) or 'typescript-compiler' not in accepted:
        raise RuntimeError('src/compiler must be accepted and timed')
    rows = json.loads((folder / 'results.json').read_text())
    by_input = {}
    for row in rows:
        name, mode = row['input'], row['compiler']
        if name not in accepted or mode not in MODES or mode in by_input.setdefault(name, {}):
            raise RuntimeError('unexpected or duplicated timing population')
        samples = row['seconds']
        if (len(samples) != 10 or any(type(x) not in (int, float) or not math.isfinite(x) or x <= 0 for x in samples)
                or not math.isclose(row['mean'], statistics.mean(samples), rel_tol=1e-9)
                or not math.isclose(row['stdev'], statistics.stdev(samples), rel_tol=1e-9, abs_tol=1e-12)):
            raise RuntimeError('invalid timing samples or statistics')
        by_input[name][mode] = row
    if set(by_input) != set(accepted) or any(set(modes) != set(MODES) for modes in by_input.values()):
        raise RuntimeError('missing native or reference measurements')
    return by_input


def table(summary):
    lines = ['| Phase / input | Pass | Fail | Native s | Node s | Go single s | Go default s | 1.78 s bar |',
             '|---|---:|---:|---:|---:|---:|---:|---|']
    for name in SUITES:
        row = (summary.get('correctness') or {}).get('suites', {}).get(name, {})
        lines.append(f"| correctness: {name} | {row.get('passed', 'error')} | {row.get('failed', 'error')} | | | | | |")
    performance = summary['performance']
    if performance['status'] == 'complete':
        for name, modes in performance['inputs'].items():
            cells = [f"{modes[mode]['mean']:.4f} ± {modes[mode]['stdev']:.4f}" for mode in MODES]
            bar = ('met' if modes['native']['mean'] < BAR_SECONDS else 'missed') if name == 'typescript-compiler' else ''
            lines.append('| timing: ' + name + ' | | | ' + ' | '.join(cells) + ' | ' + bar + ' |')
    else:
        lines.append(f"| timing: {performance['status']} | | | | | | | |")
    return '\n'.join(lines) + '\n'


def execute(binary, output):
    manifest = json.loads((ROOT / 'selection.json').read_text())
    initial_hash = digest(binary)
    summary = {'schema_version': 1, 'tsc': str(binary), 'tsc_sha256': initial_hash,
               'oracle': 'unchanged TypeScript 6.0.3 on Node goldens and upstream diagnostic summaries',
               'correctness': None, 'performance': {'status': 'skipped', 'reason': 'correctness not complete'},
               'compiler_bar_seconds': BAR_SECONDS, 'success': False}
    code = 2
    try:
        argv = [str(ROOT / 'run.sh'), '--tsc', str(binary), str(output / 'correctness')]
        write_json(output / 'correctness-command.json', argv)
        with (output / 'correctness.log').open('wb') as log:
            completed = subprocess.run(argv, stdout=log, stderr=log)
        summary['correctness'] = json.loads((output / 'correctness/summary.json').read_text())
        if not verdict_passes(summary['correctness'], completed.returncode, manifest):
            summary['performance']['reason'] = 'one or more correctness suites failed'
            code = 1
        else:
            if digest(binary) != initial_hash:
                raise RuntimeError('compiler changed after correctness; refusing timing')
            argv = [str(PERFORMANCE), '--native', str(binary), str(output / 'performance')]
            write_json(output / 'performance-command.json', argv)
            summary['performance'] = {'status': 'failed', 'reason': 'performance runner did not complete'}
            with (output / 'performance.log').open('wb') as log:
                completed = subprocess.run(argv, stdout=log, stderr=log)
            if completed.returncode:
                preflight_path = output / 'performance/preflight.json'
                preflight = json.loads(preflight_path.read_text()) if preflight_path.exists() else {}
                summary['performance'].update(exit=completed.returncode, preflight=preflight)
                if preflight.get('native_mismatches'):
                    summary['performance']['reason'] = 'native diagnostics differ in performance preflight'
                code = 1 if preflight.get('native_mismatches') else 2
            else:
                inputs = measured_rows(output / 'performance')
                if digest(binary) != initial_hash:
                    raise RuntimeError('compiler changed during timing')
                native_mean = inputs['typescript-compiler']['native']['mean']
                summary['performance'] = {'status': 'complete', 'inputs': inputs,
                    'compiler_bar_met': native_mean < BAR_SECONDS,
                    'excluded': json.loads((output / 'performance/preflight.json').read_text())['excluded']}
                summary['success'], code = True, 0
    except Exception as error:
        summary['harness_error'] = str(error)
        code = 2
    write_json(output / 'summary.json', summary)
    rendered = table(summary)
    (output / 'summary.md').write_text(rendered)
    print(rendered, end='', flush=True)
    if 'harness_error' in summary:
        print('Harness error: ' + summary['harness_error'], file=sys.stderr)
    elif summary['performance']['status'] != 'complete':
        print('Timing ' + summary['performance']['status'] + ': ' + summary['performance']['reason'], file=sys.stderr)
    return code


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--tsc', type=Path, required=True)
    parser.add_argument('output', type=Path)
    args = parser.parse_args()
    binary, output = args.tsc.resolve(), args.output.resolve()
    if not binary.is_file() or not os.access(binary, os.X_OK):
        parser.error('compiler must be one executable file')
    if output.exists():
        parser.error('output directory must be new')
    output.mkdir(parents=True)
    return execute(binary, output)


if __name__ == '__main__':
    raise SystemExit(main())
