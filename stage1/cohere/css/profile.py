#!/usr/bin/env python3
"""Checked Callgrind accounting and five interleaved full-corpus CSS timings.

The scanner's parser collapses same-name DWARF records and requires all self
costs to add up to Callgrind's summary. Inclusive costs overlap, never sum them.
"""
import argparse
import importlib.util
import json
import os
from pathlib import Path
import re
import resource
import signal
import statistics
import subprocess
import time

ROOT = Path(__file__).resolve().parents[3]
SPEC = importlib.util.spec_from_file_location(
    'scanner_profile', ROOT / 'stage1/typescript/scanner/profile.py')
SCANNER = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(SCANNER)


CPU_SECONDS = 120


def limit_cpu():
    # CPU time keeps this hang probe independent of scheduling delays on loaded boxes.
    resource.setrlimit(resource.RLIMIT_CPU, (CPU_SECONDS, CPU_SECONDS + 1))


def run_bounded(command, environment):
    result = subprocess.run(command, capture_output=True, env=environment,
                            preexec_fn=limit_cpu)
    if result.returncode == -signal.SIGXCPU:
        raise RuntimeError(f'stalled: exceeded {CPU_SECONDS}s of CPU time ({command[0]})')
    return result


def summarize(path):
    return SCANNER.summarize(Path(path))


def benchmark(baseline, final, prettier, output):
    baseline, final = Path(baseline).resolve(), Path(final).resolve()
    cases = final / 'shared.txt'
    if cases.read_bytes() != (baseline / 'shared.txt').read_bytes():
        raise ValueError('baseline and final corpus differ')
    want = b'4952 of 4952 stylesheets formatted, 510298 units\n'
    runner = ROOT / 'oracle/node.mjs'
    script = ROOT / 'stage1/cohere/css/testdata/print_library.mjs'
    sides = [
        ('native baseline', [str(baseline / 'printer'), str(cases), 'count', 'once']),
        ('native final', [str(final / 'printer'), str(cases), 'count', 'once']),
        ('Node baseline', ['node', '--disable-warning=ExperimentalWarning', str(runner), str(baseline / 'source/css/print_main.ts'), str(cases), 'count', 'once']),
        ('Node final', ['node', '--disable-warning=ExperimentalWarning', str(runner), str(final / 'source/css/print_main.ts'), str(cases), 'count', 'once']),
        ('Prettier fork', ['node', str(script), str(ROOT / 'cohere/internal/format/prettier/bundles'), str(cases), 'default', 'fork', 'count', 'once']),
        ('Prettier npm 3.9.6', ['node', str(script), prettier, str(cases), 'default', 'npm', 'count', 'once']),
        ('Go', [str(final / 'go-printer'), '-test.v', '-test.run=^TestAdamicPrinterThroughput$']),
    ]
    observations = []
    for round_index in range(5):
        for name, command in (sides if round_index % 2 == 0 else list(reversed(sides))):
            environment = dict(os.environ, ADAMIC_PORT_REQUEST=str(final / 'go-request.json'))
            started = time.perf_counter()
            result = run_bounded(command, environment)
            elapsed = time.perf_counter() - started
            if result.returncode or result.stderr:
                raise ValueError((name, result.returncode, result.stderr.decode()))
            if name == 'Go':
                match = re.search(rb'Go round 1: 4952 stylesheets in ([\d.]+)(ms|s), (\d+) stylesheets/s; 510298 units', result.stdout)
                if not match:
                    raise ValueError((name, result.stdout.decode()))
                elapsed = float(match[1]) * (0.001 if match[2] == b'ms' else 1)
            elif result.stdout != want:
                raise ValueError((name, result.stdout.decode()))
            rate = 4952 / elapsed
            observation = dict(side=name, round=round_index + 1, seconds=elapsed, stylesheets_per_second=rate, stdout=result.stdout.decode())
            observations.append(observation)
            print(json.dumps(observation), flush=True)
    medians = {name: statistics.median(o['stylesheets_per_second'] for o in observations if o['side'] == name) for name, _ in sides}
    payload = dict(stylesheets=4952, output_units=510298, rounds=observations, median_stylesheets_per_second=medians)
    Path(output).write_text(json.dumps(payload, indent=2) + '\n')
    print(json.dumps(medians, indent=2), flush=True)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--summarize')
    parser.add_argument('--benchmark', nargs=2, metavar=('BASELINE', 'FINAL'))
    parser.add_argument('--prettier', default=os.environ.get('ADAMIC_CSS_PRINTER_LIBRARY'))
    parser.add_argument('--output')
    args = parser.parse_args()
    if args.summarize:
        print(json.dumps(summarize(args.summarize), indent=2))
    elif args.benchmark and args.prettier and args.output:
        benchmark(*args.benchmark, args.prettier, args.output)
    else:
        parser.error('use --summarize FILE or --benchmark BASELINE FINAL --prettier DIR --output FILE')
