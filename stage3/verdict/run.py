#!/usr/bin/env python3
"""Compare a candidate CLI with live Node tsc on each selected case."""
import argparse
from concurrent.futures import ThreadPoolExecutor
from contextlib import ExitStack
import hashlib
import json
import os
import posixpath
import re
import stat
from pathlib import Path, PurePosixPath
from queue import Empty, Queue
import subprocess
import sys
from threading import Thread
import time

ROOT = Path(__file__).resolve().parent
DEFAULT_TIMEOUT_SECONDS = 180
DRIVER = ROOT.parent / 'drivers/tsc'
sys.path.insert(0, str(DRIVER))
from corpus import EMIT_PATHS, PIN, materialize, materialize_command, baseline_output
from driver import project_args, project_config, project_inputs, project_rows
from compiler_process import (CLEANUP_SECONDS, OWNER_DEADLINE_EXIT, OWNER_DEADLINE_MESSAGE,
                              read_owner_receipt, write_owner_receipt)


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


def output_tree(root):
    files = {}
    def unreadable(error):
        raise error

    for directory, subdirs, names in os.walk(root, onerror=unreadable):
        for entry in [*(Path(directory) / name for name in subdirs),
                      *(Path(directory) / name for name in names)]:
            mode = entry.lstat().st_mode
            if stat.S_ISDIR(mode):
                continue
            path = entry.relative_to(root).as_posix()
            if stat.S_ISREG(mode):
                files[path] = ('file', entry.read_bytes())
            elif stat.S_ISLNK(mode):
                files[path] = ('symlink', os.fsencode(os.readlink(entry)))
            else:
                files[path] = ('special', str(stat.S_IFMT(mode)).encode())
    return files


def compare_trees(expected, actual):
    differences = {}
    for path in sorted(expected.keys() | actual.keys()):
        wanted = expected.get(path)
        got = actual.get(path)
        if wanted == got:
            continue
        reference_kind, reference = wanted or (None, b'')
        candidate_kind, candidate = got or (None, b'')
        offset = next((index for index, (a, b) in enumerate(zip(reference, candidate)) if a != b),
                      min(len(reference), len(candidate)))
        start = max(0, offset - 8)
        differences[path] = {
            'kind': ('extra' if wanted is None else 'missing' if got is None
                     else 'type' if reference_kind != candidate_kind else 'bytes'),
            'byte_offset': offset, 'hex_offset': start,
            'expected_hex': reference[start:offset + 8].hex(' '),
            'actual_hex': candidate[start:offset + 8].hex(' '),
            'expected_size': len(reference), 'actual_size': len(candidate),
        }
    return differences


def run_compiler(argv, *, cwd, stdout, stderr, timeout, env):
    deadline = time.monotonic() + timeout
    cleanup_deadline = deadline + CLEANUP_SECONDS
    label = f'{env["STAGE3_VERDICT_CASE"]}: {env["STAGE3_VERDICT_SIDE"]}'
    receipt = Path(stdout.name).absolute().with_suffix(".owner.json")
    result = Queue(maxsize=1)
    # The worker owns both capture descriptors and the eventual supervisor,
    # even if the caller returns while Popen is still pending.
    with ExitStack() as setup:
        capture_out = setup.enter_context(os.fdopen(os.dup(stdout.fileno()), 'wb'))
        capture_err = setup.enter_context(os.fdopen(os.dup(stderr.fileno()), 'wb'))

        def supervise():
            try:
                with capture_out, capture_err:
                    code = wait_for_compiler(argv, cwd, capture_out, capture_err, env,
                                             deadline, cleanup_deadline, label, receipt)
            except Exception as error:
                result.put((None, error))
            else:
                result.put((code, None))

        # The detached cleanup owner claims the captures before it can
        # launch a supervisor. This thread only collects its bounded result.
        Thread(target=supervise, daemon=True).start()
        setup.pop_all()

    try:
        code, error = result.get(timeout=max(0, deadline - time.monotonic()))
    except Empty:
        if claim_expired_receipt(receipt):
            raise RuntimeError(deadline_message(label, receipt))
        try:
            code, error = result.get(timeout=max(0, cleanup_deadline - time.monotonic()))
        except Empty as expired:
            raise RuntimeError(deadline_message(label, receipt)) from expired
    else:
        # A worker can report its late refusal as the caller wakes.
        if time.monotonic() >= deadline:
            claim_expired_receipt(receipt)
    if error is not None:
        raise error
    if time.monotonic() >= cleanup_deadline:
        raise RuntimeError(deadline_message(label, receipt))
    if code is None:
        raise subprocess.TimeoutExpired(argv, timeout)
    return subprocess.CompletedProcess(argv, code)


