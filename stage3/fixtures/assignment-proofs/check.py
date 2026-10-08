#!/usr/bin/env python3
"""Run only this bucket; source Node is the oracle, not emitted JavaScript."""
import argparse
import json
import pathlib
import subprocess
import tempfile

bucket = pathlib.Path(__file__).resolve().parent
repository = bucket.parents[2]
parser = argparse.ArgumentParser()
parser.add_argument('--mutant', type=int)
parser.add_argument('--require-adamic', action='store_true')
args = parser.parse_args()
contract = json.loads((bucket / 'expectations.json').read_text())
records = contract['records']
assert contract['upstream_source']['function_text'] in (bucket / '01_scanner_keyword.a').read_text(), 'scanner upstream function provenance'
logs = pathlib.Path(tempfile.mkdtemp(prefix='step12-assignment-'))
status, mutants = [], []

def node(file):
    return subprocess.run(['node', '--disable-warning=ExperimentalWarning', str(repository / 'oracle/node.mjs'), str(file)], capture_output=True)

def compare(run, expected, name):
    assert run.stdout == expected['stdout'].encode(), f'{name}: Node stdout byte comparison'
    assert run.stderr == expected['stderr'].encode(), f'{name}: Node stderr byte comparison'
    assert run.returncode == expected['exit'], f'{name}: Node exit comparison'

def typecheck(file):
    run = subprocess.run(['node', str(bucket / 'check-types.cjs'), str(file)], capture_output=True)
    (logs / (file.stem + '.types.log')).write_bytes(run.stdout + run.stderr)
    assert run.returncode == 0, run.stderr.decode()

for index, record in enumerate(records, 1):
    if args.mutant is not None and args.mutant != index:
        continue
    file = bucket / record['file']
    source = file.read_text()
    assert source.startswith('// a-check: pass;'), record['file']
    if args.mutant:
        mutation = record['mutant']
        assert source.count(mutation['replace']) == 1
        file = logs / record['file']
        file.write_text(source.replace(mutation['replace'], mutation['with']))
    typecheck(file)
    run = node(file)
    (logs / (file.stem + '.node.stdout')).write_bytes(run.stdout)
    (logs / (file.stem + '.node.stderr')).write_bytes(run.stderr)
    compare(run, record['node'], record['file'])
    binary = logs / (file.stem + '.native')
    build = subprocess.run(['go', 'run', './cmd/adamic', 'build', str(file), '-o', str(binary)], cwd=repository, capture_output=True)
    text = (build.stdout + build.stderr).decode()
    (logs / (file.stem + '.build.log')).write_text(text)
    if build.returncode == 0:
        native = subprocess.run([str(binary)], capture_output=True)
        assert (native.stdout, native.stderr, native.returncode) == (run.stdout, run.stderr, run.returncode), f'SILENT MISCOMPILE: {file}; logs {logs}'
        outcome = 'Compiles'
        what = ''
    else:
        outcome = 'NotYet' if "can't lower" in text or 'not yet' in text.lower() else 'Refused' if 'refuses' in text or 'refused' in text else 'Checker'
        what = text.removeprefix("adamic: ").removesuffix("exit status 1\n").rstrip("\n").replace(str(repository) + "/", "")
    if args.require_adamic:
        assert outcome == record['expected_adamic']['outcome'], f'{file}: expected Compiles, observed {outcome}'
    status.append({'file': record['file'], 'tsc': record['source'], 'reason': record['reason'], 'node': record['node'], 'stage0': {'outcome': outcome, 'what': what}})
if not args.mutant:
    for index, record in enumerate(records, 1):
        run = subprocess.run(['python3', str(__file__), '--mutant', str(index)], capture_output=True)
        (logs / (record['file'] + '.mutant.log')).write_bytes(run.stdout + run.stderr)
        assert run.returncode == 1 and b'Node stdout byte comparison' in run.stderr, f'{record["file"]}: mutant did not fail the intended comparison: {run.stderr.decode()}'
        mutants.append({'file': record['file'], 'exit': run.returncode, 'caught_by': 'source Node stdout byte comparison after zero stock TypeScript diagnostics', 'mutation': record['mutant']})
    (bucket / 'status.json').write_text(json.dumps(status, indent=2) + '\n')
    (bucket / 'mutants.json').write_text(json.dumps(mutants, indent=2) + '\n')
    counts = {outcome: sum(row['stage0']['outcome'] == outcome for row in status) for outcome in ['Compiles', 'NotYet', 'Refused', 'Checker']}
    (bucket / 'counts.md').write_text('# Assignment-proof fixture counts\n\n6 fixtures, 6 source Node goldens, 6 caught Node input/statement mutants.\n\nCurrent main: ' + ', '.join(f'{name}={count}' for name, count in counts.items()) + '.\n\nRuled target: all 6 compile and match Node; proof propagation adds no redundant proof check. Native allocation counts are unavailable for NotYet fixtures. No central oracle fixtures were added.\n')
    print(f'PASS: 6 Node goldens, 6 TypeScript checks, 6 caught mutants; current main {counts}; logs {logs}')
