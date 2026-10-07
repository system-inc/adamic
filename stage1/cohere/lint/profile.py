#!/usr/bin/env python3
"""Measure lint artifact snapshots. Program output always goes to files."""
import argparse
import importlib.util
import json
import os
from pathlib import Path
import platform
import subprocess
import sys
import time

sys.dont_write_bytecode = True
REPO = Path(__file__).resolve().parents[3]
spec = importlib.util.spec_from_file_location('scanner_profile', REPO / 'stage1/typescript/scanner/profile.py')
accounting = importlib.util.module_from_spec(spec)
spec.loader.exec_module(accounting)

def run(command, directory, stem, env=None):
    with (directory / (stem + '.stdout')).open('wb') as out, (directory / (stem + '.stderr')).open('wb') as err:
        started = time.perf_counter()
        subprocess.run(command, stdout=out, stderr=err, env=env, check=True)
        elapsed = time.perf_counter() - started
    return elapsed, (directory / (stem + '.stdout')).read_bytes()

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('directory', type=Path)
    parser.add_argument('--valgrind', default='valgrind')
    parser.add_argument('--summarize-only', action='store_true')
    parser.add_argument('--existing-profile', action='store_true')
    args = parser.parse_args()
    directory = args.directory.resolve()
    measurements = {}
    if not args.summarize_only:
        manifest = ['--manifest', str(directory / 'compiler.txt'), '--count']
        commands = dict(native=[str(directory / 'scanner'), *manifest],
                        Go=[str(directory / 'oracle'), *manifest],
                        Node=['node', '--disable-warning=ExperimentalWarning', str(REPO / 'oracle/node.mjs'), str(directory / 'main.ts'), *manifest])
        samples = {name: [] for name in commands}
        want = None
        load_before = Path('/proc/loadavg').read_text().strip()
        for round_number in range(1, 6):
            for name, command in commands.items():
                elapsed, answer = run(command, directory, f'round-{round_number}-{name}')
                if want is None:
                    want = answer
                if answer != want or (directory / f'round-{round_number}-{name}.stderr').stat().st_size:
                    raise RuntimeError(f'{name} count/stderr differs')
                samples[name].append(elapsed)
        findings = int(want)
        measurements = dict(findings=findings, rounds=samples,
                            best={name: dict(seconds=min(values), findings_per_second=findings / min(values)) for name, values in samples.items()},
                            machine=platform.uname()._asdict(), cpu=next(line for line in Path('/proc/cpuinfo').read_text().splitlines() if line.startswith('model name')),
                            nproc=os.cpu_count(), cpu_max=Path('/sys/fs/cgroup/cpu.max').read_text().strip(),
                            load_before=load_before, load_after=Path('/proc/loadavg').read_text().strip())
        _, answer = run([str(directory / 'counted'), *manifest], directory, 'counted')
        if answer != want:
            raise RuntimeError('counted build count differs')
        measurements['allocation_counts'] = (directory / 'counted.stderr').read_text().strip()
        if not args.existing_profile:
            _, answer = run([args.valgrind, '--tool=callgrind', '--callgrind-out-file=' + str(directory / 'callgrind.out'), str(directory / 'profiled'), *manifest], directory, 'callgrind')
        else:
            answer = (directory / 'callgrind.stdout').read_bytes()
        if answer != want:
            raise RuntimeError('profiled build count differs')
    result = accounting.summarize(directory / 'callgrind.out')
    measurements['instructions'] = result['total']
    if 'findings' in measurements:
        measurements['instructions_per_finding'] = result['total'] / measurements['findings']
        (directory / 'measurements.json').write_text(json.dumps(measurements, indent=2) + '\n')
    for mode in ('inclusive', 'self'):
        print(f'\nTop 20 {mode} instructions (same-name inline records collapsed):')
        for row in result[mode][:20]:
            print(f'{row["instructions"]:>14,} {row["percent"]:6.2f}% {row["function"]}')
    print('\n' + json.dumps(measurements, indent=2))

if __name__ == '__main__':
    main()
