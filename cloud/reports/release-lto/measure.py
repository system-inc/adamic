#!/usr/bin/env python3
"""Alternate five release runs, compare every byte, then collect whole-process Ir."""
import gzip
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import time

scratch = Path(sys.argv[1]).resolve()
results = {}

def same(left, right):
    if left.read_bytes() != right.read_bytes():
        raise RuntimeError(f'MISCOMPILE: output differs: {left} {right}')

def run(arguments, stem):
    before = os.getloadavg()
    start = time.perf_counter()
    with (scratch / (stem + '.stdout')).open('wb') as stdout, (scratch / (stem + '.stderr')).open('wb') as stderr:
        subprocess.run(list(map(str, arguments)), stdout=stdout, stderr=stderr, check=True)
    return {'seconds': time.perf_counter() - start, 'load_before': before, 'load_after': os.getloadavg()}

workloads = {'parse': ['--manifest', scratch / 'compiler.txt', '--count'], 'service': [Path('/tmp/wasm-requests-profile/requests.jsonl'), 'run']}
# Full response verification is outside the timed interval.
for mode in ['baseline', 'thin']:
    run([scratch / mode / 'service', '/tmp/wasm-requests-profile/requests.jsonl', 'verify'], mode + '-service-verify')
    actual = scratch / (mode + '-service-verify.stdout')
    expected = Path('/tmp/wasm-requests-profile/responses.jsonl').read_bytes() + b'7394547\n'
    if actual.read_bytes() != expected:
        raise RuntimeError('MISCOMPILE: service response bytes differ from Node: ' + mode)
print('service full Node byte parity PASS', flush=True)
for workload, arguments in workloads.items():
    for mode in ['baseline', 'thin']:
        run([scratch / mode / workload, *arguments], mode + '-' + workload + '-warmup')
    same(scratch / ('baseline-' + workload + '-warmup.stdout'), scratch / ('thin-' + workload + '-warmup.stdout'))
    same(scratch / ('baseline-' + workload + '-warmup.stderr'), scratch / ('thin-' + workload + '-warmup.stderr'))
    rounds = []
    for round_index in range(5):
        order = ['baseline', 'thin'] if round_index % 2 == 0 else ['thin', 'baseline']
        row = {'round': round_index + 1, 'order': order}
        for mode in order:
            stem = f'{mode}-{workload}-round-{round_index + 1}'
            row[mode] = run([scratch / mode / workload, *arguments], stem)
            same(scratch / (stem + '.stdout'), scratch / ('baseline-' + workload + '-warmup.stdout'))
            same(scratch / (stem + '.stderr'), scratch / ('baseline-' + workload + '-warmup.stderr'))
        rounds.append(row)
        print(workload, json.dumps(row), flush=True)
    results[workload] = {'rounds': rounds, 'best_seconds': {mode: min(row[mode]['seconds'] for row in rounds) for mode in ['baseline', 'thin']}}
    (scratch / 'timing-results.json').write_text(json.dumps(results, indent=2) + '\n')
# These serial profiles start only after all wall measurements finish.
valgrind = scratch / 'valgrind/usr/bin/valgrind'
os.environ['VALGRIND_LIB'] = str(scratch / 'valgrind/usr/libexec/valgrind')
for workload, arguments in workloads.items():
    for mode in ['baseline', 'thin']:
        profile = scratch / (mode + '-' + workload + '.callgrind')
        run([valgrind, '--tool=callgrind', '--callgrind-out-file=' + str(profile), scratch / mode / workload, *arguments], mode + '-' + workload + '-profile')
        same(scratch / (mode + '-' + workload + '-profile.stdout'), scratch / ('baseline-' + workload + '-warmup.stdout'))
        summary = [line for line in profile.read_text().splitlines() if line.startswith('summary:')]
        if len(summary) != 1:
            raise RuntimeError('missing Callgrind summary')
        results[workload].setdefault('instructions', {})[mode] = int(summary[0].split()[1])
        with gzip.open(str(profile) + '.gz', 'wb') as output:
            output.write(profile.read_bytes())
        (scratch / 'measurement-results.json').write_text(json.dumps(results, indent=2) + '\n')
        print(workload, mode, summary[0], flush=True)
print('all measured output comparisons PASS', flush=True)
