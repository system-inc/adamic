#!/usr/bin/env python3
"""Compare a real native component against a measured full or changed-only Node cache."""
import argparse
import hashlib
import json
from pathlib import Path
import subprocess
import sys
import time

ROOT = Path(__file__).resolve().parent
parser = argparse.ArgumentParser()
parser.add_argument('--timeout', type=float, default=120)
parser.add_argument('cache', type=Path)
parser.add_argument('output', type=Path)
parser.add_argument('command', nargs=argparse.REMAINDER)
args = parser.parse_args()
command = args.command[1:] if args.command[:1] == ['--'] else args.command
if not command:
    parser.error('native command required')
manifest = json.loads((args.cache / 'manifest.json').read_text())
for name, key in [('golden.stdout', 'sha256'), ('request.json', 'requestSha256')]:
    if hashlib.sha256((args.cache / name).read_bytes()).hexdigest() != manifest[key]:
        raise RuntimeError('cache integrity failure: ' + name)
request = json.loads((args.cache / 'request.json').read_text())
if [project['id'] for project in request['projects']] != manifest['selected']:
    raise RuntimeError('selected input population differs')
started = time.perf_counter()
result = subprocess.run([sys.executable, str(ROOT / 'compare.py'), '--timeout', str(args.timeout),
    '--output', str(args.output), str(args.cache / 'golden.stdout'), '--',
    *command, str((args.cache / 'request.json').resolve())])
print(json.dumps({'selected': len(request['projects']), 'seconds': time.perf_counter() - started,
    'comparison_exit': result.returncode}))
raise SystemExit(result.returncode)
