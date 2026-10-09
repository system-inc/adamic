#!/usr/bin/env python3
"""Record or verify the exact Node and current-main stage-0 observations."""
import argparse
import copy
import json
from pathlib import Path
import re
import subprocess
import tempfile

bucket = Path(__file__).resolve().parent
repo = bucket.parents[2]
parser = argparse.ArgumentParser()
recording = parser.add_mutually_exclusive_group()
recording.add_argument('--record', type=Path, help='metadata array with file, tsc and reason')
recording.add_argument('--update-stage0', action='store_true', help='refresh compiler records only after exact Node agreement')
parser.add_argument('--mutants', action='store_true')
parser.add_argument('--logs', type=Path, required=True)
args = parser.parse_args()
args.logs.mkdir(parents=True, exist_ok=True)
status = json.loads((args.record or bucket / 'status.json').read_text())


def run(command, label):
    result = subprocess.run(command, cwd=repo, capture_output=True)
    (args.logs / (label + '.stdout')).write_bytes(result.stdout)
    (args.logs / (label + '.stderr')).write_bytes(result.stderr)
    (args.logs / (label + '.exit')).write_text(str(result.returncode) + '\n')
    return dict(stdout=result.stdout.decode(), stderr=result.stderr.decode(), exit=result.returncode)


def node(file, label):
    return run(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(file)], label)


def compare(observation, expected, label):
    if observation != expected:
        raise AssertionError(label + ': exact observation differs')


def normalize_stage0(text):
    # Match fixtures_test.go's two path replacements. Its direct err.Error()
    # has neither the CLI prefix nor go run's trailer/final newline.
    text = text.removeprefix('adamic: ')
    text = text.replace(str(bucket.parent) + '/', 'stage3/fixtures/')
    text = text.replace(str(repo) + '/', '')
    return text.removesuffix('exit status 1\n').removesuffix('\n')


def compare_stage0(observation, expected, label):
    compare({**observation, 'what': normalize_stage0(observation['what'])},
            {**expected, 'what': normalize_stage0(expected['what'])}, label)


def update_stage0_records():
    # As in fixtures_test.go's replaceStage0, leave all bytes outside stage0
    # intact, including the independently recorded Node oracle and provenance.
    path = bucket / 'status.json'
    original = path.read_text()
    decoder = json.JSONDecoder()
    fields = list(re.finditer(r'"stage0"\s*:\s*', original))
    assert len(fields) == len(status)
    replacements = []
    for field, row in zip(fields, status):
        start = field.end()
        _, end = decoder.raw_decode(original, start)
        line_start = original.rfind('\n', 0, field.start()) + 1
        indentation = original[line_start:field.start()]
        assert indentation.isspace()
        encoded = json.dumps(row['stage0'], ensure_ascii=False, indent=2).replace('\n', '\n' + indentation)
        replacements.append((start, end, encoded))
    result = original
    for start, end, replacement in reversed(replacements):
        result = result[:start] + replacement + result[end:]
    assert path.read_text() == original, 'status.json changed during observation'
    path.write_text(result)


