#!/usr/bin/env python3
"""One verdict for exact driver goldens and the declared upstream CLI subset."""
import argparse
from concurrent.futures import ThreadPoolExecutor
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import time

ROOT = Path(__file__).resolve().parent
DRIVER = ROOT.parent / 'drivers/tsc'
sys.path.insert(0, str(DRIVER))
from corpus import PIN
from cases import parse, configurations, arguments, layout, expected_exit, pretty_diagnostics, config_options, has_type_packages, diagnostics as map_diagnostics
from census import summary_bytes


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
    selected_sources = {row['source'] for row in manifest['cases']}
    excluded_sources = {row['source'] for row in manifest['exclusions']}
    identities = {(row['source'], row.get('configuration', '')) for row in manifest['cases']}
    skipped = manifest.get('configuration_exclusions', [])
    skipped_identities = {(row['source'], row['configuration']) for row in skipped}
    if (manifest['upstream_commit'] != PIN
            or manifest['selected'] != len(selected_sources)
            or manifest.get('configurations', len(manifest['cases'])) != len(manifest['cases'])
            or manifest['excluded'] != len(manifest['exclusions'])
            or manifest['total'] != len(selected_sources) + len(excluded_sources)
            or selected_sources & excluded_sources
            or len(excluded_sources) != len(manifest['exclusions'])
            or len(identities) != len(manifest['cases'])
            or identities & skipped_identities
            or len(skipped_identities) != len(skipped)
            or any(row['source'] not in selected_sources | excluded_sources for row in skipped)):
        raise RuntimeError('invalid baseline census')


