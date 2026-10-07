#!/usr/bin/env python3
"""Check the fs seat's unchanged real host fixtures against recorded Node bytes.

A blocked fixture is recorded and makes this gate fail; it is never a pass.
"""
import argparse
import json
from pathlib import Path
import subprocess
import tempfile

repo = Path('/workspace/adamic')
bucket = repo / 'stage3/fixtures/host'
parser = argparse.ArgumentParser()
parser.add_argument('--compiler', type=Path, required=True)
parser.add_argument('--logs', type=Path, required=True)
parser.add_argument('--report', type=Path, required=True)
parser.add_argument('--mutants', action='store_true')
parser.add_argument('--all', action='store_true', help='observe all 25 host fixtures')
parser.add_argument('--only', nargs='+', help='observe only these two-digit fixture numbers')
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


report = []
with tempfile.TemporaryDirectory(prefix='fs-file-host-') as scratch:
    scratch = Path(scratch)
    for row in json.loads((bucket / 'status.json').read_text()):
        prefix = row['file'][:2]
        if args.only and prefix not in args.only:
            continue
        if prefix not in owned and not args.all:
            continue
        path = bucket / row['file']
        truth = node(path, row['file'] + '.node')
        assert truth == row['node'], row['file'] + ': recorded Node observation differs'
        result = {'file': row['file'], 'nodeMatches': True}
        binary = scratch / path.stem
        native = run([str(args.compiler), 'build', str(path), '-o', str(binary)], row['file'] + '.native-build')
        javascript = run([str(args.compiler), 'js', str(path)], row['file'] + '.javascript-build')
        for backend, built in [('native', native), ('javascript', javascript)]:
            outcome = stage(built)
            result[backend] = {'stage': outcome, 'diagnostic': '' if outcome == 'Compiles' else built['stdout'] + built['stderr']}
            if outcome == 'Compiles':
                if backend == 'native':
                    observed = run([str(binary)], row['file'] + '.native')
                else:
                    module = scratch / (path.stem + '.mjs')
                    module.write_text(built['stdout'])
                    observed = node(module, row['file'] + '.javascript')
                if observed == truth:
                    result[backend]['agrees'] = True
                else:
                    result[backend]['stage'] = 'Disagrees'
                    result[backend]['diagnostic'] = 'Runtime disagrees with Node: ' + json.dumps(observed, sort_keys=True)
                    result[backend]['observed'] = observed
                    result[backend]['expected'] = truth
        if args.mutants and prefix in mutations:
            before, after = mutations[prefix]
            source = path.read_text()
            assert before in source, row['file'] + ': mutant changed nothing'
            mutant = scratch / row['file']
            mutant.write_text(source.replace(before, after, 1))
            assert node(mutant, row['file'] + '.mutant') != truth, 'SURVIVING MUTANT: ' + row['file']
            result['nodeMutantCaught'] = True
        report.append(result)
        print(row['file'] + ': Node agrees; native=' + result['native']['stage'] + '; javascript=' + result['javascript']['stage'])
args.report.parent.mkdir(parents=True, exist_ok=True)
args.report.write_text(json.dumps(report, indent=2) + '\n')
raise SystemExit(0 if all(row[backend]['stage'] == 'Compiles' for row in report for backend in ['native', 'javascript']) else 1)
