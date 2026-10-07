#!/usr/bin/env python3
"""Check the four owned, unmodified host fixtures on Node and both backends.

status.json's Node bytes are the truth; its initial Checker stage is historical.
A blocked frontend is recorded explicitly and returns 1, never a green result.
"""
import argparse
import json
from pathlib import Path
import subprocess
import tempfile

repo = Path(__file__).resolve().parents[2]
bucket = repo / 'stage3/fixtures/host'
parser = argparse.ArgumentParser()
parser.add_argument('--logs', required=True, type=Path)
parser.add_argument('--all', action='store_true')
args = parser.parse_args()
args.logs.mkdir(parents=True, exist_ok=True)
owned = {'02_readFile_utf16le.a', '03_readFile_utf16be.a', '21_createHash.a', '22_createHash_fallback.a'}
rows = [row for row in json.loads((bucket / 'status.json').read_text()) if args.all or row['file'] in owned]
assert len(rows) == (25 if args.all else 4)


def run(command, label):
    result = subprocess.run(command, cwd=repo, capture_output=True)
    observation = dict(stdout=result.stdout.decode(), stderr=result.stderr.decode(), exit=result.returncode)
    for key, value in observation.items():
        (args.logs / (label + '.' + key)).write_text(str(value))
    return observation


def stage(observation):
    if observation['exit'] == 0:
        return 'Compiles'
    diagnostic = observation['stdout'] + observation['stderr']
    if ': error TS' in diagnostic:
        return 'Checker'
    if "stage 0 can't lower" in diagnostic:
        return 'NotYet'
    if 'Adamic 0.1 refuses' in diagnostic:
        return 'Refused'
    return 'ToolFailure'


report = []
with tempfile.TemporaryDirectory(prefix='adamic-buffer-host-') as directory:
    compiler = Path(directory) / 'adamic'
    built = run(['go', 'build', '-o', str(compiler), './cmd/adamic'], 'compiler')
    assert built['exit'] == 0, built
    for row in rows:
        file = row['file']
        path = str((bucket / file).relative_to(repo))
        expected = row['node']
        node = run(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', path], file + '.node')
        assert node == expected, file + ': source Node observation differs from status.json'
        binary = Path(directory) / file.removesuffix('.a')
        native_build = run([str(compiler), 'build', path, '-o', str(binary)], file + '.native-build')
        native_stage = stage(native_build)
        if native_stage == 'Compiles':
            assert run([str(binary)], file + '.native') == expected, file + ': native differs from status.json'
        js_build = run([str(compiler), 'js', path], file + '.js-build')
        js_stage = stage(js_build)
        if js_stage == 'Compiles':
            generated = Path(directory) / (file + '.mjs')
            generated.write_text(js_build['stdout'])
            assert run(['node', '--disable-warning=ExperimentalWarning', str(repo / 'oracle/node.mjs'), str(generated)], file + '.javascript') == expected, file + ': JavaScript differs from status.json'
        result = dict(file=file, node='Agrees', native=native_stage, javascript=js_stage,
                      native_diagnostic=native_build['stderr'] if native_stage != 'Compiles' else '',
                      javascript_diagnostic=js_build['stderr'] if js_stage != 'Compiles' else '')
        report.append(result)
        print(file + ': Node agrees; native=' + native_stage + '; JavaScript=' + js_stage, flush=True)
(args.logs / 'stages.json').write_text(json.dumps(report, indent=2) + '\n')
raise SystemExit(0 if all(row['native'] == row['javascript'] == 'Compiles' for row in report) else 1)