def claim_expired_receipt(receipt):
    try:
        fd = os.open(receipt, os.O_CREAT | os.O_EXCL | os.O_WRONLY, 0o600)
    except FileExistsError:
        return False
    with os.fdopen(fd, 'w') as stream:
        json.dump({'state': 'expired', 'expired_at': time.time_ns()}, stream)
        stream.write('\n')
    return True


def deadline_message(label, receipt):
    message = f'{label} compiler supervisor startup/status exceeded its deadline'
    try:
        owner, proof = read_owner_receipt(receipt)
    except FileNotFoundError:
        return message
    except (OSError, ValueError) as error:
        return f'{message}. Cannot verify cleanup owner receipt {receipt}: {error}'
    if proof == 'expired':
        return f'{message}. No cleanup owner was admitted. Receipt: {receipt}'
    if owner['state'] == 'owned':
        detail = 'owner death proven by identity' if proof == 'dead' else 'liveness not proven'
        return (f'{message}. Cleanup owner acknowledged at {owner["acknowledged_at"]} '
                f'by pid {owner["pid"]}; {detail}. Receipt: {receipt}')
    return message


def release_quarantine(receipt):
    try:
        owner, proof = read_owner_receipt(receipt)
        if proof == 'expired':
            raise ValueError('the receipt expired without an owner. Nothing to release')
        if proof == 'alive':
            raise ValueError('the recorded cleanup owner is alive by identity')
        if proof == 'dead':
            released = dict(owner, state='released', released_by='hand', released_at=time.time_ns())
            if owner['state'] == 'uncertain':
                released['released_from'] = 'uncertain'
            write_owner_receipt(receipt, released)
            if owner['state'] == 'uncertain':
                print("The supervisor's subtree cleanup was never verified.")
    except (OSError, ValueError) as error:
        print(f'Cannot release {receipt}: {error}', file=sys.stderr)
        return 2
    print(f'Release recorded in {receipt}')
    return 0


def wait_for_compiler(argv, cwd, stdout, stderr, env, deadline, cleanup_deadline, label, receipt):
    command = [sys.executable, '-B', str(ROOT / 'compiler_process.py'), '--owner',
               str(receipt), str(deadline), str(cleanup_deadline), str(stdout.fileno()), *argv]
    owner = subprocess.Popen(command, cwd=cwd, stdin=subprocess.DEVNULL, stdout=subprocess.PIPE,
                             stderr=stderr, env=env, pass_fds=(stdout.fileno(),), start_new_session=True)
    try:
        status, _ = owner.communicate(timeout=max(0, cleanup_deadline - time.monotonic()))
    except subprocess.TimeoutExpired as error:
        # The owner has its own lifetime and status pipe to its supervisor.
        # Closing this pipe cannot interrupt the owner's cleanup.
        owner.stdout.close()
        raise RuntimeError(deadline_message(label, receipt)) from error
    if owner.returncode == OWNER_DEADLINE_EXIT:
        raise RuntimeError(f'{label} {OWNER_DEADLINE_MESSAGE}')
    if time.monotonic() >= cleanup_deadline:
        raise RuntimeError(deadline_message(label, receipt))
    if owner.returncode:
        raise RuntimeError(f'{label} compiler supervision failed ({owner.returncode})')
    return json.loads(status)


