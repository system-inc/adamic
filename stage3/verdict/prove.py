#!/usr/bin/env python3
"""Measure A, B and C, then require the mutant and empty binary to fail."""
import argparse
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys

ROOT = Path(__file__).resolve().parent


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--baseline-limit', type=int)
    parser.add_argument('output', type=Path)
    args = parser.parse_args()
    output = args.output.resolve()
    output.mkdir(parents=True)
    measured = {}
    for label, standin in [('A', 'node.sh'), ('B', 'mutant.sh'), ('C', 'empty.sh')]:
        argv = [str(ROOT / 'run.sh'), '--tsc', str(ROOT / 'standins' / standin)]
        if args.baseline_limit:
            argv += ['--baseline-limit', str(args.baseline_limit)]
        argv.append(str(output / label))
        env = dict(os.environ, STAGE3_VERDICT_MUTANT_TARGET='argument.ts,ArrowFunctionExpression1.ts')
        with (output / (label + '.log')).open('wb') as log:
            completed = subprocess.run(argv, env=env, stdout=log, stderr=log)
        summary = json.loads((output / label / 'summary.json').read_text())
        if summary['harness_errors']:
            raise RuntimeError(f'{label} harness errors: {summary["harness_errors"]}')
        if completed.returncode != (0 if summary['success'] else 1):
            raise RuntimeError('summary and exit disagree')
        measured[label] = summary
        print(label, {name: {key: row[key] for key in ('total', 'passed', 'failed', 'excluded')}
                      for name, row in summary['suites'].items()}, flush=True)
    for name, original in measured['A']['suites'].items():
        mutant = measured['B']['suites'][name]
        empty = measured['C']['suites'][name]
        if original['passed'] == 0 or mutant['failed'] <= original['failed']:
            raise RuntimeError(f'{name}: diagnostic mutant did not introduce a failure')
        introduced = {row['case'] for row in mutant['failures']} - {row['case'] for row in original['failures']}
        for row in mutant['failures']:
            if row['case'] in introduced:
                if set(row['differences']) != {'stdout'}:
                    raise RuntimeError(f'{name}: mutant caught by a different comparison')
                difference = row['differences']['stdout']
                if difference['actual_size'] != difference['expected_size']:
                    raise RuntimeError(f'{name}: mutant changed output length')
        if empty['passed'] != 0 or empty['failed'] != empty['total']:
            raise RuntimeError(f'{name}: empty-output exit-1 binary was accepted')
        # Retain the exact changed bytes for each introduced failure.
        for row in mutant['failures']:
            if row['case'] not in introduced:
                continue
            folder = output / 'B' / name / row.get('capture', row['case'])
            source = output / 'A' / name / row.get('capture', row['case'])
            before = (source / 'actual.stdout').read_bytes()
            after = (folder / 'actual.stdout').read_bytes()
            if len(before) != len(after) or sum(a != b for a, b in zip(before, after)) != 1:
                raise RuntimeError('stand-in B did not change exactly one byte')
    evidence = output / 'evidence'
    evidence.mkdir()
    for label in measured:
        for filename in ('summary.json', 'summary.md'):
            shutil.copyfile(output / label / filename, evidence / (label + '-' + filename))
        shutil.copyfile(output / (label + '.log'), evidence / (label + '.log'))
    (evidence / 'proof.json').write_text(json.dumps({'proved': True, 'baseline_limit': args.baseline_limit,
        'mutant_targets': ['argument.ts', 'ArrowFunctionExpression1.ts'],
        'catch': 'one diagnostic byte, stdout only; C passes zero in every suite'}, indent=2) + '\n')
    print('all stand-in assertions passed', flush=True)


if __name__ == '__main__':
    main()
