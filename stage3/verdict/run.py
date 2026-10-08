#!/usr/bin/env python3
"""One verdict for exact driver goldens and the declared upstream CLI subset."""
import argparse
from concurrent.futures import ThreadPoolExecutor
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import time

ROOT = Path(__file__).resolve().parent
DRIVER = ROOT.parent / 'drivers/tsc'
sys.path.insert(0, str(DRIVER))
from corpus import PIN, materialize, baseline_output


def write_json(path, value):
    path.write_text(json.dumps(value, indent=2) + '\n')


def first_difference(expected, actual):
    offset = next((index for index, (a, b) in enumerate(zip(expected, actual)) if a != b),
                  min(len(expected), len(actual)))
    start = max(0, offset - 40)
    return {'byte_offset': offset, 'expected_size': len(expected), 'actual_size': len(actual),
            'expected_context': repr(expected[start:offset + 80]),
            'actual_context': repr(actual[start:offset + 80])}


def compare(folder, expected, actual_overrides=None):
    differences = {}
    for stream, wanted in expected.items():
        actual = (actual_overrides or {}).get(stream)
        if actual is None:
            actual = (folder / ('actual.' + stream)).read_bytes()
        if wanted != actual:
            differences[stream] = first_difference(wanted, actual)
    return differences


def driver_suite(binary, output, tiny):
    output.mkdir()
    env = dict(os.environ, TSC_RESULTS=str(output), PYTHONDONTWRITEBYTECODE='1')
    env.pop('NATIVE_TSC', None)
    argv = [sys.executable, str(DRIVER / 'driver.py')]
    if tiny:
        argv.append('--tiny')
    with (output / 'run.log').open('wb') as log:
        completed = subprocess.run(argv + ['--', str(binary)], env=env, stdout=log, stderr=log)
    report_path = output / 'report.json'
    if not report_path.is_file():
        raise RuntimeError(f'driver did not finish: exit {completed.returncode}, see {output}/run.log')
    report = json.loads(report_path.read_text())
    expected_count = 1 if tiny else 301
    if report['cases'] != expected_count:
        raise RuntimeError(f'driver population changed: {report["cases"]} != {expected_count}')
    failures = []
    for identifier in report['failed']:
        expected_dir = DRIVER / ('tiny' if identifier == 'tiny' else 'corpus/' + identifier)
        differences = compare(output / identifier, {suffix: (expected_dir / ('golden.' + suffix)).read_bytes()
                              for suffix in ('stdout', 'stderr', 'exit')})
        failures.append({'case': identifier, 'differences': differences})
    if completed.returncode != bool(failures):
        raise RuntimeError('driver exit/report disagree')
    return {'total': report['cases'], 'passed': report['passed'], 'failed': len(failures),
            'excluded': 0, 'failures': failures, 'seconds': report['seconds']}


def upstream_tree(output):
    configured = os.environ.get('STAGE3_VERDICT_UPSTREAM')
    if configured:
        tree = Path(configured).resolve()
    else:
        tree = output / 'upstream'
        mirror = Path(os.environ.get('STAGE3_CACHE', str(Path.home() / '.cache/adamic-stage3'))) / 'typescript.git'
        with (output / 'upstream.log').open('wb') as log:
            source = str(mirror) if mirror.is_dir() else 'https://github.com/microsoft/TypeScript.git'
            subprocess.run(['git', 'clone', '--depth', '1', '--branch', 'v6.0.3', source, str(tree)],
                           stdout=log, stderr=log, check=True)
    actual = subprocess.check_output(['git', '-C', str(tree), 'rev-parse', 'HEAD'], text=True).strip()
    if actual != PIN:
        raise RuntimeError(f'upstream pin mismatch: {actual}')
    return tree


def baseline_diagnostics(stdout, folder):
    # Upstream removeTestPathPrefixes strips /.src/ everywhere in the summary.
    # Map only this case's equivalent real root; preserve all other path bytes.
    return stdout.replace((str(folder) + '/').encode(), b'')


def validate_manifest(manifest):
    rows = manifest['cases'] + manifest['exclusions']
    if (manifest['upstream_commit'] != PIN
            or manifest['selected'] != len(manifest['cases'])
            or manifest['excluded'] != len(manifest['exclusions'])
            or manifest['total'] != len(rows)
            or len({row['source'] for row in rows}) != len(rows)):
        raise RuntimeError('invalid baseline census')