def run_side(command, folder, project, files, options, timeout, suite):
    folder.mkdir(parents=True, exist_ok=True)
    project.mkdir()
    for name, content in files.items():
        (project / name).write_text(content)
    write_json(project / 'tsconfig.json', project_config(files, options))
    initial = output_tree(project)
    argv = command + project_args(options)
    write_json(folder / 'command.json', argv)
    write_json(folder / 'cwd.json', str(project.absolute()))
    environment = dict(os.environ, STAGE3_VERDICT_SUITE=suite,
                       STAGE3_VERDICT_CASE=folder.parent.name, STAGE3_VERDICT_SIDE=folder.name)
    timed_out = False
    with (folder / 'actual.stdout').open('wb') as stdout, (folder / 'actual.stderr').open('wb') as stderr:
        try:
            code = run_compiler(argv, cwd=project, stdout=stdout, stderr=stderr,
                                timeout=timeout, env=environment).returncode
        except subprocess.TimeoutExpired:
            code, timed_out = 124, True
    (folder / 'actual.exit').write_text(str(code) + '\n')
    artifacts = None if timed_out else output_tree(project)
    archive = folder / 'project'
    if project != archive:
        project.rename(archive)
    return initial, artifacts, timed_out


def validate_emit_paths(options):
    for option in EMIT_PATHS:
        if option not in options:
            continue
        value = options[option]
        # TypeScript's normalizeSlashes precedes root and component handling.
        path = value.replace('\\', '/')
        if path.startswith('/') or re.match(r'^[A-Za-z]:', path) or '://' in path or '..' in path.split('/'):
            raise RuntimeError(f'output directive @{option} escapes the case: {value}')


def resolve_config_path(value, project_path):
    value = value.replace('\\', '/')
    # TS recognizes the prefix case-insensitively, then replaces the literal token.
    if value.lower().startswith('${configdir}'):
        value = value.replace('${configDir}', './', 1)
    if re.match(r'^[A-Za-z]:', value) or '://' in value:
        raise RuntimeError(f'cannot resolve a non-POSIX path in the verdict project: {value}')
    return posixpath.normpath(posixpath.join(project_path, value))


def relative_build_info_stem(root, stem):
    # TypeScript compares normalized POSIX path components ignoring case,
    # then appends the original destination components after the shared prefix.
    source = PurePosixPath(root).parts
    target = PurePosixPath(stem).parts
    common = 0
    for left, right in zip(source, target):
        if left.upper() != right.upper():
            break
        common += 1
    return '/'.join(('..',) * (len(source) - common) + target[common:])


def no_emit_build_info(options, project):
    if options.get('noEmit') is not True or not (options.get('incremental') or options.get('composite')):
        return None

    # The generated tsconfig.json supplies the config directory. TypeScript
    # resolves file options before getTsBuildInfoEmitOutputFilePath sees them.
    project_path = project.absolute().as_posix()
    if options.get('tsBuildInfoFile'):
        path = resolve_config_path(options['tsBuildInfoFile'], project_path)
    elif options.get('outFile'):
        stem = resolve_config_path(options['outFile'], project_path)
        for extension in ('.d.ts', '.d.mts', '.d.cts', '.mjs', '.mts', '.cjs',
                          '.cts', '.ts', '.js', '.tsx', '.jsx', '.json'):
            if len(stem) > len(extension) and stem.endswith(extension):
                stem = stem[:-len(extension)]
                break
        path = stem + '.tsbuildinfo'
    else:
        stem = posixpath.join(project_path, 'tsconfig')
        if options.get('outDir'):
            directory = resolve_config_path(options['outDir'], project_path)
            relative_stem = 'tsconfig'
            if options.get('rootDir'):
                root = resolve_config_path(options['rootDir'], project_path)
                relative_stem = relative_build_info_stem(root, stem)
            stem = posixpath.join(directory, relative_stem)
        path = stem + '.tsbuildinfo'
    relative = posixpath.relpath(path, project_path)
    validate_emit_paths({'tsBuildInfoFile': relative})
    return relative


def no_emit_artifacts(side, original, tree, build_info, differences):
    allowed = dict(original)
    if build_info is not None:
        entry = tree.get(build_info)
        if entry is None or entry[0] != 'file' or build_info in original:
            differences[side + '_build_info'] = {'path': build_info, 'expected': 'new regular file'}
        else:
            allowed[build_info] = entry
    forbidden = compare_trees(allowed, tree)
    if forbidden:
        differences[side + '_files'] = forbidden


