#!/usr/bin/env python3
"""Verify full service bytes and collect instruction profiles of production-built executables."""
import gzip
import json
import os
from pathlib import Path
import subprocess
import sys

s = Path(sys.argv[1]).resolve()
tools = Path(sys.argv[2]).resolve()
os.environ['VALGRIND_LIB'] = str(tools / 'valgrind/usr/libexec/valgrind')
results = {}
for mode in ['baseline', 'thin']:
    with (s / (mode + '-service-verify.stdout')).open('wb') as out, (s / (mode + '-service-verify.stderr')).open('wb') as err:
        subprocess.run([str(s / mode / 'service'), '/tmp/wasm-requests-profile/requests.jsonl', 'verify'], stdout=out, stderr=err, check=True)
    expected = Path('/tmp/wasm-requests-profile/responses.jsonl').read_bytes() + b'7394547\n'
    if (s / (mode + '-service-verify.stdout')).read_bytes() != expected:
        raise RuntimeError('MISCOMPILE: full service bytes differ from Node: ' + mode)
print('full service Node byte parity PASS', flush=True)
for workload, arguments in [('parse', ['--manifest', str(s / 'compiler.txt'), '--count']), ('service', ['/tmp/wasm-requests-profile/requests.jsonl', 'run'])]:
    results[workload] = {}
    for mode in ['baseline', 'thin']:
        path = s / (mode + '-' + workload + '.callgrind')
        with (s / (mode + '-' + workload + '-profile.stdout')).open('wb') as out, (s / (mode + '-' + workload + '-profile.stderr')).open('wb') as err:
            subprocess.run(['taskset', '-c', '3', str(tools / 'valgrind/usr/bin/valgrind'), '--tool=callgrind', '--callgrind-out-file=' + str(path), str(s / mode / workload), *arguments], stdout=out, stderr=err, check=True)
        expected = b'0\n' if workload == 'parse' else b'7394547\n'
        if (s / (mode + '-' + workload + '-profile.stdout')).read_bytes() != expected:
            raise RuntimeError('MISCOMPILE: instruction-run output differs')
        summaries = [line for line in path.read_text().splitlines() if line.startswith('summary:')]
        if len(summaries) != 1:
            raise RuntimeError('missing instruction summary')
        results[workload][mode] = int(summaries[0].split()[1])
        with gzip.open(str(path) + '.gz', 'wb') as z:
            z.write(path.read_bytes())
        print(workload, mode, results[workload][mode], flush=True)
        (s / 'shipped-instructions.json').write_text(json.dumps(results, indent=2) + '\n')
