#!/usr/bin/env python3
"""Serial three-sample count-mode timing on the already compared corpora."""
from pathlib import Path
import hashlib
import json
import statistics
import subprocess
import sys
import time

REPOSITORY = Path(__file__).resolve().parents[4]
ARTIFACTS = Path(sys.argv[1]).resolve()
OUT = Path(__file__).resolve().parent / 'timing'
OUT.mkdir(exist_ok=True)
rows = []
for population, config, manifest in [
    ('compiler', '/workspace/wave16-corpus/typescript/src/compiler/tsconfig.json',
     '/workspace/wave16-artifacts/compiler.manifest'),
    ('repository', str(REPOSITORY / 'tsconfig.json'),
     '/workspace/wave16-artifacts/repository.manifest'),
]:
    for group in ['first', 'followup', 'third', 'fourth', 'fifth']:
        samples = {'native': [], 'go': []}
        hashes = {}
        for repeat in range(3):
            streams = {}
            # Alternate execution order to reduce systematic warm-cache bias.
            order = ['native', 'go'] if repeat % 2 == 0 else ['go', 'native']
            for mode in order:
                binary = ARTIFACTS / group / ('wave16' if mode == 'native' else 'wave16-oracle')
                hashes[mode] = hashlib.sha256(binary.read_bytes()).hexdigest()
                command = [str(binary), config, manifest, '--count']
                prefix = OUT / f'{population}-{group}-{repeat}-{mode}'
                with prefix.with_suffix('.stdout').open('wb') as stdout, \
                     prefix.with_suffix('.stderr').open('wb') as stderr:
                    start = time.perf_counter()
                    result = subprocess.run(command, stdout=stdout, stderr=stderr)
                    seconds = time.perf_counter() - start
                if result.returncode:
                    raise AssertionError(f'{command}: exit {result.returncode}')
                if mode == 'native' and prefix.with_suffix('.stderr').stat().st_size:
                    raise AssertionError(f'{command}: native stderr')
                streams[mode] = prefix.with_suffix('.stdout').read_bytes()
                samples[mode].append(seconds)
            if streams['native'] != streams['go']:
                raise AssertionError(f'{population} {group}: count mismatch')
        native = statistics.median(samples['native'])
        go = statistics.median(samples['go'])
        rows.append({'population': population, 'group': group, 'samples': samples,
                     'native_median_seconds': native, 'go_median_seconds': go,
                     'native_over_go': native / go, 'binary_sha256': hashes,
                     'config': config, 'manifest': manifest})
        print(f'{population} {group}: native {native:.6f}s, Go {go:.6f}s, '
              f'{native / go:.2f}x; all three count bytes match', flush=True)
(OUT / 'results.json').write_text(json.dumps(rows, indent=2) + '\n')
print('PASS 60 successful processes, 30 matching native/Go count pairs')
