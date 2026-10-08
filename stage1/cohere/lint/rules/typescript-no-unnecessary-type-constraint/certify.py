#!/usr/bin/env python3
"""Run the seven winning wave08 rules without changing the shared harness."""
import argparse
import json
import os
from pathlib import Path
import subprocess
import tempfile

owned = Path(__file__).resolve().parent
repository = owned.parents[4]
parser = argparse.ArgumentParser()
parser.add_argument('--typescript', type=Path, required=True)
parser.add_argument('--log', type=Path, required=True)
args = parser.parse_args()
with tempfile.TemporaryDirectory(prefix='wave08-certification-') as temporary:
    overlay = Path(temporary) / 'overlay.json'
    overlay.write_text(json.dumps({'Replace': {
        str(repository / 'stage1/cohere/lint/wave08_certification_test.go'):
        str(owned / 'certification_test.go.txt'),
    }}))
    environment = os.environ.copy()
    environment['ADAMIC_TYPESCRIPT_SOURCE'] = str(args.typescript.resolve())
    environment['ADAMIC_GATE_UNCACHED'] = '1'
    command = ['go', 'test', '-overlay=' + str(overlay), './stage1/cohere/lint',
               '-run', '^TestWave08Complete', '-count=1', '-v', '-timeout=45m', '-parallel=2']
    with args.log.open('w') as log:
        result = subprocess.run(command, cwd=repository, env=environment,
                                stdout=log, stderr=subprocess.STDOUT)
    print(f'exit={result.returncode} log={args.log.resolve()}')
    raise SystemExit(result.returncode)
