#!/usr/bin/env python3
"""Check the fs seat's unchanged read/decode/hash host fixtures against recorded Node bytes.

A blocked fixture is recorded and makes this gate fail; it is never a pass.
"""
import argparse
import json
from pathlib import Path
import subprocess
import tempfile

repo = Path(__file__).resolve().parents[2]
bucket = repo / 'stage3/fixtures/host'
parser = argparse.ArgumentParser()
parser.add_argument('--compiler', type=Path, required=True)
parser.add_argument('--logs', type=Path, required=True)
parser.add_argument('--report', type=Path, required=True)
parser.add_argument('--mutants', action='store_true')
args = parser.parse_args()
args.logs.mkdir(parents=True, exist_ok=True)
owned = {'01', '02', '03', '04', '21', '22'}
mutations = {
    '01': ('return buffer.toString("utf8", 3);', 'return buffer.toString("utf8", 0);'),
    '02': ('if (len >= 2 && buffer[0] === 0xFF && buffer[1] === 0xFE)', 'if (false)'),
    '03': ('buffer[i] = buffer[i + 1];', 'buffer[i] = temp;'),
    '04': ('return undefined;', 'return "";'),
    '21': ('createHash("sha256")', 'createHash("sha1")'),
    '22': ('let acc = 5381;', 'let acc = 5382;'),
}


def run(command, label):
    result = subprocess.run(command, cwd=repo, capture_output=True)
    for suffix, value in [('stdout', result.stdout), ('stderr', result.stderr),
                          ('exit', str(result.returncode).encode() + b'\n')]:
        (args.logs / (label + '.' + suffix)).write_bytes(value)
    return {'stdout': result.stdout.decode(), 'stderr': result.stderr.decode(), 'exit': result.returncode}


def node(path, label):
    return run(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(path)], label)


def stage(observation):
    diagnostic = observation['stdout'] + observation['stderr']
    if observation['exit'] == 0:
        return 'Compiles'
    if ': error TS' in diagnostic:
        return 'Checker'
    if "stage 0 can't lower" in diagnostic:
        return 'NotYet'
    if 'Adamic 0.1 refuses' in diagnostic:
        return 'Refused'
    raise AssertionError('unexpected compiler/tool failure: ' + diagnostic)


assert subprocess.check_output(['node', '--version'], text=True).strip() == 'v24.19.0', 'Node 24.19.0 is required'
report = []
with tempfile.TemporaryDirectory(prefix='fs-file-host-') as scratch:
    scratch = Path(scratch)
    for row in json.loads((bucket / 'status.json').read_text()):
        prefix = row['file'][:2]
        if prefix not in owned:
            continue
        path = bucket / row['file']
        truth = node(path, row['file'] + '.node')
        assert truth == row['node'], row['file'] + ': recorded Node observation differs'
        result = {'file': row['file'], 'nodeMatches': True}
        binary = scratch / path.stem
        native = run([str(args.compiler), 'build', str(path), '-o', str(binary)], row['file'] + '.native-build')
        javascript = run([str(args.compiler), 'js', str(path)], row['file'] + '.javascript-build')
        wasm = scratch / (path.stem + '.wasm')
        wasi = run([str(args.compiler), 'build', '--target', 'wasm32-wasi', str(path), '-o', str(wasm)], row['file'] + '.wasi-build')
        for backend, built in [('native', native), ('javascript', javascript), ('wasi', wasi)]:
            outcome = stage(built)
            result[backend] = {'stage': outcome, 'diagnostic': '' if outcome == 'Compiles' else built['stdout'] + built['stderr']}
            if outcome == 'Compiles':
                if backend == 'native':
                    observed = run([str(binary)], row['file'] + '.native')
                elif backend == 'wasi':
                    observed = run(['node', '--disable-warning=ExperimentalWarning', 'oracle/wasi.mjs', str(wasm)], row['file'] + '.wasi')
                else:
                    module = scratch / (path.stem + '.mjs')
                    module.write_text(built['stdout'])
                    observed = node(module, row['file'] + '.javascript')
                assert observed == truth, 'SILENT MISCOMPILE: ' + row['file'] + ' ' + backend
                result[backend]['agrees'] = True
        if args.mutants and prefix in mutations:
            before, after = mutations[prefix]
            source = path.read_text()
            assert before in source, row['file'] + ': mutant changed nothing'
            mutant = scratch / row['file']
            mutant.write_text(source.replace(before, after, 1))
            changed = node(mutant, row['file'] + '.mutant')
            assert changed['exit'] == truth['exit'] == 0 and changed['stderr'] == truth['stderr'] == '', 'UNCLEAN MUTANT: ' + row['file']
            assert changed['stdout'] != truth['stdout'], 'SURVIVING MUTANT: ' + row['file']
            result['nodeMutantCaught'] = True
        report.append(result)
        print(row['file'] + ': Node agrees; native=' + result['native']['stage'] + '; javascript=' + result['javascript']['stage'] + '; wasi=' + result['wasi']['stage'])
args.report.parent.mkdir(parents=True, exist_ok=True)
args.report.write_text(json.dumps(report, indent=2) + '\n')
raise SystemExit(0 if all(row[backend]['stage'] == 'Compiles' for row in report for backend in ['native', 'javascript', 'wasi']) else 1)