def baseline_suite(binary, tree, output, limit, manifest=None):
    started = time.monotonic()
    output.mkdir()
    for ancestor in output.parents:
        if (ancestor / 'node_modules').exists() or (ancestor / 'package.json').exists():
            raise RuntimeError(f'outside package metadata at {ancestor}; choose an isolated output directory such as /tmp')
    manifest = manifest or json.loads((ROOT / 'selection.json').read_text())
    validate_manifest(manifest)
    rows = manifest['cases'] if limit is None else manifest['cases'][:limit]
    if not rows:
        raise RuntimeError('empty baseline suite')
    write_json(output / 'exclusions.json', manifest['exclusions'])
    write_json(output / 'configuration-exclusions.json', manifest.get('configuration_exclusions', []))
    write_json(output / 'selection.json', rows)
    if any(parse((tree / row['source']).read_bytes(), row['name'])[1].get('__namespace') for row in rows):
        from resources import prepare
        prepare(tree, output)
    def execute(item):
        index, row = item
        folder = output / f'{index:05d}_{Path(row["source"]).stem}'
        folder.mkdir()
        raw = (tree / row['source']).read_bytes()
        if hashlib.sha256(raw).hexdigest() != row['source_sha256']:
            raise RuntimeError(f'source hash mismatch: {row["source"]}')
        units, settings, roots = parse(raw, row['name'])
        options = dict(configurations(settings))[row.get('configuration', '')]
        if options != row['options']:
            raise RuntimeError('header options changed')
        expected = b''
        if row['baseline']:
            raw_baseline = (tree / row['baseline']).read_bytes()
            if hashlib.sha256(raw_baseline).hexdigest() != row['baseline_sha256']:
                raise RuntimeError(f'baseline hash mismatch: {row["baseline"]}')
            expected = summary_bytes(raw_baseline)
        if hashlib.sha256(expected).hexdigest() != row['expected_sha256']:
            raise RuntimeError('baseline summary changed')
        cwd, physical, mapped_options = layout(units, options, folder, settings)
        for unit in units:
            destination = physical(unit['name'])
            destination.parent.mkdir(parents=True, exist_ok=True)
            destination.write_text(unit['content'])
        for target, destination in settings['__links']:
            link = physical(destination)
            link.parent.mkdir(parents=True, exist_ok=True)
            resolved = physical(target)
            if settings.get('__namespace'):
                resolved = Path('/') / resolved.relative_to(folder / 'filesystem')
            link.symlink_to(resolved, target_is_directory=True)
        write_json(folder / 'units.json', {'units': units, 'roots': roots, 'options': options})
        mapped_roots = [name if re.match(r'^[A-Za-z]:/', name) else str(physical(name)) if cwd != folder else name for name in roots]
        project = row.get('project')
        effective = row.get('effective_options', options)
        if settings.get('__namespace'):
            for key in ('outDir', 'declarationDir'):
                value = effective.get(key)
                if isinstance(value, str) and value.startswith('/'):
                    physical(value).mkdir(parents=True, exist_ok=True)
        config_values = config_options(mapped_options)
        synthetic = None
        if settings.get('__namespace') and row.get('project_files') and Path(project).parent.as_posix() == '/':
            # CLI default root glob would also see read-only runtime mounts.
            # Append the API's explicit file list after existing JSON members,
            # keeping every original diagnostic location and option AST intact.
            path = physical(project)
            from project_scope import append_files
            names = row['project_files']
            text = append_files(path.read_text(), names)
            path.write_text(text)
            write_json(folder / 'project-scope.json', {'files': names, 'original_source_sha256': row['source_sha256']})
        if config_values:
            synthetic = (physical(project).parent if project else cwd) / '__verdict_options__.json'
            if synthetic.exists():
                raise RuntimeError('upstream unit collides with generated option config')
            config = {'compilerOptions': config_values}
            if project:
                config['extends'] = './' + physical(project).name
            else:
                config['files'] = [os.path.relpath(physical(name), cwd) for name in roots]
            if settings.get('__namespace'):
                prefix = str(folder / 'filesystem') + '/'
                config['files'] = [name.replace(prefix, '/') for name in config.get('files', [])]
            write_json(synthetic, config)
            mapped_options = {key: value for key, value in mapped_options.items() if key not in config_values}
        ambient_roots = not project or 'typeRoots' not in effective or 'typeRoots' in options
        if (has_type_packages(units, settings) or effective.get('types')) and 'typeRoots' not in effective:
            # Explicit typeRoots enables a resolver fallback that bypasses
            # package exports. With provided @types, preserve inferred roots.
            ambient_roots = False
        if not ambient_roots:
            mapped_options.pop('typeRoots', None)
        argv = [str(binary), *arguments(mapped_options, [] if project or synthetic else mapped_roots,
                                       ambient_roots, effective if project else None)]
        if project or synthetic:
            argv += ['--project', os.path.relpath(synthetic or physical(project), cwd)]
        write_json(folder / 'working-directory.json', str(cwd))
        if settings.get('__namespace'):
            from namespace import command, virtual_arguments
            from cases import virtual_directory
            argv = command(virtual_arguments(argv, folder), folder, virtual_directory(settings), tree)
        write_json(folder / 'command.json', argv)
        timed_out = False
        with (folder / 'actual.stdout').open('wb') as stdout, (folder / 'actual.stderr').open('wb') as stderr:
            try:
                completed = subprocess.run(argv, cwd=cwd, stdout=stdout, stderr=stderr,
                                           timeout=float(os.environ.get('TSC_TIMEOUT', '60')))
                code = completed.returncode
            except subprocess.TimeoutExpired:
                code, timed_out = 124, True
        (folder / 'actual.exit').write_text(str(code) + '\n')
        # CLI reports outputs-skipped only when emit returns emitSkipped. noEmit
        # itself returns false; noEmitOnError with diagnostics returns true.
        exit_code = expected_exit(row.get('effective_options', options), expected)
        wanted = {'stdout': expected, 'stderr': b'', 'exit': f'{exit_code}\n'.encode()}
        for suffix, value in wanted.items():
            (folder / ('expected.' + suffix)).write_bytes(value)
        diagnostics = map_diagnostics((folder / 'actual.stdout').read_bytes(), folder, cwd, units, settings)
        if row.get('library_placeholders'):
            from library_summary import project as project_library_summary
            diagnostics = project_library_summary(diagnostics, row['library_placeholders'])
        if synthetic:
            # This location belongs to generated plumbing, never a test unit.
            # Preserve option codes/messages; upstream API options have no AST.
            synthetic_name = os.path.relpath(synthetic, cwd).encode()
            diagnostics = re.sub(rb'^' + re.escape(synthetic_name) + rb'\(\d+,\d+\): ', b'', diagnostics, flags=re.M)
        if options.get('pretty'):
            # Upstream stores the diagnostic formatter block before annotated
            # sources and its footer after them. Judge that formatter block.
            diagnostics = pretty_diagnostics(diagnostics)
        (folder / 'actual.diagnostics').write_bytes(diagnostics)
        differences = compare(folder, wanted, {'stdout': diagnostics})
        if timed_out:
            differences['timeout'] = {'seconds': float(os.environ.get('TSC_TIMEOUT', '60'))}
        return {'case': row['source'], 'configuration': row.get('configuration', ''),
                'capture': folder.name, 'differences': differences} if differences else None
    with ThreadPoolExecutor(max_workers=int(os.environ.get('TSC_JOBS', '4'))) as executor:
        failures = [result for result in executor.map(execute, enumerate(rows, 1)) if result]
    report = {'total': len(rows), 'passed': len(rows) - len(failures), 'failed': len(failures),
              'excluded': manifest['excluded'], 'deferred': len(manifest['cases']) - len(rows),
              'selected_inputs': manifest['selected'], 'configurations': len(manifest['cases']),
              'excluded_configurations': len(manifest.get('configuration_exclusions', [])),
              'census_total': manifest['total'], 'exclusion_reasons': manifest['reasons'],
              'failures': failures, 'seconds': round(time.monotonic() - started, 3)}
    write_json(output / 'report.json', report)
    return report


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    mode = parser.add_mutually_exclusive_group(required=True)
    mode.add_argument('--tsc', type=Path, help='one executable, not a shell command')
    mode.add_argument('--compare', nargs=2, type=Path, metavar=('LEFT', 'RIGHT'), help='run two binaries against unchanged expectations and compare their captures')
    parser.add_argument('--baseline-limit', type=int, help='explicit smoke subset; deferred cases remain counted')
    parser.add_argument('output', type=Path)
    args = parser.parse_args()
    output = args.output.resolve()
    binaries = [binary.resolve() for binary in (args.compare or [args.tsc])]
    if any(not binary.is_file() or not os.access(binary, os.X_OK) for binary in binaries):
        parser.error('each compiler must be an executable file')
    if args.baseline_limit is not None and args.baseline_limit < 1:
        parser.error('baseline limit must be positive')
    if output.exists():
        parser.error('output directory must be new')
    output.mkdir(parents=True)
    if args.compare:
        from comparison import compare_compilers
        try:
            return compare_compilers(*binaries, output, args.baseline_limit)
        except Exception as error:
            write_json(output / 'comparison-error.json', {'harness_error': str(error)})
            print(f'Comparison harness error: {error}', file=sys.stderr)
            return 2
    binary = binaries[0]
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
