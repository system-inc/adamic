#!/usr/bin/env python3
"""Validate owned ports against the exact published harness, without shared edits.
Source the toolchain env.sh first. Setup is run separately, outside this runner.
"""
import argparse
from pathlib import Path
import shutil
import subprocess
import tarfile
import os

owned = Path(__file__).resolve().parent
repository = owned.parents[4]
parser = argparse.ArgumentParser()
parser.add_argument('--scratch', type=Path, required=True)
parser.add_argument('--typescript', type=Path, required=True)
args = parser.parse_args()
scratch = args.scratch.resolve()
scratch.mkdir(parents=True, exist_ok=True)
revision = '2650ad595b82220c368631ea13139fad4b306ed6'
archive = scratch / 'harness.tar'
with archive.open('wb') as stream:
    subprocess.run(['git', 'archive', revision], cwd=repository, stdout=stream, check=True)
checkout = scratch / 'checkout'
checkout.mkdir(exist_ok=True)
with tarfile.open(archive) as stream:
    stream.extractall(checkout, filter='data')
cohere = checkout / 'cohere'
if cohere.is_dir() and not cohere.is_symlink():
    cohere.rmdir()
if not cohere.exists():
    cohere.symlink_to(repository / 'cohere', target_is_directory=True)
for slug in ['typescript-no-non-null-asserted-optional-chain', 'typescript-no-unnecessary-type-constraint', 'typescript-no-this-alias']:
    shutil.copytree(owned.parent / slug, checkout / 'stage1/cohere/lint/rules' / slug, dirs_exist_ok=True)
shutil.copyfile(owned / 'validation_test.go.txt', checkout / 'stage1/cohere/lint/wave12_test.go')
environment = os.environ.copy()
environment['ADAMIC_TYPESCRIPT_SOURCE'] = str(args.typescript.resolve())
environment['ADAMIC_STAGE1_SOURCE'] = str(repository / 'stage1')
commands = [
    ('upstream.log', ['go','test','./stage1/cohere/lint','-run','^TestCompletedUpstream$','-count=1','-v','-timeout=10m']),
    ('corpus.log', ['go','test','./stage1/cohere/lint','-run','^(TestCompletedCorpus|TestCompletedThroughput|TestCompletedCorners)$','-count=1','-v','-timeout=20m']),
    ('mutants.log', ['go','test','./stage1/cohere/lint','-run','^TestCompletedMutants$','-count=1','-v','-timeout=10m']),
    ('registry.log', ['go','test','./stage1/cohere/lint/registry','-count=1','-v']),
]
for name, command in commands:
    print(' '.join(command), flush=True)
    with (scratch / name).open('w') as log:
        result = subprocess.run(command, cwd=checkout, env=environment, stdout=log, stderr=subprocess.STDOUT)
    print(f'exit={result.returncode} log={scratch / name}', flush=True)
    if result.returncode:
        raise SystemExit(result.returncode)
