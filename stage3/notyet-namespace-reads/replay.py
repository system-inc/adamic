#!/usr/bin/env python3
"""Replay the five exact census sites with the complete, hash-verified project."""
import argparse
import hashlib
import json
from pathlib import Path
import subprocess
import time

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('adapted', type=Path)
parser.add_argument('output', type=Path)
parser.add_argument('--worker', type=Path, help='reuse a census-overlay replay worker')
parser.add_argument('--expect-before', action='store_true')
args = parser.parse_args()
repository = Path(__file__).resolve().parents[2]
manifest = repository / 'stage3/meter/runs/20261008T035244Z.latent-full/tsc/source-manifest.json'
verified = []
for row in json.loads(manifest.read_text()):
    path = args.adapted / row['file']
    actual = hashlib.sha256(path.read_bytes()).hexdigest()
    assert actual == row['sha256'], (path, actual, row['sha256'])
    verified.append(row)
args.output.mkdir(parents=True, exist_ok=True)
(args.output / 'verified-sources.json').write_text(json.dumps(verified, indent=2) + '\n')
sites = [
    ('debug-core', 'core.ts:108:5', 'core.ts:106:1', 'reading Debug'),
    ('debug-binder', 'binder.ts:582:9', 'binder.ts:571:5', 'reading Debug'),
    ('debug-relative', 'core.ts:878:13', 'core.ts:871:1', 'reading Debug'),
    ('performance-binder', 'binder.ts:503:5', 'binder.ts:502:1', 'reading performance'),
    ('performance-tracing', 'tracing.ts:191:9', 'tracing.ts:187:5', 'reading performance'),
]
rows = []
for name, position, selected, reason in sites:
    prefix = str(args.adapted.resolve() / 'src/compiler') + '/'
    command = ([str(args.worker.resolve())] if args.worker else ['go', 'run', './stage3/census/latent/replay'])
    command += ['-project', str(args.adapted.resolve() / 'src/tsc/tsc.ts'), '-where', prefix + position, '-kind', 'NotYet', '-reason', reason]
    started = time.monotonic()
    with (args.output / (name + '.json')).open('w') as stdout, (args.output / (name + '.log.txt')).open('w') as stderr:
        result = subprocess.run(command, cwd=repository, stdout=stdout, stderr=stderr)
    record = json.loads((args.output / (name + '.json')).read_text())
    assert len(record['units']) == 1 and record['units'][0]['where'] == prefix + selected
    assert record['units'][0]['status'] == 'attempted'
    matches = [x for x in record['findings'] if x['kind'] == 'NotYet' and x['phase'] == 'lowering' and x['where'] == prefix + position and x['reason'] == reason]
    expected = args.expect_before and reason == 'reading performance'
    assert bool(matches) == expected and result.returncode == (0 if expected else 1), (name, result.returncode, matches)
    rows.append({'name': name, 'argv': command, 'exit': result.returncode, 'seconds': time.monotonic() - started, 'reproduced': bool(matches)})
    print(name + ': ' + ('reproduced' if matches else 'signature absent'), flush=True)
(args.output / 'runs.json').write_text(json.dumps(rows, indent=2) + '\n')