with tempfile.TemporaryDirectory(prefix='adamic-host-check-') as scratch:
    for row in status:
        file = bucket / row['file']
        observed = node(file.relative_to(repo), row['file'] + '.node')
        # A fixture must not silently record a crash as its intended Node behavior.
        assert observed['stderr'] == '', (row['file'], observed)
        assert observed['exit'] == ({'18_exit_0.a': 0, '19_exit_1.a': 1, '20_exit_2.a': 2}.get(row['file'], 0))
        if not args.record:
            compare(observed, row['node'], row['file'] + ' Node')
        binary = Path(scratch) / row['file'].replace('.a', '')
        built = run(['go', 'run', './cmd/adamic', 'build', str(file.relative_to(repo)), '-o', str(binary)], row['file'] + '.stage0')
        diagnostic = normalize_stage0(built['stdout'] + built['stderr'])
        if built['exit'] != 0:
            assert any(marker in diagnostic for marker in ["stage 0 can't lower", 'Adamic 0.1 refuses', ': error TS']), ('unexpected build failure', diagnostic)
        outcome = 'Compiles' if built['exit'] == 0 else 'NotYet' if "stage 0 can't lower" in diagnostic else 'Refused' if 'Adamic 0.1 refuses' in diagnostic else 'Checker'
        if outcome == 'Compiles':
            native = run([str(binary)], row['file'] + '.native')
            if native != observed:
                raise AssertionError('SILENT MISCOMPILE: ' + row['file'] + ': native differs from Node')
        stage0 = dict(outcome=outcome, what='' if outcome == 'Compiles' else diagnostic)
        if args.record:
            row['node'], row['stage0'] = observed, stage0
        elif args.update_stage0:
            row['stage0'] = stage0
        else:
            compare_stage0(stage0, row['stage0'], row['file'] + ' stage0')
        print(row['file'] + ': Node exit=' + str(observed['exit']) + ', stage0=' + outcome)
    if args.record:
        (bucket / 'status.json').write_text(json.dumps(status, ensure_ascii=False, indent=2) + '\n')
    elif args.update_stage0:
        update_stage0_records()
    if args.mutants:
        mutations = [
            ('return buffer.toString("utf8", 3);', 'return buffer.toString("utf8", 0);'),
            ('if (len >= 2 && buffer[0] === 0xFF && buffer[1] === 0xFE)', 'if (false)'),
            ('buffer[i] = buffer[i + 1];', 'buffer[i] = temp;'),
            ('return undefined;', 'return "";'),
            ('data = byteOrderMarkIndicator + data;', 'data = data;'),
            ('return stat.isFile();', 'return stat.isDirectory();'),
            ('return stat.isDirectory();', 'return stat.isFile();'),
            ('directories.sort();', 'directories.reverse();'),
            ('return path;', 'return _path.resolve(path);'),
            ('return statSync(path)?.mtime;', 'return undefined;'),
            ('_fs.utimesSync(path, time, time);', '_fs.utimesSync(path, new Date(0), new Date(0));'),
            ('return _fs.unlinkSync(path);', 'return;'),
            ('if (!nodeSystem.directoryExists(directoryName))', 'if (false)'),
            ('callback = undefined!;', '// callback stays live'),
            ('__filename.endsWith("sys.js")', '__filename.endsWith("sys.cjs")'),
            ('return process.env[name] || "";', 'return "";'),
            ('process.stdout.write(s);', 'process.stdout.write(s + "?");'),
            ('nodeSystem.exit(0);', 'nodeSystem.exit(1);'),
            ('nodeSystem.exit(1);', 'nodeSystem.exit(2);'),
            ('nodeSystem.exit(2);', 'nodeSystem.exit(0);'),
            ('createHash("sha256")', 'createHash("sha1")'),
            ('let acc = 5381;', 'let acc = 5382;'),
            ('newLine: _os.EOL', 'newLine: "\\r\\n"'),
            ('return !fileExists(swapCase(__filename));', 'return fileExists(swapCase(__filename));'),
            ('if (extensions && !fileExtensionIsOneOf(name, extensions)) continue;', 'if (false) continue;'),
        ]
        assert len(mutations) == len(status)
        for row, (before, after) in zip(status, mutations):
            source = (bucket / row['file']).read_text()
            assert before in source, (row['file'], before)
            changed = source.replace(before, after, 1)
            mutant = Path(scratch) / row['file']
            mutant.write_text(changed)
            observed = node(mutant, row['file'] + '.mutant')
            try:
                compare(observed, row['node'], row['file'] + ' mutant')
            except AssertionError:
                print(row['file'] + ': caught mutant ' + before + ' -> ' + after)
            else:
                raise AssertionError('SURVIVING MUTANT: ' + row['file'])
        # Every compiler record must retain message, source line and code after
        # normalization. Change one field at a time; use the real comparison.
        for row in status:
            wrapped = copy.deepcopy(row['stage0'])
            wrapped['what'] = ('adamic: ' + wrapped['what'].replace('stage3/fixtures/', str(bucket.parent) + '/')
                               + '\nexit status 1\n')
            compare_stage0(wrapped, row['stage0'], row['file'] + ' CLI normalization')
            for kind in ('diagnostic', 'line', 'code'):
                mutated_status = copy.deepcopy(row['stage0'])
                text = mutated_status['what']
                if kind == 'diagnostic':
                    changed, count = re.subn(r'(error TS\d+: )', r'\1[changed diagnostic] ', text, count=1)
                elif kind == 'line':
                    changed, count = re.subn(r':(\d+)(:\d+: error TS)',
                                             lambda match: ':' + str(int(match[1]) + 1) + match[2], text, count=1)
                else:
                    changed, count = re.subn(r'error TS\d+:', 'error TS999999:', text, count=1)
                assert count == 1 and changed != text, (row['file'], kind)
                mutated_status['what'] = changed
                try:
                    compare_stage0(row['stage0'], mutated_status, row['file'] + ' stage0 ' + kind + ' mutant')
                except AssertionError:
                    print(row['file'] + ': caught stage0 ' + kind + ' mutant after normalization')
                else:
                    raise AssertionError('SURVIVING STATUS MUTANT: ' + row['file'] + ' ' + kind)