def compare_case(binary, reference, folder, files, options, legacy=None, normalize_legacy=False, *,
                 suite='', candidate_options=None):
    if candidate_options is None:
        candidate_options = options
    validate_emit_paths(options)
    validate_emit_paths(candidate_options)
    timeout = float(os.environ.get('TSC_TIMEOUT', DEFAULT_TIMEOUT_SECONDS))
    expected = folder / 'reference'
    actual = folder / 'candidate'
    # A rootDir above the project can make the scratch directory name part
    # of the output path. Run incremental noEmit sides at the same location
    # in sequence, archiving each tree before starting the next invocation.
    shared_project = folder / 'project'
    reference_build_info = no_emit_build_info(options, shared_project)
    candidate_build_info = no_emit_build_info(candidate_options, shared_project)
    if reference_build_info is not None or candidate_build_info is not None:
        expected_root = actual_root = shared_project
    else:
        expected_root = expected / 'project'
        actual_root = actual / 'project'
    reference_inputs, expected_files, reference_timeout = run_side(
        reference, expected, expected_root, files, options, timeout, suite)
    if reference_timeout:
        return {'timeout': {'side': 'Node', 'seconds': timeout}}
    candidate_inputs, actual_files, candidate_timeout = run_side(
        [str(binary)], actual, actual_root, files, candidate_options, timeout, suite)
    if candidate_timeout:
        return {'timeout': {'side': 'native', 'seconds': timeout}}
    wanted = {stream: (expected / ('actual.' + stream)).read_bytes() for stream in ('stdout', 'stderr', 'exit')}
    # The two project roots stand for the same source directory in diagnostics.
    wanted['stdout'] = baseline_diagnostics(wanted['stdout'], expected_root)
    diagnostics = baseline_diagnostics((actual / 'actual.stdout').read_bytes(), actual_root)
    differences = compare(actual, wanted, {'stdout': diagnostics})
    if legacy is not None:
        legacy_differences = compare(actual, legacy, {'stdout': diagnostics} if normalize_legacy else None)
        if legacy_differences:
            differences['legacy'] = legacy_differences
    if options.get('noEmit') is True:
        no_emit_artifacts('reference', reference_inputs, expected_files,
                          reference_build_info, differences)
    if candidate_options.get('noEmit') is True:
        no_emit_artifacts('candidate', candidate_inputs, actual_files,
                          candidate_build_info, differences)
    # Leave out inputs only when both sides left them alone. In particular,
    # distinct untouched tsconfig files are inputs, while a removal on one
    # side must still be visible in the cross-side comparison.
    unchanged = {path for path in expected_files.keys() & actual_files.keys()
                 if expected_files[path] == reference_inputs.get(path)
                 and actual_files[path] == candidate_inputs.get(path)}
    tree_differences = compare_trees(
        {path: entry for path, entry in expected_files.items() if path not in unchanged},
        {path: entry for path, entry in actual_files.items() if path not in unchanged},
    )
    if tree_differences:
        differences['files'] = tree_differences
    return differences


def measure_cases(rows, execute):
    with ThreadPoolExecutor(max_workers=int(os.environ.get('TSC_JOBS', '4'))) as executor:
        results = list(executor.map(execute, rows))
    failures = []
    timeouts = []
    for outcome, _ in results:
        if outcome is None:
            continue
        timing = outcome['differences'].get('timeout')
        if timing is None:
            failures.append(outcome)
        else:
            timeouts.append({'case': outcome['case'], **timing})
    failed = len(failures) + len(timeouts)
    no_emit = sum(no_emit for _, no_emit in results)
    return {'total': len(results), 'passed': len(results) - failed, 'failed': failed,
            'no_emit': no_emit, 'emit_enabled': len(results) - no_emit,
            'failures': failures, 'timeouts': timeouts}


