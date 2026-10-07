#!/usr/bin/env python3
"""Replay the bounded evidence through scratch Go overlays; shared files stay intact."""
import argparse
import json
import os
from pathlib import Path
import shutil
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument('--compiler', required=True, help='TypeScript v6.0.3 checkout')
parser.add_argument('--scratch', default='/tmp/lint-wave1-13-loops-replay')
args = parser.parse_args()
evidence = Path(__file__).resolve().parent
repository = evidence.parents[5]
scratch = Path(args.scratch).resolve()
scratch.mkdir(parents=True, exist_ok=True)
replacements = {}
for name, relative in [
    ('registry.go', 'stage1/cohere/lint/registry/registry.go'),
    ('lint_test.go', 'stage1/cohere/lint/lint_test.go'),
    ('profile_test.go', 'stage1/cohere/lint/profile_test.go'),
]:
    target = scratch / name
    shutil.copyfile(evidence / (name + '.txt'), target)
    replacements[str(repository / relative)] = str(target)
overlay = scratch / 'overlay.json'
overlay.write_text(json.dumps({'Replace': replacements}))
environment = dict(os.environ, ADAMIC_TYPESCRIPT_SOURCE=str(Path(args.compiler).resolve()))
commands = [
    ('corpus.log', ['go', 'test', '-overlay=' + str(overlay), './stage1/cohere/lint',
                    '-run', '^TestWave13Loops', '-count=1', '-v', '-timeout=20m']),
    ('registry.log', ['go', 'test', '-overlay=' + str(overlay), './stage1/cohere/lint/registry', '-count=1', '-v']),
    ('vet.log', ['go', 'vet', '-overlay=' + str(overlay), './...']),
]
for name, command in commands:
    with (scratch / name).open('wb') as log:
        result = subprocess.run(command, cwd=repository, env=environment, stdout=log, stderr=log)
    print(name, 'exit=' + str(result.returncode), flush=True)
    if result.returncode:
        raise SystemExit(result.returncode)
