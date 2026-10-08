"""Measure one production checker program reused across a complete manifest."""
import argparse
import json
import os
from pathlib import Path
import re
import subprocess
import time

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('native')
parser.add_argument('oracle')
parser.add_argument('config')
parser.add_argument('manifest')
parser.add_argument('directory', type=Path)
parser.add_argument('--rounds', type=int, default=3)
parser.add_argument('--repetitions', type=int, default=1)
parser.add_argument('--profile', action='store_true')
args = parser.parse_args()
if args.rounds < 1 or args.repetitions < 1:
    parser.error('positive rounds and repetitions required')
args.directory.mkdir()
records = []
for round_number in range(1, args.rounds + 1):
    order = ['go', 'native'] if round_number % 2 else ['native', 'go']
    outputs = {}
    for implementation in order:
        stem = args.directory / f'{round_number}-{implementation}'
        command = [args.oracle if implementation == 'go' else args.native,
                   args.config, '--manifest', args.manifest]
        environment = dict(os.environ, ADAMIC_TSGO_TIMING='1')
        if implementation == 'native':
            command.append(str(args.repetitions))
            if args.profile:
                environment['ADAMIC_TSGO_PROFILE'] = str(stem.with_suffix('.pprof').resolve())
        started = time.perf_counter_ns()
        result = subprocess.run(command, env=environment, capture_output=True, check=False)
        elapsed = time.perf_counter_ns() - started
        stem.with_suffix('.stdout').write_bytes(result.stdout)
        stem.with_suffix('.stderr').write_bytes(result.stderr)
        if result.returncode:
            raise RuntimeError(f'{stem}: exit {result.returncode}; see stderr')
        outputs[implementation] = result.stdout
        record = dict(implementation=implementation, round=round_number, process_ns=elapsed,
                      repetitions=args.repetitions if implementation == 'native' else 1)
        record.update({k: int(v) for k, v in re.findall(rb'([a-z_]+)=(\d+)', result.stderr)})
        record = {k.decode() if isinstance(k, bytes) else k: v for k, v in record.items()}
        records.append(record)
    if outputs['go'] != outputs['native']:
        raise RuntimeError('whole-manifest finding/fix bytes differ')
(args.directory / 'results.json').write_text(json.dumps(records, indent=2) + '\n')