def driver_suite(binary, reference, output, tiny):
    started = time.monotonic()
    output.mkdir()
    rows = project_rows(tiny)
    expected_count = 1 if tiny else 301
    if len(rows) != expected_count:
        raise RuntimeError(f'driver population changed: {len(rows)} != {expected_count}')

    def execute(item):
        identifier, row, source = item
        files, options = project_inputs(identifier, row, source)
        legacy = {stream: (source / ('golden.' + stream)).read_bytes()
                  for stream in ('stdout', 'stderr', 'exit')}
        differences = compare_case(binary, reference, output / identifier, files, options, legacy,
                                   suite="tiny" if tiny else "acceptance")
        failure = {'case': identifier, 'differences': differences} if differences else None
        return failure, options.get('noEmit') is True

    report = measure_cases(rows, execute)
    report.update(excluded=0, seconds=round(time.monotonic() - started, 3))
    write_json(output / 'report.json', report)
    return report


def upstream_tree():
    """The pinned upstream checkout, as apply reads it, with the verdict's own override first: STAGE3_VERDICT_UPSTREAM,
    else the gate inputs' TypeScript checkout, ADAMIC_TYPESCRIPT_SOURCE. The verdict never clones one, from a mirror or the network, so a runner without
    either refuses plainly. (cohere/TypeScript is Adamic's fork, not the pinned upstream: it holds no tests/.)"""
    configured = os.environ.get('STAGE3_VERDICT_UPSTREAM') or os.environ.get('ADAMIC_TYPESCRIPT_SOURCE')
    if not configured:
        raise RuntimeError(f'set ADAMIC_TYPESCRIPT_SOURCE to the gate inputs\' TypeScript v6.0.3 checkout at {PIN}, '
                           'or STAGE3_VERDICT_UPSTREAM to another checkout at it; the verdict never clones one')
    tree = Path(configured).resolve()
    head = subprocess.run(['git', '-C', str(tree), 'rev-parse', 'HEAD'], capture_output=True, text=True).stdout.strip()
    if head != PIN:
        raise RuntimeError(f'upstream checkout {tree} is at {head or "no commit"}, not the pinned {PIN}')
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


def baseline_inputs(tree, row):
    raw = (tree / row['source']).read_bytes()
    if hashlib.sha256(raw).hexdigest() != row['source_sha256']:
        raise RuntimeError(f'source hash mismatch: {row["source"]}')
    if materialize(raw)[1] != row['options']:
        raise RuntimeError(f'header options changed: {row["source"]}')
    content, options = materialize_command(raw)
    expected = b''
    if row['baseline']:
        raw_baseline = (tree / row['baseline']).read_bytes()
        if hashlib.sha256(raw_baseline).hexdigest() != row['baseline_sha256']:
            raise RuntimeError(f'baseline hash mismatch: {row["baseline"]}')
        expected = baseline_output(raw_baseline)
    if hashlib.sha256(expected).hexdigest() != row['expected_sha256']:
        raise RuntimeError(f'baseline summary changed: {row["source"]}')
    legacy = {'stdout': expected, 'stderr': b'', 'exit': b'2\n' if expected else b'0\n'}
    return row, content, options, legacy


def baseline_suite(binary, tree, output, limit, reference=None):
    started = time.monotonic()
    output.mkdir()
    manifest = json.loads((ROOT / 'selection.json').read_text())
    validate_manifest(manifest)
    rows = manifest['cases'] if limit is None else manifest['cases'][:limit]
    if not rows:
        raise RuntimeError('empty baseline suite')
    write_json(output / 'exclusions.json', manifest['exclusions'])
    write_json(output / 'selection.json', rows)
    # Validate the pinned inputs before running either compiler. The artifact
    # audit exercises these guards without needing a Node build.
    prepared = [baseline_inputs(tree, row) for row in rows]
    if reference is None:
        raise RuntimeError('baseline comparison requires a Node reference')

    def execute(item):
        index, (row, content, options, legacy) = item
        folder = output / f'{index:05d}_{Path(row["source"]).stem}'
        differences = compare_case(binary, reference, folder, {row['name']: content}, options,
                                   legacy, normalize_legacy=True, suite="baselines")
        failure = {'case': row['source'], 'capture': folder.name, 'differences': differences} if differences else None
        return failure, options.get('noEmit') is True

    report = measure_cases(enumerate(prepared, 1), execute)
    report.update(excluded=manifest['excluded'], deferred=manifest['selected'] - len(rows),
                  census_total=manifest['total'], exclusion_reasons=manifest['reasons'],
                  seconds=round(time.monotonic() - started, 3))
    write_json(output / 'report.json', report)
    return report


