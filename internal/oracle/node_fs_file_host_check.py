#!/usr/bin/env python3
"""Check the fs seat's unchanged real host fixtures against recorded Node bytes.

A blocked fixture is recorded and makes this gate fail; it is never a pass.
"""
import argparse
import json
from pathlib import Path
import subprocess
import tempfile

repo = Path(__file__).resolve().parents[2]
bucket = repo / 'stage3/fixtures/host'
import sys
sys.path.insert(0, str(repo / 'internal/hostconfiguration'))
sys.dont_write_bytecode = True
from configuration import node_for

parser = argparse.ArgumentParser()
parser.add_argument('--compiler', type=Path, required=True)
parser.add_argument('--logs', type=Path, required=True)
parser.add_argument('--report', type=Path, required=True)
parser.add_argument('--mutants', action='store_true')
args = parser.parse_args()
args.logs.mkdir(parents=True, exist_ok=True)
owned = {'01', '02', '03', '04', '05', '06', '10', '11', '12', '13'}
mutations = {
    '01': ('return buffer.toString("utf8", 3);', 'return buffer.toString("utf8", 0);'),
    '02': ('if (len >= 2 && buffer[0] === 0xFF && buffer[1] === 0xFE)', 'if (false)'),
    '03': ('buffer[i] = buffer[i + 1];', 'buffer[i] = temp;'),
    '04': ('return undefined;', 'return "";'),
    '05': ('data = byteOrderMarkIndicator + data;', 'data = data;'),
    '06': ('return stat.isFile();', 'return stat.isDirectory();'),
    '10': ('return statSync(path)?.mtime;', 'return undefined;'),
    '11': ('_fs.utimesSync(path, time, time);', '_fs.utimesSync(path, new Date(0), new Date(0));'),
    '12': ('return _fs.unlinkSync(path);', 'return;'),
    '13': ('if (!nodeSystem.directoryExists(directoryName))', 'if (false)'),
}


def run(command, label):
    result = subprocess.run(command, cwd=repo, capture_output=True)
    for suffix, value in [('stdout', result.stdout), ('stderr', result.stderr),
                          ('exit', str(result.returncode).encode() + b'\n')]:
        (args.logs / (label + '.' + suffix)).write_bytes(value)
    return {'stdout': result.stdout.decode(), 'stderr': result.stderr.decode(), 'exit': result.returncode}


def node(path, label, binary):
    return run([binary, '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(path)], label)


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


report = []
with tempfile.TemporaryDirectory(prefix='fs-file-host-') as scratch:
    scratch = Path(scratch)
    for index, row in enumerate(json.loads((bucket / 'status.json').read_text())):
        prefix = row['file'][:2]
        if prefix not in owned:
            continue
        binary_node = node_for(row, index + 1)
        path = bucket / row['file']
        truth = row['node']
        result = {'file': row['file'], 'nodeMatches': False}
        if binary_node is not None:
            assert node(path, row['file'] + '.node', binary_node) == truth, row['file'] + ': recorded Node observation differs'
            result['nodeMatches'] = True
        else:
            result['nodeReferenceSkipped'] = True
            result['configuration'] = row['configuration']
        binary = scratch / path.stem
        native = run([str(args.compiler), 'build', str(path), '-o', str(binary)], row['file'] + '.native-build')
        javascript = run([str(args.compiler), 'js', str(path)], row['file'] + '.javascript-build')
        for backend, built in [('native', native), ('javascript', javascript)]:
            outcome = stage(built)
            result[backend] = {'stage': outcome, 'diagnostic': '' if outcome == 'Compiles' else built['stdout'] + built['stderr']}
            if outcome == 'Compiles':
                if backend == 'javascript' and binary_node is None:
                    result[backend]['nodeExecutionSkipped'] = True
                    continue
                if backend == 'native':
                    observed = run([str(binary)], row['file'] + '.native')
                else:
                    module = scratch / (path.stem + '.mjs')
                    module.write_text(built['stdout'])
                    observed = node(module, row['file'] + '.javascript', binary_node)
                assert observed == truth, 'SILENT MISCOMPILE: ' + row['file'] + ' ' + backend
                result[backend]['agrees'] = True
        if args.mutants and binary_node is not None:
            before, after = mutations[prefix]
            source = path.read_text()
            assert before in source, row['file'] + ': mutant changed nothing'
            mutant = scratch / row['file']
            mutant.write_text(source.replace(before, after, 1))
            assert node(mutant, row['file'] + '.mutant', binary_node) != truth, 'SURVIVING MUTANT: ' + row['file']
            result['nodeMutantCaught'] = True
        report.append(result)
        print(row['file'] + (': Node agrees; native=' if binary_node is not None else ': recorded Node bytes; native=') + result['native']['stage'] + '; javascript=' + result['javascript']['stage'])
args.report.parent.mkdir(parents=True, exist_ok=True)
args.report.write_text(json.dumps(report, indent=2) + '\n')
raise SystemExit(0 if all(row[backend]['stage'] == 'Compiles' for row in report for backend in ['native', 'javascript']) else 1)
