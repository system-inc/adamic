#!/usr/bin/env python3
"""Interleave frozen runtime snapshots. Verify output; save every subprocess log."""
import argparse
import json
import os
from pathlib import Path
import statistics
import subprocess
import time


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('snapshots', type=Path)
    parser.add_argument('css_source', type=Path)
    parser.add_argument('css_corpus', type=Path)
    parser.add_argument('--rounds', type=int, default=7)
    args = parser.parse_args()
    root = args.snapshots.resolve()
    repo = Path(__file__).resolve().parents[4]
    commands = {
        'CSS baseline': [str(root / 'baseline/css'), str(args.css_corpus), 'count'],
        'CSS final': [str(root / 'final/css'), str(args.css_corpus), 'count'],
        'CSS Node': ['node', '--disable-warning=ExperimentalWarning', str(repo / 'oracle/node.mjs'), str(args.css_source), str(args.css_corpus), 'count'],
    }
    for side in ('baseline', 'final'):
        commands['scanner ' + side] = [str(root / side / 'scanner/scanner'), '--manifest', str(root / side / 'scanner/compiler.txt'), '--count']
    commands['scanner Node'] = ['node', '--disable-warning=ExperimentalWarning', str(repo / 'oracle/node.mjs'), str(root / 'final/scanner/main.ts'), '--manifest', str(root / 'final/scanner/compiler.txt'), '--count']
    samples = {name: [] for name in commands}
    answers = {}
    for round_number in range(args.rounds):
        names = list(commands)
        if round_number % 2:
            names.reverse()
        for name in names:
            stem = root / ('round-' + str(round_number + 1) + '-' + name.replace(' ', '-'))
            with Path(str(stem) + '.stdout').open('wb') as out, Path(str(stem) + '.stderr').open('wb') as err:
                started = time.perf_counter()
                result = subprocess.run(commands[name], stdout=out, stderr=err, check=True)
                duration = time.perf_counter() - started
            answer = Path(str(stem) + '.stdout').read_bytes()
            error = Path(str(stem) + '.stderr').read_bytes()
            workload = name.split()[0]
            if error or answer != answers.setdefault(workload, answer):
                raise RuntimeError(name + ' checksum/stderr differs')
            samples[name].append(duration)
    units = {'CSS': int(answers['CSS'].split()[0]), 'scanner': int(answers['scanner'])}
    report = {
        'rounds_seconds': samples,
        'answers': {name: answer.decode() for name, answer in answers.items()},
        'median_per_second': {name: units[name.split()[0]] / statistics.median(values) for name, values in samples.items()},
        'commands': commands,
        'loadavg_after': Path('/proc/loadavg').read_text().strip(),
        'cpu_count': os.cpu_count(),
    }
    (root / 'measurements.json').write_text(json.dumps(report, indent=2) + '\n')
    print(json.dumps(report, indent=2))


if __name__ == '__main__':
    main()
