"""Interleaved, uninstrumented wave-01 continuation timings with full output agreement."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import time

parser = argparse.ArgumentParser()
parser.add_argument('--artifacts', type=Path, required=True)
parser.add_argument('--typescript', type=Path, required=True)
parser.add_argument('--repository', type=Path, required=True)
parser.add_argument('--output', type=Path, required=True)
args = parser.parse_args()
args.output.mkdir(parents=True, exist_ok=True)
records = []
configs = [('compiler', args.typescript / 'src/compiler/tsconfig.json'),
           ('repository', args.repository / 'tsconfig.json')]
for corpus, config in configs:
    for round_number in range(1, 4):
        for mode, binary in [('go', 'oracle'), ('native', 'native')]:
            stem = args.output / f'{corpus}-{mode}-{round_number}'
            started = time.perf_counter_ns()
            with Path(str(stem) + '.stdout').open('wb') as stdout, Path(str(stem) + '.timing').open('wb') as stderr:
                subprocess.run([str(args.artifacts / binary), str(config),
                                str(args.artifacts / (corpus + '.manifest'))],
                               stdout=stdout, stderr=stderr, check=True,
                               env={**os.environ, 'ADAMIC_TSGO_TIMING': '1'})
            wall = time.perf_counter_ns() - started
            data = Path(str(stem) + '.stdout').read_bytes()
            fields = {}
            for item in Path(str(stem) + '.timing').read_text().split():
                if '=' in item:
                    key, value = item.split('=')
                    fields[key] = int(value)
            records.append({'corpus': corpus, 'round': round_number, 'mode': mode,
                            'wall_ns': wall, **fields, 'bytes': len(data),
                            'sha256': hashlib.sha256(data).hexdigest()})
        go = args.output / f'{corpus}-go-{round_number}.stdout'
        native = args.output / f'{corpus}-native-{round_number}.stdout'
        if go.read_bytes() != native.read_bytes():
            raise RuntimeError(f'{corpus} timed outputs differ')
(args.output / 'measurements.json').write_text(json.dumps(records, indent=2) + '\n')
