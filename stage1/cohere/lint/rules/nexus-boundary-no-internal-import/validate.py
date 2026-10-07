#!/usr/bin/env python3
"""Reproduce the selected Nexus checks using the unmodified published harness.
Source the installed toolchain env.sh first. Setup must run separately, never on this symlink checkout.
"""
import argparse
import os
from pathlib import Path
import shutil
import subprocess
import tarfile

owned = Path(__file__).resolve().parent
repository = owned.parents[4]
parser = argparse.ArgumentParser()
parser.add_argument('--scratch', required=True, type=Path)
parser.add_argument('--typescript', required=True, type=Path)
args = parser.parse_args()
scratch = args.scratch.resolve()
scratch.mkdir(parents=True, exist_ok=False)
archive = scratch / 'harness.tar'
with archive.open('wb') as stream:
    subprocess.run(['git', 'archive', 'f4d98cab50048692781da3599131317dc569d466'], cwd=repository, stdout=stream, check=True)
checkout = scratch / 'checkout'
checkout.mkdir()
with tarfile.open(archive) as stream:
    stream.extractall(checkout, filter='data')
cohere = checkout / 'cohere'
if cohere.is_dir():
    cohere.rmdir()
cohere.symlink_to(repository / 'cohere', target_is_directory=True)
for slug in ['nexus-boundary-no-internal-import', 'nexus-boundary-no-nexus-outside-import', 'nexus-boundary-no-project-import']:
    shutil.copytree(owned.parent / slug, checkout / 'stage1/cohere/lint/rules' / slug)
shutil.copyfile(owned / 'validation_test.go.txt', checkout / 'stage1/cohere/lint/wave12_nexus_test.go')
environment = os.environ.copy()
environment['ADAMIC_TYPESCRIPT_SOURCE'] = str(args.typescript.resolve())
environment['ADAMIC_STAGE1_SOURCE'] = str(repository / 'stage1')
commands = [
    ('fixtures-mutants-corners.log', ['go', 'test', './stage1/cohere/lint', '-run', '^TestNexus(Upstream|Mutants|Corners)$', '-count=1', '-v', '-timeout=15m']),
    ('corpus-throughput.log', ['go', 'test', './stage1/cohere/lint', '-run', '^TestNexus(Corpus|Throughput)$', '-count=1', '-v', '-timeout=15m']),
    ('owned.log', ['go', 'test', './stage1/cohere/lint/rules/nexus-boundary-no-internal-import', '-count=1', '-v', '-timeout=10m']),
    ('registry.log', ['go', 'test', './stage1/cohere/lint/registry', '-count=1', '-v']),
    ('vet.log', ['go', 'vet', './stage1/cohere/lint', './stage1/cohere/lint/registry', './stage1/cohere/lint/rules/nexus-boundary-no-internal-import']),
]
for name, command in commands:
    print(' '.join(command), flush=True)
    with (scratch / name).open('w') as log:
        result = subprocess.run(command, cwd=checkout, env=environment, stdout=log, stderr=subprocess.STDOUT)
    print(f'exit={result.returncode} log={scratch / name}', flush=True)
    if result.returncode:
        raise SystemExit(result.returncode)
