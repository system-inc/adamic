#!/usr/bin/env python3
"""Retain a warm batch and time the harness's serve markers using file-backed stderr."""
import json
import os
from pathlib import Path
import subprocess
import sys
import time

scratch = Path(sys.argv[1]).resolve()
rounds = []
for index in range(5):
    row = {'round': index + 1, 'order': ['baseline', 'thin'] if index % 2 == 0 else ['thin', 'baseline']}
    for mode in row['order']:
        stem = scratch / f'{mode}-service-warm-round-{index + 1}'
        out_path, err_path = Path(str(stem) + '.stdout'), Path(str(stem) + '.stderr')
        before = os.getloadavg()
        whole_start = time.perf_counter()
        with out_path.open('wb') as out, err_path.open('wb') as err:
            process = subprocess.Popen([str(scratch / mode / 'service'), '/tmp/wasm-requests-profile/requests.jsonl', 'warm'], stdout=out, stderr=err)
            start = stop = None
            while True:
                data = err_path.read_bytes()
                now = time.perf_counter()
                if start is None and b'serve:start\n' in data:
                    start = now
                if stop is None and b'serve:stop\n' in data:
                    stop = now
                if process.poll() is not None:
                    break
                time.sleep(0.001)
            if process.returncode != 0 or start is None or stop is None or stop <= start:
                raise RuntimeError('service marker timing failed')
        if out_path.read_bytes() != b'7394547\n' or err_path.read_bytes() != b'serve:start\nserve:stop\n':
            raise RuntimeError('MISCOMPILE: warmed service output differs')
        row[mode] = {'seconds': stop - start, 'whole_seconds': time.perf_counter() - whole_start, 'load_before': before, 'load_after': os.getloadavg()}
    rounds.append(row)
    print(json.dumps(row), flush=True)
result = {'rounds': rounds, 'best_seconds': {mode: min(row[mode]['seconds'] for row in rounds) for mode in ['baseline', 'thin']}}
(scratch / 'warm-service-results.json').write_text(json.dumps(result, indent=2) + '\n')
print(json.dumps(result['best_seconds']), flush=True)
