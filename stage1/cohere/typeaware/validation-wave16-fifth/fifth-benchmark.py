import hashlib
import json
import os
from pathlib import Path
import re
import statistics
import subprocess
import time

root = Path('/workspace/adamic')
artifacts = Path('/workspace/wave16-artifacts')
binaries = artifacts / 'fifth-final'
rows = []
for corpus, config in [('compiler', '/workspace/wave16-corpus/typescript/src/compiler/tsconfig.json'), ('repository', str(root / 'tsconfig.json'))]:
    manifest = str(artifacts / (corpus + '.manifest'))
    for round_number in range(3):
        for implementation in (['native', 'go'] if round_number % 2 == 0 else ['go', 'native']):
            binary = binaries / ('wave16' if implementation == 'native' else 'wave16-oracle')
            stem = artifacts / f'fifth-bench-{corpus}-{round_number + 1}-{implementation}'
            environment = dict(os.environ)
            if implementation == 'native':
                environment['ADAMIC_TSGO_TIMING'] = '1'
            with stem.with_suffix('.stdout').open('wb') as output, stem.with_suffix('.stderr').open('wb') as error:
                started = time.perf_counter_ns()
                subprocess.run([str(binary), config, manifest], cwd=root, env=environment, stdout=output, stderr=error, check=True)
                elapsed = time.perf_counter_ns() - started
            data = stem.with_suffix('.stdout').read_bytes()
            phases = {key: int(value) for key, value in re.findall(r'(\w+)=(\d+)', stem.with_suffix('.stderr').read_text())}
            rows.append(dict(corpus=corpus, round=round_number + 1, implementation=implementation, process_ns=elapsed,
                             phases=phases, bytes=len(data), sha256=hashlib.sha256(data).hexdigest()))
    selected = [row for row in rows if row['corpus'] == corpus]
    if len({row['sha256'] for row in selected}) != 1:
        raise RuntimeError(f'{corpus}: timed streams differ')
    for implementation in ['native', 'go']:
        subset = [row for row in selected if row['implementation'] == implementation]
        medians = {key: statistics.median(row['phases'][key] for row in subset) / 1e9 for key in ['load_ns', 'run_ns']}
        medians['process'] = statistics.median(row['process_ns'] for row in subset) / 1e9
        print(corpus, implementation, json.dumps(medians), flush=True)
(artifacts / 'fifth-measurements.json').write_text(json.dumps(rows, indent=2) + '\n')