def timeout_message(suite, outcome):
    return f'Timeout in {suite}: {outcome["case"]} ({outcome["side"]}, {outcome["seconds"]:g}s)'


def main():
    if sys.platform != 'linux':
        print(f"Stage 3 verdict runs on Loom's Linux boxes and requires the Linux "
              f"child supervisor (unsupported host: {sys.platform}).", file=sys.stderr)
        raise SystemExit(2)
    if sys.argv[1:2] == ['release']:
        parser = argparse.ArgumentParser(prog=f'{sys.argv[0]} release',
                                         description='Record a proven quarantine release.')
        parser.add_argument('receipt', type=Path)
        return release_quarantine(parser.parse_args(sys.argv[2:]).receipt)
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--tsc', required=True, type=Path, help='one executable, not a shell command')
    parser.add_argument('--node-tsc', type=Path, default=os.environ.get('STAGE3_VERDICT_NODE_TSC'),
                        help='adapted tsc.js (default: STAGE3_VERDICT_NODE_TSC)')
    parser.add_argument('--baseline-limit', type=int, help='explicit smoke subset; deferred cases remain counted')
    parser.add_argument('output', type=Path)
    args = parser.parse_args()
    binary, output = args.tsc.resolve(), args.output.resolve()
    if args.node_tsc is None or not args.node_tsc.is_file():
        parser.error('set --node-tsc or STAGE3_VERDICT_NODE_TSC to the adapted Node tsc.js')
    reference = ['node', str(args.node_tsc.resolve())]
    if not binary.is_file() or not os.access(binary, os.X_OK):
        parser.error('tsc must be an executable file')
    if args.baseline_limit is not None and args.baseline_limit < 1:
        parser.error('baseline limit must be positive')
    if output.exists():
        parser.error('output directory must be new')
    output.mkdir(parents=True)
    suites = {}
    errors = []
    for name, task in [('acceptance', lambda: driver_suite(binary, reference, output / 'acceptance', False)),
                       ('tiny', lambda: driver_suite(binary, reference, output / 'tiny', True)),
                       ('baselines', lambda: baseline_suite(binary, upstream_tree(), output / 'baselines', args.baseline_limit, reference))]:
        print(f'running {name}', flush=True)
        try:
            suites[name] = task()
            print(f'{name}: {suites[name]["passed"]}/{suites[name]["total"]}', flush=True)
            for timeout in suites[name]['timeouts']:
                print(timeout_message(name, timeout), flush=True)
        except Exception as error:
            errors.append({'suite': name, 'error': str(error)})
            suites[name] = {'status': 'harness_error', 'error': str(error)}
    success = not errors and all(row['failed'] == 0 for row in suites.values())
    summary = {'schema_version': 3, 'tsc': str(binary), 'upstream_commit': PIN, 'reference': reference,
               'success': success, 'suites': suites, 'harness_errors': errors}
    write_json(output / 'summary.json', summary)
    lines = ['# Stage 3 verdict', '', '| Suite | Pass | Fail | Timeouts | Excluded | Deferred | Emit enabled | No emit |',
             '|---|---:|---:|---:|---:|---:|---:|---:|']
    for name, row in suites.items():
        if 'error' in row:
            lines.append(f'| {name} | harness error | | | | | | |')
        else:
            lines.append(f'| {name} | {row["passed"]} | {row["failed"]} | {len(row["timeouts"])} | {row["excluded"]} | {row.get("deferred", 0)} | {row["emit_enabled"]} | {row["no_emit"]} |')
    lines.extend(['', 'Acceptance includes tiny; the separate tiny row repeats that project.', ''])
    lines.extend(['Timeouts are included in Fail.', ''])
    for name, row in suites.items():
        for timeout in row.get('timeouts', ()):
            lines.extend([timeout_message(name, timeout), ''])
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
