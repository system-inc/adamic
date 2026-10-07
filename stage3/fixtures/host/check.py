#!/usr/bin/env python3
"""Record or verify the exact Node and selected compiler stage-0 observations."""
import argparse
import copy
import json
import re
from pathlib import Path
import subprocess
import tempfile

bucket = Path(__file__).resolve().parent
repo = bucket.parents[2]
parser = argparse.ArgumentParser()
parser.add_argument('--record', type=Path, help='metadata array with file, tsc and reason')
parser.add_argument('--mutants', action='store_true')
parser.add_argument('--mutants-only', action='store_true')
parser.add_argument('--compiler-repo', type=Path, default=repo)
parser.add_argument('--logs', type=Path, required=True)
args = parser.parse_args()
args.logs.mkdir(parents=True, exist_ok=True)
# The selected host-capable library compiler may predate main's option ruling.
# A Go overlay changes only its two house-style options and erasable syntax
# policy to current main, without editing the compiler checkout.
loader_path = args.compiler_repo.resolve() / 'internal/load/load.go'
loader_source = loader_path.read_text()
for style_option in ('NoImplicitReturns', 'NoFallthroughCasesInSwitch'):
    loader_source = re.sub(r'^\s*' + style_option + r':\s*core.TSTrue,\n', '', loader_source, flags=re.MULTILINE)
loader_source = re.sub(r'(ErasableSyntaxOnly:\s*)core.TSTrue', r'\1core.TSFalse', loader_source)
assert not re.search(r'(NoImplicitReturns|NoFallthroughCasesInSwitch):', loader_source)
overlay_source = args.logs.resolve() / 'load.go'
overlay_source.write_text(loader_source)
overlay = args.logs.resolve() / 'checker-overlay.json'
overlay.write_text(json.dumps({'Replace': {str(loader_path): str(overlay_source)}}, indent=2) + '\n')
status = json.loads((args.record or bucket / 'status.json').read_text())


def run(command, label, cwd=repo):
    result = subprocess.run(command, cwd=cwd, capture_output=True)
    (args.logs / (label + '.stdout')).write_bytes(result.stdout)
    (args.logs / (label + '.stderr')).write_bytes(result.stderr)
    (args.logs / (label + '.exit')).write_text(str(result.returncode) + '\n')
    return dict(stdout=result.stdout.decode(), stderr=result.stderr.decode(), exit=result.returncode)


def node(file, label):
    return run(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(file)], label)


def compare(observation, expected, label):
    if observation != expected:
        raise AssertionError(label + ': exact observation differs')


with tempfile.TemporaryDirectory(prefix='adamic-host-check-') as scratch:
    for row in ([] if args.mutants_only else status):
        file = bucket / row['file']
        observed = node(file.relative_to(repo), row['file'] + '.node')
        # A fixture must not silently record a crash as its intended Node behavior.
        assert observed['stderr'] == '', (row['file'], observed)
        assert observed['exit'] == ({'18_exit_0.a': 0, '19_exit_1.a': 1, '20_exit_2.a': 2}.get(row['file'], 0))
        binary = Path(scratch) / row['file'].replace('.a', '')
        built = run(['go', 'run', '-overlay=' + str(overlay), './cmd/adamic', 'build', str(file), '-o', str(binary)], row['file'] + '.stage0', args.compiler_repo)
        diagnostic = built['stdout'] + built['stderr']
        diagnostic = diagnostic.removeprefix('adamic: ').removesuffix('exit status 1\n').rstrip('\n')
        diagnostic = diagnostic.replace(str(bucket) + '/', 'stage3/fixtures/host/').replace(str(args.compiler_repo.resolve()) + '/', '')
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
        else:
            compare(observed, row['node'], row['file'] + ' Node')
            compare(stage0, row['stage0'], row['file'] + ' stage0')
        print(row['file'] + ': Node exit=' + str(observed['exit']) + ', stage0=' + outcome)
    if args.record:
        (bucket / 'status.json').write_text(json.dumps(status, ensure_ascii=False, indent=2) + '\n')
    if args.mutants or args.mutants_only:
        mutations = [
            ('return buffer.toString("utf8", 3);', 'return buffer.toString("utf8", 0);'),
            ('if (len >= 2 && buffer[0]! === 0xFF && buffer[1]! === 0xFE)', 'if (false)'),
            ('buffer[i] = buffer[i + 1]!;', 'buffer[i] = temp;'),
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
        # Prove that an inaccurate recorded compiler diagnostic is noticed too.
        mutated_status = copy.deepcopy(status[0]['stage0'])
        mutated_status['what'] += 'invented diagnostic\n'
        try:
            compare(status[0]['stage0'], mutated_status, 'stage0 diagnostic mutant')
        except AssertionError:
            print('caught status diagnostic mutant')
        else:
            raise AssertionError('SURVIVING STATUS MUTANT')