def baseline_suite(binary, tree, output, limit):
    started = time.monotonic()
    output.mkdir()
    manifest = json.loads((ROOT / 'selection.json').read_text())
    validate_manifest(manifest)
    rows = manifest['cases'] if limit is None else manifest['cases'][:limit]
    if not rows:
        raise RuntimeError('empty baseline suite')
    write_json(output / 'exclusions.json', manifest['exclusions'])
    write_json(output / 'selection.json', rows)
    def execute(item):
        index, row = item
        folder = output / f'{index:05d}_{Path(row["source"]).stem}'
        folder.mkdir()
        raw = (tree / row['source']).read_bytes()
        if hashlib.sha256(raw).hexdigest() != row['source_sha256']:
            raise RuntimeError(f'source hash mismatch: {row["source"]}')
        content, options = materialize(raw)
        if options != row['options']:
            raise RuntimeError('header options changed')
        expected = b''
        if row['baseline']:
            raw_baseline = (tree / row['baseline']).read_bytes()
            if hashlib.sha256(raw_baseline).hexdigest() != row['baseline_sha256']:
                raise RuntimeError(f'baseline hash mismatch: {row["baseline"]}')
            expected = baseline_output(raw_baseline)
        if hashlib.sha256(expected).hexdigest() != row['expected_sha256']:
            raise RuntimeError('baseline summary changed')
        (folder / row['name']).write_text(content)
        config = {'compilerOptions': {'types': [], 'skipDefaultLibCheck': True,
                  'noErrorTruncation': True, 'ignoreDeprecations': '6.0', **options},
                  'files': [row['name']]}
        write_json(folder / 'tsconfig.json', config)
        argv = [str(binary), '--project', 'tsconfig.json', '--noEmit', '--pretty', 'false']
        write_json(folder / 'command.json', argv)
        timed_out = False
        with (folder / 'actual.stdout').open('wb') as stdout, (folder / 'actual.stderr').open('wb') as stderr:
            try:
                completed = subprocess.run(argv, cwd=folder, stdout=stdout, stderr=stderr,
                                           timeout=float(os.environ.get('TSC_TIMEOUT', '60')))
                code = completed.returncode
            except subprocess.TimeoutExpired:
                code, timed_out = 124, True
        (folder / 'actual.exit').write_text(str(code) + '\n')
        wanted = {'stdout': expected, 'stderr': b'', 'exit': b'2\n' if expected else b'0\n'}
        for suffix, value in wanted.items():
            (folder / ('expected.' + suffix)).write_bytes(value)
        diagnostics = baseline_diagnostics((folder / 'actual.stdout').read_bytes(), folder)
        (folder / 'actual.diagnostics').write_bytes(diagnostics)
        differences = compare(folder, wanted, {'stdout': diagnostics})
        if timed_out:
            differences['timeout'] = {'seconds': float(os.environ.get('TSC_TIMEOUT', '60'))}
        return {'case': row['source'], 'capture': folder.name, 'differences': differences} if differences else None
    with ThreadPoolExecutor(max_workers=int(os.environ.get('TSC_JOBS', '4'))) as executor:
        failures = [result for result in executor.map(execute, enumerate(rows, 1)) if result]
    report = {'total': len(rows), 'passed': len(rows) - len(failures), 'failed': len(failures),
              'excluded': manifest['excluded'], 'deferred': manifest['selected'] - len(rows),
              'census_total': manifest['total'], 'exclusion_reasons': manifest['reasons'],
              'failures': failures, 'seconds': round(time.monotonic() - started, 3)}
    write_json(output / 'report.json', report)
    return report


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--tsc', required=True, type=Path, help='one executable, not a shell command')
    parser.add_argument('--baseline-limit', type=int, help='explicit smoke subset; deferred cases remain counted')
    parser.add_argument('output', type=Path)
    args = parser.parse_args()
    binary, output = args.tsc.resolve(), args.output.resolve()
    if not binary.is_file() or not os.access(binary, os.X_OK):
        parser.error('tsc must be an executable file')
    if args.baseline_limit is not None and args.baseline_limit < 1:
        parser.error('baseline limit must be positive')
    if output.exists():
        parser.error('output directory must be new')
    output.mkdir(parents=True)
    suites = {}
    errors = []
    for name, task in [('acceptance', lambda: driver_suite(binary, output / 'acceptance', False)),
                       ('tiny', lambda: driver_suite(binary, output / 'tiny', True)),
                       ('baselines', lambda: baseline_suite(binary, upstream_tree(output), output / 'baselines', args.baseline_limit))]:
        print(f'running {name}', flush=True)
        try:
            suites[name] = task()
            print(f'{name}: {suites[name]["passed"]}/{suites[name]["total"]}', flush=True)
        except Exception as error:
            errors.append({'suite': name, 'error': str(error)})
            suites[name] = {'status': 'harness_error', 'error': str(error)}
    success = not errors and all(row['failed'] == 0 for row in suites.values())
    summary = {'schema_version': 1, 'tsc': str(binary), 'upstream_commit': PIN,
               'success': success, 'suites': suites, 'harness_errors': errors}
    write_json(output / 'summary.json', summary)
    lines = ['# Stage 3 verdict', '', '| Suite | Pass | Fail | Excluded | Deferred |',
             '|---|---:|---:|---:|---:|']
    for name, row in suites.items():
        if 'error' in row:
            lines.append(f'| {name} | harness error | | | |')
        else:
            lines.append(f'| {name} | {row["passed"]} | {row["failed"]} | {row["excluded"]} | {row.get("deferred", 0)} |')
    lines.extend(['', 'Acceptance includes tiny; the separate tiny row repeats that project.', ''])
    for name, row in suites.items():
        if row.get('failures'):
            first = row['failures'][0]
            lines.extend([f'First difference in {name}: `{first["case"]}`', '',
                          '```json', json.dumps(first['differences'], indent=2), '```', ''])
    for error in errors:
        lines.append(f'Harness error in {error["suite"]}: {error["error"]}')
    (output / 'summary.md').write_text('\n'.join(lines) + '\n')
    return 2 if errors else (0 if success else 1)


if __name__ == '__main__':
    raise SystemExit(main())
