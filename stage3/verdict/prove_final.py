#!/usr/bin/env python3
"""Check completed A/B/C final-command artifacts and exact timing exclusion."""
import argparse
import json
from pathlib import Path
from final import MODES, SUITES, measured_rows
from run import ROOT, write_json


def prove(paths, output):
    output.mkdir(parents=True)
    summaries = {label: json.loads((folder / 'summary.json').read_text()) for label, folder in paths.items()}
    a, b, c = [summaries[label] for label in ('A', 'B', 'C')]
    for label, summary in summaries.items():
        if summary.get('harness_error') or summary['correctness']['harness_errors']:
            raise RuntimeError(label + ': harness error is not a correctness observation')
    if not a['success'] or a['performance']['status'] != 'complete':
        raise RuntimeError('A must pass and finish real timing')
    measurements = measured_rows(paths['A'] / 'performance')
    if measurements != a['performance']['inputs']:
        raise RuntimeError('A final and performance timing artifacts differ')
    for label in ('B', 'C'):
        summary = summaries[label]
        if summary['success'] or summary['performance']['status'] != 'skipped':
            raise RuntimeError(label + ': failed correctness reached performance')
        folder = paths[label]
        if any((folder / name).exists() for name in ('performance', 'performance-command.json', 'performance.log')):
            raise RuntimeError(label + ': performance invoked despite rejection')
    for name in SUITES:
        original, mutant, empty = [summaries[label]['correctness']['suites'][name] for label in ('A', 'B', 'C')]
        if original['failed'] or original.get('deferred', 0):
            raise RuntimeError(name + ': A did not pass the full population')
        if mutant['total'] != original['total'] or not mutant['failed']:
            raise RuntimeError(name + ': B failed to introduce a diagnostic difference')
        if empty['total'] != original['total'] or empty['passed'] or empty['failed'] != empty['total']:
            raise RuntimeError(name + ': C did not fail every case')
        for failure in mutant['failures']:
            if set(failure['differences']) != {'stdout'}:
                raise RuntimeError(name + ': B was caught by a different check')
            suffix = failure.get('capture', failure['case'])
            left = paths['A'] / 'correctness' / name / suffix
            right = paths['B'] / 'correctness' / name / suffix
            # Baseline raw paths contain different scratch roots. Projection is
            # the compared stream, and the wrapper is proved separately below.
            stream = 'actual.diagnostics' if name == 'baselines' else 'actual.stdout'
            before, after = [(p / stream).read_bytes() for p in (left, right)]
            if len(before) != len(after) or sum(x != y for x, y in zip(before, after)) != 1:
                raise RuntimeError(name + ': B changed more than one compared byte')
            raw_before, raw_after = [(p / 'actual.stdout').read_bytes() for p in (left, right)]
            if len(raw_before) != len(raw_after) or sum(x != y for x, y in zip(raw_before, raw_after)) != 1:
                raise RuntimeError(name + ': B changed more than one raw diagnostic byte')
            for stream in ('stderr', 'exit'):
                if (left / ('actual.' + stream)).read_bytes() != (right / ('actual.' + stream)).read_bytes():
                    raise RuntimeError(name + ': B changed ' + stream)
    for name in measurements:
        folder = paths['A'] / 'performance/raw' / name
        for mode in MODES:
            if len(list(folder.glob(mode + '-run-*.command.json'))) != 10:
                raise RuntimeError('timing invocation count changed: ' + name + '/' + mode)
            for command in folder.glob(mode + '-run-*.command.json'):
                argv = json.loads(command.read_text())['argv']
                if '--noEmit' not in argv:
                    raise RuntimeError('performance fixture did not invoke a checker')
    proof = {'proved': True, 'runs': {label: str(folder) for label, folder in paths.items()},
             'A_passes': {name: a['correctness']['suites'][name]['passed'] for name in SUITES},
             'B_failures': {name: b['correctness']['suites'][name]['failed'] for name in SUITES},
             'C_failures': {name: c['correctness']['suites'][name]['failed'] for name in SUITES},
             'A_timed_inputs': list(measurements), 'A_native_timed_processes': 10 * len(measurements),
             'B_performance_invoked': False, 'C_performance_invoked': False,
             'catch': 'one diagnostic byte in every B failure; no performance command, directory or log for B/C'}
    write_json(output / 'proof.json', proof)
    print(json.dumps(proof, indent=2))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('A', type=Path)
    parser.add_argument('B', type=Path)
    parser.add_argument('C', type=Path)
    parser.add_argument('output', type=Path)
    args = parser.parse_args()
    prove({label: getattr(args, label).resolve() for label in ('A', 'B', 'C')}, args.output.resolve())


if __name__ == '__main__':
    main()
