#!/usr/bin/env python3
"""Run the real tsc CLI and compare all bytes and its process exit status."""
import argparse
from concurrent.futures import ThreadPoolExecutor
import difflib
import json
import os
from pathlib import Path
import shlex
import shutil
import subprocess
import tempfile
import time

from corpus import ROOT, materialize


def command_argv(arguments):
    result = shlex.split(arguments[0]) if len(arguments) == 1 else arguments
    if not result:
        raise ValueError('empty tsc command')
    result[0] = shutil.which(result[0]) or str(Path(result[0]).resolve())
    for index in range(1, len(result)):
        if Path(result[index]).is_file():
            result[index] = str(Path(result[index]).resolve())
    return result


def compare(folder, expected, identifier):
    failures = []
    for suffix in ('stdout', 'stderr', 'exit'):
        actual = (folder / ('actual.' + suffix)).read_bytes()
        wanted = (expected / ('golden.' + suffix)).read_bytes()
        if actual != wanted:
            failures.append(suffix)
            print(f'FAIL {identifier}: {suffix}')
            print(''.join(difflib.unified_diff(wanted.decode('utf-8', 'replace').splitlines(True),
                                             actual.decode('utf-8', 'replace').splitlines(True),
                                             fromfile='golden.' + suffix, tofile='actual.' + suffix)), end='')
    return failures


def run(command, results, record=False, tiny_only=False, golden_root=ROOT):
    selected = json.loads((ROOT / 'selection.json').read_text())['cases']
    rows = [] if tiny_only else [(row['id'], row, ROOT / 'corpus' / row['id']) for row in selected]
    rows.append(('tiny', {'options': {'strict': True, 'target': 'es2020'}}, ROOT / 'tiny'))
    failures = []
    started = time.monotonic()
    def execute(item):
        identifier, row, source = item
        folder = results / identifier
        folder.mkdir(parents=True)
        options = {'types': [], 'skipDefaultLibCheck': True, 'noErrorTruncation': True,
                   'ignoreDeprecations': '6.0', **row['options']}
        files = []
        for program in sorted(source.glob('*.a')):
            name = Path(row['source']).name if identifier != 'tiny' else program.with_suffix('.ts').name
            content, headers = materialize(program.read_bytes())
            if identifier != 'tiny' and headers != row['options']:
                raise RuntimeError(f'header options changed: {identifier}')
            (folder / name).write_text(content)
            files.append(name)
        config = {'compilerOptions': options, 'files': files}
        (folder / 'tsconfig.json').write_text(json.dumps(config, indent=2) + '\n')
        argv = command + ['--project', 'tsconfig.json', '--noEmit', '--pretty', 'false']
        (folder / 'command.json').write_text(json.dumps(argv) + '\n')
        try:
            with (folder / 'actual.stdout').open('wb') as stdout, (folder / 'actual.stderr').open('wb') as stderr:
                completed = subprocess.run(argv, cwd=folder, stdout=stdout, stderr=stderr,
                                           timeout=float(os.environ.get('TSC_TIMEOUT', '60')))
            code = completed.returncode
        except subprocess.TimeoutExpired:
            code = 124
        (folder / 'actual.exit').write_text(str(code) + '\n')
        expected = golden_root / ('tiny' if identifier == 'tiny' else 'corpus/' + identifier)
        if record:
            mismatches = [suffix for suffix in ('stdout', 'stderr', 'exit')
                          if (folder / ('actual.' + suffix)).read_bytes() != (source / ('expected.' + suffix)).read_bytes()]
            if mismatches:
                print(f'FAIL expectation {identifier}: {mismatches}')
                for suffix in mismatches:
                    print((folder / ('actual.' + suffix)).read_text(errors='replace'))
            else:
                for suffix in ('stdout', 'stderr', 'exit'):
                    destination = source / ('golden.' + suffix)
                    if destination.exists():
                        if destination.read_bytes() != (folder / ('actual.' + suffix)).read_bytes():
                            raise RuntimeError('refusing to overwrite golden: ' + str(destination))
                        continue
                    shutil.copyfile(folder / ('actual.' + suffix), destination)
            return identifier if mismatches else None
        return identifier if compare(folder, expected, identifier) else None

    with ThreadPoolExecutor(max_workers=int(os.environ.get('TSC_JOBS', '4'))) as executor:
        failures = [item for item in executor.map(execute, rows) if item is not None]
    report = {'command': command, 'cases': len(rows), 'passed': len(rows) - len(failures),
              'failed': failures, 'seconds': round(time.monotonic() - started, 3)}
    (results / 'report.json').write_text(json.dumps(report, indent=2) + '\n')
    print(json.dumps(report))
    return bool(failures)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--record', action='store_true', help='verify upstream expectations, then save fresh goldens')
    parser.add_argument('--tiny', action='store_true')
    parser.add_argument('--golden-root', type=Path, default=ROOT)
    parser.add_argument('command', nargs=argparse.REMAINDER)
    args = parser.parse_args()
    arguments = args.command[1:] if args.command[:1] == ['--'] else args.command
    if not arguments:
        parser.error('a tsc command is required')
    command = command_argv(arguments)
    result_path = os.environ.get('TSC_RESULTS')
    results = Path(result_path if result_path else tempfile.mkdtemp(prefix='tsc-diagnostics-')).resolve()
    results.mkdir(parents=True, exist_ok=True)
    if (results / 'report.json').exists():
        parser.error('results directory already contains a run')
    print('results: ' + str(results))
    failed = run(command, results, args.record, args.tiny, args.golden_root.resolve())
    native = os.environ.get('NATIVE_TSC')
    if native and not args.record:
        failed = run(command_argv([native]), results / 'native', False, args.tiny, args.golden_root.resolve()) or failed
    raise SystemExit(1 if failed else 0)


if __name__ == '__main__':
    main()
