#!/usr/bin/env python3
"""The three independent byte checks must each notice its own mutation."""
from contextlib import nullcontext, redirect_stderr, redirect_stdout
import errno
from io import StringIO
from pathlib import Path
import json
import os
import signal
import subprocess
import sys
import tempfile
from threading import Event, Thread, get_native_id
from types import SimpleNamespace
import time
from textwrap import dedent
import unittest
import copy
import re
from unittest.mock import patch
from census import diagnostic_sources
import compiler_process
from run import (PIN, baseline_diagnostics, claim_expired_receipt, compare, compare_case, compare_trees,
                 first_difference, main, measure_cases, no_emit_build_info, output_tree, run_compiler,
                 release_quarantine, upstream_tree, validate_manifest)
from corpus import EMIT_PATHS, materialize_command
from driver import project_args, project_config


class HostSupport(unittest.TestCase):
    def test_refuses_non_linux_before_parsing_arguments(self):
        for platform in ('darwin', 'win32', 'freebsd'):
            with self.subTest(platform=platform), patch('run.sys.platform', platform), \
                    patch('run.argparse.ArgumentParser', side_effect=AssertionError('parsed unsupported host')) as parser, redirect_stderr(StringIO()) as stderr:
                with self.assertRaises(SystemExit) as exit_context:
                    main()
                self.assertEqual(exit_context.exception.code, 2)
                parser.assert_not_called()
                self.assertIn("Loom's Linux boxes", stderr.getvalue())
                self.assertIn('requires the Linux child supervisor', stderr.getvalue())
                self.assertIn(platform, stderr.getvalue())


class Comparisons(unittest.TestCase):
    def test_independent_stream_mutants(self):
        expected = {'stdout': b'f.ts(1,1): error TS2322: bad\n', 'stderr': b'', 'exit': b'2\n'}
        with tempfile.TemporaryDirectory() as scratch:
            folder = Path(scratch)
            for mutant in ('stdout', 'stderr', 'exit'):
                for stream, value in expected.items():
                    actual = value
                    if stream == mutant:
                        actual = value.replace(b'error', b'Error', 1) if stream == 'stdout' else (b'x' if stream == 'stderr' else b'1\n')
                    (folder / ('actual.' + stream)).write_bytes(actual)
                self.assertEqual(set(compare(folder, expected)), {mutant})

    def test_baseline_root_mapping_and_filename_mutant(self):
        folder = Path('/scratch/case')
        expected = b"case.ts(1,1): error TS2694: Namespace '\"case\".N' has no exported member 'X'.\n"
        actual = expected.replace(b'"case"', b'"/scratch/case/case"')
        self.assertEqual(baseline_diagnostics(actual, folder), expected)
        mutant = actual.replace(b'/scratch/case/case"', b'/scratch/case/casE"')
        self.assertNotEqual(baseline_diagnostics(mutant, folder), expected)
        unrelated = actual.replace(b'/scratch/case/', b'/scratch/other/')
        self.assertNotEqual(baseline_diagnostics(unrelated, folder), expected)

    def test_library_placeholder_census_mutant(self):
        summary = b"f.ts(1,1): error TS2374: duplicate.\nlib.es5.d.ts(--,--): error TS2374: duplicate.\n"
        wanted = ['f.ts', 'lib.es5.d.ts']
        self.assertEqual(diagnostic_sources(summary), wanted)
        old_guard = re.compile(rb'^([^\n]+?)\(\d+,\d+\):', re.M)
        with patch('census.DIAGNOSTIC_SOURCES', old_guard):
            with self.assertRaises(AssertionError):
                self.assertEqual(diagnostic_sources(summary), wanted)

    def test_census_mutants(self):
        manifest = {'upstream_commit': PIN, 'total': 2, 'selected': 1, 'excluded': 1,
                    'cases': [{'source': 'a.ts'}], 'exclusions': [{'source': 'b.ts'}]}
        validate_manifest(manifest)
        for mutation in ('pin', 'selected', 'excluded', 'duplicate'):
            changed = copy.deepcopy(manifest)
            if mutation == 'pin':
                changed['upstream_commit'] = 'wrong'
            elif mutation == 'duplicate':
                changed['exclusions'][0]['source'] = 'a.ts'
            else:
                changed[mutation] += 1
            with self.assertRaisesRegex(RuntimeError, 'invalid baseline census'):
                validate_manifest(changed)

    def test_eof_difference(self):
        self.assertEqual(first_difference(b'abc', b'ab')['byte_offset'], 2)
        self.assertEqual(first_difference(b'', b'x')['byte_offset'], 0)


class EmitComparisons(unittest.TestCase):
    def test_directives_reach_both_commands(self):
        raw = (b'// @noEmit: true\n// @declaration: true\n// @sourceMap: true\n'
               b'// @outDir: nested/out\n// @outFile: bundle.js\nconst x = 1;\n')
        content, options = materialize_command(raw)
        self.assertEqual(content, 'const x = 1;\n')
        self.assertEqual(options, dict(noEmit=True, declaration=True, sourceMap=True,
                                       outDir='nested/out', outFile='bundle.js'))
        self.assertEqual(project_config({'x.ts': content}, options)['compilerOptions']['outDir'], 'nested/out')
        self.assertIn('--noEmit', project_args(options))
        self.assertNotIn('--noEmit', project_args({}))
        self.assertNotIn('--noEmit', project_args({'noEmit': False}))
        with self.assertRaisesRegex(ValueError, '@unknownEmit: true'):
            materialize_command(b'// @unknownEmit: true\nx;')
        with self.assertRaisesRegex(RuntimeError, '@outDir.*escapes'):
            compare_case('candidate', ['reference'], Path('unused'), {}, {'outDir': '../elsewhere'})

    def test_emit_path_validation_precedes_both_invocations(self):
        for directive in EMIT_PATHS:
            for separator in ('/', chr(92)):
                # A normalized parent component is rejected by policy even
                # with an in-project result. Neither compiler may be called.
                path = separator.join(('nested', '..', 'out'))
                with self.subTest(directive=directive, separator=separator), \
                        patch('run.run_side', side_effect=AssertionError('compiler invoked before validation')) as invoke:
                    with self.assertRaisesRegex(RuntimeError, f'@{directive}.*escapes'):
                        compare_case('candidate', ['reference'], Path('unused'), {}, {directive: path})
                    invoke.assert_not_called()

    def test_rooted_emit_paths_are_rejected_before_run(self):
        for path in ('/out', r'\out', 'C:/out', 'file://host/out'):
            with self.subTest(path=path), patch('run.run_side') as invoke:
                with self.assertRaisesRegex(RuntimeError, '@outDir.*escapes'):
                    compare_case('candidate', ['reference'], Path('unused'), {}, {'outDir': path})
                invoke.assert_not_called()

    def test_mixed_no_emit_compares_outputs_and_keeps_side_checks(self):
        for side in ('reference', 'candidate'):
            for extra in (False, True):
                def compile(command, cwd, **kwargs):
                    if extra and command[0] == side:
                        (cwd / 'actual.js').write_text('extra output')
                    return subprocess.CompletedProcess(command, 0)

                with self.subTest(no_emit=side, extra=extra), tempfile.TemporaryDirectory() as scratch, \
                        patch('run.run_compiler', compile):
                    result = compare_case(
                        'candidate', ['reference'], Path(scratch), {'input.d.ts': 'declare const n: number;'},
                        {'noEmit': side == 'reference'}, candidate_options={'noEmit': side == 'candidate'},
                    )
                    if not extra:
                        self.assertEqual(result, {})
                    else:
                        self.assertEqual(set(result), {'files', f'{side}_files'})
                        self.assertEqual(set(result['files']), {'actual.js'})
                        self.assertEqual(result['files']['actual.js']['kind'],
                                         'extra' if side == 'candidate' else 'missing')
                        self.assertEqual(result[f'{side}_files']['actual.js']['kind'], 'extra')

    def test_reference_build_info_must_be_regular_and_present(self):
        for symlink in (False, True):
            def compile(command, cwd, **kwargs):
                path = cwd / 'tsconfig.tsbuildinfo'
                if command[0] == 'candidate':
                    path.write_bytes(b'build info')
                elif symlink:
                    path.symlink_to('input.ts')
                return subprocess.CompletedProcess(command, 0)

            with self.subTest(symlink=symlink), tempfile.TemporaryDirectory() as scratch, \
                    patch('run.run_compiler', compile):
                folder = Path(scratch)
                result = compare_case('candidate', ['reference'], folder, {'input.ts': ''},
                                      {'noEmit': True, 'incremental': True})
                self.assertEqual(result['reference_build_info']['path'], 'tsconfig.tsbuildinfo')
                self.assertNotIn('candidate_build_info', result)
                self.assertEqual(json.loads((folder / 'reference/cwd.json').read_text()),
                                 str(folder / 'project'))
                self.assertEqual(json.loads((folder / 'candidate/cwd.json').read_text()),
                                 str(folder / 'project'))

    def test_extra_file(self):
        diff = compare_trees({}, {'extra.js': ('file', b'x')})
        self.assertEqual(diff['extra.js']['kind'], 'extra')
        self.assertEqual(diff['extra.js']['byte_offset'], 0)
        self.assertEqual(diff['extra.js']['actual_hex'], '78')

    def test_missing_file(self):
        diff = compare_trees({'nested/out.js': ('file', b'x')}, {})
        self.assertEqual(diff['nested/out.js']['kind'], 'missing')
        self.assertEqual(diff['nested/out.js']['byte_offset'], 0)
        self.assertEqual(diff['nested/out.js']['expected_hex'], '78')

    def test_javascript_byte(self):
        diff = compare_trees({'out.js': ('file', b'abc')}, {'out.js': ('file', b'abZ')})
        self.assertEqual(diff['out.js']['byte_offset'], 2)
        self.assertEqual(diff['out.js']['actual_hex'], '61 62 5a')

    def test_declaration_byte_and_other_formats(self):
        expected = {'nested/out.d.ts': ('file', b'abc'), 'out.js.map': ('file', b'{}'),
                    'out.tsbuildinfo': ('file', b'[]')}
        actual = {name: (kind, value[:-1]) for name, (kind, value) in expected.items()}
        diff = compare_trees(expected, actual)
        self.assertEqual(set(diff), set(expected))
        self.assertEqual(diff['nested/out.d.ts']['byte_offset'], 2)
        self.assertEqual(diff['out.js.map']['byte_offset'], 1)

    def test_recursive_tree_compares_symlink_without_following(self):
        with tempfile.TemporaryDirectory() as scratch:
            root = Path(scratch)
            (root / 'nested').mkdir()
            (root / 'nested/out.js.map').write_bytes(b'map')
            (root / 'link').symlink_to('nested', target_is_directory=True)
            self.assertEqual(output_tree(root), {'nested/out.js.map': ('file', b'map'),
                                                 'link': ('symlink', b'nested')})

    @staticmethod
    def fake_cli(command, cwd, stdout, stderr, timeout, env):
        if command[0] == 'slow':
            raise subprocess.TimeoutExpired(command, timeout)
        if command[0] == 'bad-no-emit':
            (cwd / 'wrong.js').write_bytes(b'file')
        return subprocess.CompletedProcess(command, 0)

    def test_no_emit_write(self):
        with tempfile.TemporaryDirectory() as scratch, patch('run.run_compiler', self.fake_cli):
            diff = compare_case('bad-no-emit', ['node'], Path(scratch), {'input.ts': ''}, {'noEmit': True})
            self.assertEqual(diff['files']['wrong.js']['kind'], 'extra')
            self.assertEqual(set(diff), {'files', 'candidate_files'})

    def test_no_emit_build_info_paths(self):
        project = Path('/cases/reference/project')
        settings = {'noEmit': True, 'incremental': True}
        self.assertIsNone(no_emit_build_info({'noEmit': True}, project))
        self.assertEqual(no_emit_build_info(settings, project), 'tsconfig.tsbuildinfo')
        self.assertEqual(no_emit_build_info({'noEmit': True, 'composite': True,
                                             'tsBuildInfoFile': r'cache\state.bin'}, project),
                         'cache/state.bin')
        self.assertEqual(no_emit_build_info({**settings, 'outFile': 'out/bundle.js'}, project),
                         'out/bundle.tsbuildinfo')
        self.assertEqual(no_emit_build_info({**settings, 'outFile': '.js'}, project), '.tsbuildinfo')
        self.assertEqual(no_emit_build_info({**settings, 'tsBuildInfoFile': '${configDir}/cache/state.bin'}, project),
                         'cache/state.bin')
        self.assertEqual(no_emit_build_info({**settings, 'outDir': 'out'}, project),
                         'out/tsconfig.tsbuildinfo')
        self.assertEqual(no_emit_build_info({**settings, 'outDir': 'out/sub', 'rootDir': 'src'}, project),
                         'out/tsconfig.tsbuildinfo')
        mixed_case = Path('/Cases/MixedCase/reference/project')
        self.assertEqual(no_emit_build_info({**settings, 'outDir': 'out', 'rootDir': '/cases'}, mixed_case),
                         'out/MixedCase/reference/project/tsconfig.tsbuildinfo')

    def compare_build_info_layout(self, reference_path, candidate_path, moved_to=None):
        def emit(command, cwd, **kwargs):
            options = json.loads((cwd / 'tsconfig.json').read_text())['compilerOptions']
            path = cwd / options['tsBuildInfoFile']
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(b'identical build info')
            if command[0] == 'candidate' and moved_to is not None:
                destination = cwd / moved_to
                destination.parent.mkdir(parents=True, exist_ok=True)
                path.rename(destination)
            return subprocess.CompletedProcess(command, 0)

        settings = {'noEmit': True, 'incremental': True}
        with tempfile.TemporaryDirectory() as scratch, patch('run.run_compiler', emit):
            return compare_case('candidate', ['node'], Path(scratch), {'input.ts': ''},
                                {**settings, 'tsBuildInfoFile': reference_path},
                                candidate_options={**settings, 'tsBuildInfoFile': candidate_path})

    def test_no_emit_build_info_matching_layout(self):
        self.assertEqual(self.compare_build_info_layout('state.bin', 'state.bin'), {})

    def test_no_emit_build_info_moved_to_reference_path(self):
        result = self.compare_build_info_layout('state.bin', 'other/state.bin', moved_to='state.bin')
        self.assertEqual(result['candidate_build_info']['path'], 'other/state.bin')
        self.assertEqual(result['candidate_files']['state.bin']['kind'], 'extra')
        self.assertNotIn('files', result)  # Cross-side bytes and paths alone match.

    def test_no_emit_build_info_custom_path_requires_matching_reference(self):
        self.assertEqual(self.compare_build_info_layout('custom/state.bin', 'custom/state.bin'), {})
        result = self.compare_build_info_layout('state.bin', 'custom/state.bin')
        self.assertEqual(set(result), {'files'})
        self.assertEqual({path: diff['kind'] for path, diff in result['files'].items()},
                         {'state.bin': 'missing', 'custom/state.bin': 'extra'})

    def test_no_emit_build_info_and_nothing_else(self):
        def fake_cli(command, cwd, stdout, stderr, timeout, env):
            output = cwd / 'cache' / 'state.bin'
            output.parent.mkdir()
            if command[0] != 'missing':
                output.write_bytes(b'expected' if command[0] != 'changed' else b'changed')
            if command[0] in ('extra', 'node-extra'):
                (cwd / 'wrong.d.ts').write_bytes(b'forbidden')
            return subprocess.CompletedProcess(command, 0)

        settings = {'noEmit': True, 'composite': True, 'tsBuildInfoFile': 'cache/state.bin'}
        for candidate, oracle, expected in [
            ('good', 'node', {}),
            ('missing', 'node', {'files': {'cache/state.bin': 'missing'},
                                 'candidate_build_info': {'path': 'cache/state.bin', 'expected': 'new regular file'}}),
            ('changed', 'node', {'files': {'cache/state.bin': 'bytes'}}),
            ('extra', 'node', {'files': {'wrong.d.ts': 'extra'}, 'candidate_files': {'wrong.d.ts': 'extra'}}),
            ('extra', 'node-extra', {'candidate_files': {'wrong.d.ts': 'extra'},
                                     'reference_files': {'wrong.d.ts': 'extra'}}),
        ]:
            with self.subTest(candidate=candidate, oracle=oracle), tempfile.TemporaryDirectory() as scratch, \
                    patch('run.run_compiler', fake_cli):
                result = compare_case(candidate, [oracle], Path(scratch), {'input.ts': ''}, settings)
                self.assertEqual({check: ({path: item['kind'] for path, item in paths.items()}
                                          if check.endswith('files') else paths)
                                  for check, paths in result.items()}, expected)

    def test_legacy_rejects_identical_live_streams(self):
        with tempfile.TemporaryDirectory() as scratch, patch('run.run_compiler', self.fake_cli):
            diff = compare_case('candidate', ['node'], Path(scratch), {'input.ts': ''}, {},
                                {'stdout': b'old diagnostic\n', 'stderr': b'', 'exit': b'0\n'})
            self.assertEqual(set(diff), {'legacy'})
            self.assertEqual(set(diff['legacy']), {'stdout'})
            self.assertEqual(diff['legacy']['stdout']['byte_offset'], 0)

    def test_timeouts_count_one_case_without_a_mismatch(self):
        for candidate, reference, side in [('slow', 'node', 'native'), ('candidate', 'slow', 'Node')]:
            with self.subTest(side=side), tempfile.TemporaryDirectory() as scratch, \
                    patch('run.run_compiler', self.fake_cli), \
                    patch.dict(os.environ, {'TSC_TIMEOUT': '1'}):
                root = Path(scratch)
                def execute(case):
                    binary, oracle = (candidate, reference) if case == 'slow-case' else ('candidate', 'node')
                    diff = compare_case(binary, [oracle], root / case, {'input.ts': ''}, {})
                    if case == 'slow-case':
                        self.assertEqual(diff, {'timeout': {'side': side, 'seconds': 1}})
                    return ({'case': case, 'differences': diff} if diff else None), False
                report = measure_cases(('slow-case', 'next-case'), execute)
                self.assertEqual((report['total'], report['passed'], report['failed']), (2, 1, 1))
                self.assertEqual(report['failures'], [])
                self.assertEqual(report['timeouts'], [{'case': 'slow-case', 'side': side, 'seconds': 1}])
                if side == 'Node':
                    self.assertFalse((root / 'slow-case' / 'candidate').exists())

    def test_default_timeout_is_180_seconds(self):
        with tempfile.TemporaryDirectory() as scratch, patch('run.run_compiler', self.fake_cli), \
                patch.dict(os.environ):
            os.environ.pop('TSC_TIMEOUT', None)
            diff = compare_case('candidate', ['slow'], Path(scratch), {'input.ts': ''}, {})
            self.assertEqual(diff, {'timeout': {'side': 'Node', 'seconds': 180}})


class DeadlineBoundaries(unittest.TestCase):
    def test_proc_scan_cannot_confirm_cleanup_after_deadline(self):
        for seconds, empty in ((4, True), (6, True), (6, False)):
            clock = [0]

            def scan():
                clock[0] += seconds
                return ()

            last = ChildProcessError() if empty else (0, 0)
            with self.subTest(seconds=seconds, empty=empty), \
                    patch.object(compiler_process.os, 'waitpid', side_effect=[(0, 0), last]), \
                    patch.object(compiler_process, 'Path', return_value=SimpleNamespace(iterdir=scan)), \
                    patch.object(compiler_process.time, 'monotonic', side_effect=lambda: clock[0]), \
                    patch.object(compiler_process.time, 'sleep'):
                if seconds < 5:
                    compiler_process.reap_descendants(5)
                else:
                    with self.assertRaisesRegex(TimeoutError, 'before the deadline'):
                        compiler_process.reap_descendants(5)
                self.assertEqual(clock[0], seconds)

    def test_pending_launch_is_owned_after_caller_deadline(self):
        entered = Event()
        release = Event()
        finished = Event()
        workers = []
        captures = []

        def track_worker(**kwargs):
            worker = Thread(**kwargs)
            workers.append(worker)
            return worker

        def launch(command, *, stderr, pass_fds, **kwargs):
            captures.extend((pass_fds[0], stderr.fileno()))
            entered.set()
            release.wait(2)

            def communicate(*, timeout):
                for fd in captures:
                    os.fstat(fd)
                finished.set()
                return b'0', None

            return SimpleNamespace(pid=123, returncode=0, communicate=communicate)

        with tempfile.TemporaryDirectory() as scratch, patch('run.CLEANUP_SECONDS', 0.1), \
                patch.dict(os.environ, {'TSC_TIMEOUT': '0.25'}), patch('run.Thread', track_worker), \
                patch('run.subprocess.Popen', side_effect=launch):
            folder = Path(scratch)
            started = time.monotonic()
            try:
                with self.assertRaisesRegex(RuntimeError, 'reference compiler supervisor .* exceeded'):
                    compare_case('candidate', ['reference'], folder, {'input.ts': ''},
                                 {'noEmit': True, 'incremental': True})
                self.assertLess(time.monotonic() - started, 1.5)
                self.assertTrue(entered.is_set())
                receipt = folder / 'reference/actual.owner.json'
                self.assertEqual(compiler_process.read_owner_receipt(receipt)[1], 'expired')
                self.assertFalse(finished.is_set())
                self.assertTrue((folder / 'project/input.ts').is_file())
                self.assertFalse((folder / 'reference/project').exists())
                self.assertFalse((folder / 'candidate').exists())
                for fd in captures:
                    os.fstat(fd)
            finally:
                release.set()
                for worker in workers:
                    worker.join(2)
                    self.assertFalse(worker.is_alive())
            self.assertTrue(finished.is_set())
            for fd in captures:
                with self.assertRaises(OSError):
                    os.fstat(fd)


if sys.platform == 'linux':
    class CompilerSupervisor(unittest.TestCase):
        def invoke(self, folder, source, timeout):
            env = dict(os.environ, STAGE3_VERDICT_CASE='fixture', STAGE3_VERDICT_SIDE='reference')
            with (folder / 'stdout').open('wb') as stdout, (folder / 'stderr').open('wb') as stderr:
                return run_compiler([sys.executable, '-c', dedent(source)], cwd=folder, stdout=stdout,
                                    stderr=stderr, timeout=timeout, env=env)

        def test_reaps_detached_child_after_compiler_exit(self):
            requests = []
            own_pid = None
            kill = os.kill

            def record_kill(pid, signum):
                requests.append(('kill', pid, signum))
                if pid == own_pid:
                    return kill(pid, signum)

            def record_group(group, signum):
                requests.append(('killpg', group, signum))

            def record_pidfd(fd, signum, *args, **kwargs):
                requests.append(('pidfd_send_signal', fd, signum))

            self.enterContext(patch('os.kill', record_kill))
            self.enterContext(patch('os.killpg', record_group))
            self.enterContext(patch('signal.pidfd_send_signal', record_pidfd))
            source = """
    import subprocess, sys
    child = subprocess.Popen([sys.executable, '-c', 'import time; time.sleep(4)'],
                             start_new_session=True)
    print(child.pid, flush=True)
    """
            with tempfile.TemporaryDirectory() as scratch, subprocess.Popen(
                    [sys.executable, '-B', '-c', 'import sys; sys.stdin.read()'],
                    stdin=subprocess.PIPE) as child:
                own_pid = child.pid
                self.addCleanup(self.assertEqual, requests, [('kill', own_pid, signal.SIGTERM)])
                folder = Path(scratch)
                result = self.invoke(folder, source, 3)
                self.assertEqual(result.returncode, 0)
                pid = int((folder / 'stdout').read_text())
                self.assertFalse(Path(f'/proc/{pid}').exists())
                child.send_signal(signal.SIGTERM)
                self.assertEqual(child.wait(timeout=5), -signal.SIGTERM)
                print(f'detached child reaped, recorder forwarded SIGTERM to its own child {own_pid}',
                      flush=True)

        def test_paused_supervisor_cannot_complete_late(self):
            source = """
    import os, time
    from pathlib import Path
    os.kill = lambda *args: Path('compiler-raw-signal').write_text(repr(args))
    Path('ready').write_text(f'987654321 {os.getpid()}')
    time.sleep(4)
    """
            popen = subprocess.Popen

            def launch(command, **kwargs):
                fixture = [sys.executable, '-B', __file__, '--paused-owner-fixture', *command[4:]]
                return popen(fixture, **kwargs)

            with tempfile.TemporaryDirectory() as scratch, patch('run.CLEANUP_SECONDS', 0.25), \
                    patch('run.subprocess.Popen', launch):
                folder = Path(scratch)
                started = time.monotonic()
                with self.assertRaisesRegex(RuntimeError, 'supervisor .* exceeded its deadline'):
                    self.invoke(folder, source, 1)
                self.assertLess(time.monotonic() - started, 2)
                stale_pid, child = map(int, (folder / 'ready').read_text().split())
                parent = int((folder / 'supervisor').read_text())
                self.assertNotEqual(stale_pid, parent)
                until = time.monotonic() + 2
                while any(Path(f'/proc/{pid}').exists() for pid in (parent, child)):
                    self.assertLess(time.monotonic(), until, 'fixture subtree was not reaped')
                    time.sleep(0.01)
                self.assertEqual((folder / 'signals').read_text().splitlines(), ['SIGSTOP', 'SIGCONT'])
                self.assertFalse((folder / 'owner-raw-signal').exists(), 'owner used a raw PID signal')
                self.assertFalse((folder / 'compiler-raw-signal').exists(), 'compiler signalled its parent')
                print('paused supervisor reaped before the bound with no signal to the stale PID', flush=True)


def run_paused_owner_fixture(arguments):
    popen, kill = subprocess.Popen, os.kill

    def record_raw(*args):
        Path('owner-raw-signal').write_text(repr(args))

    def launch(command, **kwargs):
        supervisor = popen(command, **kwargs)
        Path('supervisor').write_text(str(supervisor.pid))
        send_signal = supervisor.send_signal

        def send(signum):
            with Path('signals').open('a') as log:
                print(signum.name, file=log)
            with patch('os.kill', kill):
                send_signal(signum)

        supervisor.send_signal = send
        until = time.monotonic() + 0.7
        while not Path('ready').exists() and time.monotonic() < until:
            time.sleep(0.01)
        supervisor.send_signal(signal.SIGSTOP)
        return supervisor

    with patch('compiler_process.subprocess.Popen', launch), patch('os.kill', record_raw):
        return compiler_process.own_supervisor(*arguments[:4], arguments[4:])


class AdoptedChildren(unittest.TestCase):
    def test_reaper_signals_only_proven_children_through_pidfds(self):
        with tempfile.TemporaryDirectory() as scratch:
            root = Path(scratch)
            for pid, parent in ((101, 33), (102, 33), (103, 44), (104, 33)):
                entry = root / str(pid)
                entry.mkdir()
                fields = ['S', str(parent), *(['0'] * 17), '12345']
                (entry / 'stat').write_text(f'{pid} (fixture ) comm) ' + ' '.join(fields))
            def proc_path(path):
                if path == '/proc':
                    return root
                fd = int(path.rsplit('/', 1)[-1])
                pid = fd if fd != 104 else 105
                return SimpleNamespace(read_text=lambda: f'Pid:\t{pid}\n')

            watcher = SimpleNamespace()
            with patch.object(compiler_process, 'Path', side_effect=proc_path), \
                    patch.object(os, 'readlink', return_value='33'), \
                    patch.object(os, 'getpid', return_value=33), \
                    patch.object(os, 'waitpid', side_effect=[(0, 0), ChildProcessError()]), \
                    patch.object(os, 'pidfd_open', side_effect=lambda pid: pid), \
                    patch.object(os, 'close') as close, patch.object(os, 'kill') as raw, \
                    patch.object(compiler_process, 'poll', return_value=watcher), \
                    patch.object(signal, 'pidfd_send_signal') as send:
                # Keep the mock order independent of directory iteration order.
                def register(fd, _):
                    watcher.poll = lambda timeout: [(fd, compiler_process.POLLIN)] if fd == 102 else []
                watcher.register = register
                compiler_process.reap_descendants(None)
                raw.assert_not_called()
                send.assert_called_once_with(101, signal.SIGKILL)
                self.assertCountEqual([call.args[0] for call in close.call_args_list], [101, 102, 103, 104])
            print('one proven child signalled through a pidfd, zero raw signals', flush=True)


class OwnershipRuling(unittest.TestCase):
    def test_thread_reuse_requires_proven_tgid(self):
        ready, done = Event(), Event()
        thread_ids = []

        def wait():
            thread_ids.append(get_native_id())
            ready.set()
            done.wait()

        thread = Thread(target=wait)
        thread.start()
        try:
            self.assertTrue(ready.wait(5))
            pid = thread_ids[0]
            self.assertNotEqual(pid, os.getpid())
            with self.assertRaises(OSError) as failed_open:
                with os.fdopen(os.pidfd_open(pid), 'rb', buffering=0):
                    pass
            code = failed_open.exception.errno
            self.assertIn(code, (errno.EINVAL, errno.ENOENT))

            for case, expected in (
                    ('real status', None),
                    ('unreadable status', 'cannot read status'),
                    ('foreign reader view', 'procfs PID view does not match'),
                    ('leader Tgid', 'Tgid equals PID'),
                    ('missing Tgid', 'no valid Tgid')):
                with self.subTest(case=case), tempfile.TemporaryDirectory() as scratch:
                    receipt = Path(scratch) / 'actual.owner.json'
                    owner = {'pid': pid, 'start_time': 0, 'state': 'owned',
                             'acknowledged_at': time.time_ns(),
                             'boot_id': compiler_process.read_boot_id(),
                             'pid_namespace': compiler_process.read_pid_namespace()}
                    receipt.write_text(json.dumps(owner))
                    before = receipt.read_bytes()
                    read_text = Path.read_text

                    def read(path, *args, **kwargs):
                        if case == 'foreign reader view' and path.parent == Path('/proc/self/fdinfo'):
                            return f'Pid:\t{pid}\n'
                        if path == Path(f'/proc/{pid}/status'):
                            if case == 'unreadable status':
                                raise PermissionError('controlled unreadable status')
                            if case == 'leader Tgid':
                                return f'Tgid:\t{pid}\n'
                            if case == 'missing Tgid':
                                return 'Name:\tfixture\n'
                        return read_text(path, *args, **kwargs)

                    diagnostic, output = StringIO(), StringIO()
                    seam = nullcontext() if expected is None else patch.object(Path, 'read_text', read)
                    if expected is None:
                        initial = Path(scratch) / 'before.owner.json'
                        initial.write_bytes(before)
                    with seam, redirect_stderr(diagnostic), redirect_stdout(output):
                        status = release_quarantine(receipt)
                    self.assertTrue(thread.is_alive())

                    if expected is None:
                        self.assertEqual(status, 0, diagnostic.getvalue())
                        self.assertEqual(compiler_process.read_owner_receipt(initial)[1], 'dead')
                        updated, proof = compiler_process.read_owner_receipt(receipt)
                        self.assertEqual((proof, updated['released_by']), ('released', 'hand'))
                        detail = 'released by hand while thread was still alive'
                    else:
                        self.assertEqual(status, 2)
                        self.assertEqual(receipt.read_bytes(), before)
                        self.assertIn(expected, diagnostic.getvalue())
                        detail = diagnostic.getvalue().strip()
                    print(f'{case}: pidfd_open {errno.errorcode[code]}, release exit {status}, {detail}',
                          flush=True)
        finally:
            done.set()
            thread.join(5)
            self.assertFalse(thread.is_alive())

    def test_enoent_with_leader_status_refuses(self):
        ready, done = Event(), Event()
        thread_ids = []

        def wait():
            thread_ids.append(get_native_id())
            ready.set()
            done.wait()

        thread = Thread(target=wait)
        thread.start()
        try:
            self.assertTrue(ready.wait(5))
            pid = thread_ids[0]
            self.assertNotEqual(pid, os.getpid())
            with self.assertRaises(OSError) as failed_open:
                os.close(os.pidfd_open(pid))
            code = failed_open.exception.errno
            self.assertIn(code, (errno.ENOENT, errno.EINVAL))

            status_path = Path(f'/proc/{pid}/status')
            read_text = Path.read_text
            real_status = read_text(status_path, errors='replace')
            self.assertIn(f'Pid:\t{pid}\n', real_status)
            self.assertIn(f'Tgid:\t{os.getpid()}\n', real_status)
            leader_status = real_status.replace(f'Tgid:\t{os.getpid()}\n', f'Tgid:\t{pid}\n', 1)

            def read(path, *args, **kwargs):
                if path == status_path:
                    return leader_status
                return read_text(path, *args, **kwargs)

            with tempfile.TemporaryDirectory() as scratch:
                receipt = Path(scratch) / 'actual.owner.json'
                receipt.write_text(json.dumps({
                    'pid': pid, 'start_time': 0, 'state': 'owned',
                    'acknowledged_at': time.time_ns(),
                    'boot_id': compiler_process.read_boot_id(),
                    'pid_namespace': compiler_process.read_pid_namespace(),
                }))
                before = receipt.read_bytes()
                diagnostic = StringIO()
                with patch.object(Path, 'read_text', read), redirect_stderr(diagnostic):
                    status = release_quarantine(receipt)
                    self.assertTrue(thread.is_alive())
                    self.assertEqual(status, 2)
                    self.assertEqual(receipt.read_bytes(), before)
                    self.assertIn('Tgid equals PID', diagnostic.getvalue())
                    with self.assertRaisesRegex(ValueError, 'Tgid equals PID'):
                        compiler_process.read_owner_receipt(receipt)
                print(f'leader status: pidfd_open {errno.errorcode[code]}, release exit {status}, '
                      f'{diagnostic.getvalue().strip()}', flush=True)
        finally:
            done.set()
            thread.join(5)
            self.assertFalse(thread.is_alive())

    def test_reused_owner_pid_requires_matching_start_time(self):
        with tempfile.TemporaryDirectory() as scratch, subprocess.Popen(
                [sys.executable, '-B', '-c', 'import sys; sys.stdin.read()'],
                stdin=subprocess.PIPE) as child:
            try:
                receipt = Path(scratch) / 'actual.owner.json'
                project = Path(scratch) / 'project'
                project.mkdir()
                (project / 'input.ts').write_text('kept')
                start_time = compiler_process.process_start_time(child.pid)
                owner = {'pid': child.pid, 'start_time': start_time,
                         'acknowledged_at': time.time_ns(), 'state': 'owned',
                         'boot_id': compiler_process.read_boot_id(),
                         'pid_namespace': compiler_process.read_pid_namespace()}
                command = [sys.executable, '-B', str(Path(__file__).with_name('run.py')),
                           'release', str(receipt)]

                for recorded, proof, exit_code, state in (
                        (owner, 'alive', 2, 'owned'),
                        (dict(owner, start_time=start_time + 1), 'dead', 0, 'released')):
                    receipt.write_text(json.dumps(recorded))
                    self.assertEqual(compiler_process.read_owner_receipt(receipt)[1], proof)
                    result = subprocess.run(command, capture_output=True, text=True, timeout=5)
                    self.assertEqual(result.returncode, exit_code, result)
                    updated = json.loads(receipt.read_text())
                    if state == 'released':
                        self.assertGreaterEqual(updated.pop('released_at'), owner['acknowledged_at'])
                        self.assertEqual(updated.pop('released_by'), 'hand')
                    self.assertEqual(updated, dict(recorded, state=state))
                    self.assertEqual((project / 'input.ts').read_text(), 'kept')
                    self.assertIsNone(child.poll())
                    print(f'{proof}: release exit={result.returncode}, state={state}', flush=True)
            finally:
                child.terminate()

    def test_hand_release_records_uncertain_cleanup(self):
        with tempfile.TemporaryDirectory() as scratch:
            receipt = Path(scratch) / 'actual.owner.json'
            with subprocess.Popen([sys.executable, '-B', '-c', 'import sys; sys.stdin.read()'],
                                  stdin=subprocess.PIPE) as child:
                owner = {'pid': child.pid, 'start_time': compiler_process.process_start_time(child.pid),
                         'acknowledged_at': time.time_ns(), 'state': 'uncertain',
                         'boot_id': compiler_process.read_boot_id(),
                         'pid_namespace': compiler_process.read_pid_namespace()}
                child.stdin.close()
                child.wait(timeout=5)
            receipt.write_text(json.dumps(owner))
            result = subprocess.run(
                [sys.executable, '-B', str(Path(__file__).with_name('run.py')), 'release', str(receipt)],
                capture_output=True, text=True, timeout=5)
            self.assertEqual(result.returncode, 0, result)
            updated, proof = compiler_process.read_owner_receipt(receipt)
            self.assertEqual(proof, 'released')
            self.assertIn('released_at', updated)
            self.assertGreaterEqual(updated.pop('released_at'), owner['acknowledged_at'])
            self.assertEqual(updated, dict(owner, state='released', released_by='hand',
                                           released_from='uncertain'))
            self.assertIn("The supervisor's subtree cleanup was never verified.", result.stdout)
            print(result.stdout, end='', flush=True)

    def test_late_owner_writes_no_receipt_and_starts_no_supervisor(self):
        with tempfile.TemporaryDirectory() as scratch:
            folder = Path(scratch)
            receipt = folder / 'actual.owner.json'
            attempted = folder / 'supervisor-started'
            deadline = time.monotonic() - 1
            fixture = dedent("""
                import sys
                from pathlib import Path
                sys.path.insert(0, sys.argv[1])
                import compiler_process

                def launch(*args, **kwargs):
                    Path(sys.argv[2]).touch()
                    raise AssertionError('late owner attempted supervisor launch')

                compiler_process.subprocess.Popen = launch
                raise SystemExit(compiler_process.own_supervisor(*sys.argv[3:7], sys.argv[7:]))
            """)
            with (folder / 'capture').open('wb') as capture:
                result = subprocess.run(
                    [sys.executable, '-B', '-c', fixture, str(Path(__file__).resolve().parent),
                     str(attempted), str(receipt), str(deadline), str(deadline + 5),
                     str(capture.fileno()), 'fixture'],
                    pass_fds=(capture.fileno(),), capture_output=True, text=True, timeout=5)
            self.assertFalse(receipt.exists(), 'late owner wrote a receipt')
            self.assertFalse(attempted.exists())
            self.assertEqual(result.returncode, 3, result)
            self.assertIn('deadline', result.stderr)
            print(f'exit={result.returncode}, no receipt or supervisor: {result.stderr.strip()}',
                  flush=True)


class AdmissionFence(unittest.TestCase):
    def test_owner_reports_pidfd_open_failure_cause(self):
        causes = ((errno.ENOSYS, 'the host has no pidfds'),
                  (errno.EMFILE, 'the process reached its descriptor limit'))
        for (code, expected), (_, other) in (causes, causes[::-1]):
            with self.subTest(errno=errno.errorcode[code]), tempfile.TemporaryDirectory() as scratch:
                receipt = Path(scratch) / 'actual.owner.json'
                expiry = time.monotonic() + 5
                diagnostic = StringIO()
                with patch('compiler_process.os.pidfd_open', side_effect=OSError(code, os.strerror(code))), \
                        patch('compiler_process.subprocess.Popen') as launch, \
                        patch('compiler_process.os.close'), patch('compiler_process.os.write'), \
                        redirect_stderr(diagnostic):
                    status = compiler_process.own_supervisor(
                        str(receipt), str(expiry), str(expiry + 5), '97', ['fixture'])
                self.assertEqual((status, receipt.exists(), launch.call_count), (1, False, 0))
                self.assertIn(expected, diagnostic.getvalue())
                self.assertNotIn(other, diagnostic.getvalue())
                print(f'owner exit {status}, no receipt, no launch: {diagnostic.getvalue().strip()}',
                      flush=True)

    def test_owner_refuses_foreign_procfs_view(self):
        with tempfile.TemporaryDirectory() as scratch:
            receipt = Path(scratch) / 'actual.owner.json'
            read_text = Path.read_text
            fdinfo_reads = []
            foreign_pid = os.getpid() + 1

            def read(path, *args, **kwargs):
                if path.parent == Path('/proc/self/fdinfo'):
                    fdinfo_reads.append(path)
                    return f'Pid:\t{foreign_pid}\n'
                return read_text(path, *args, **kwargs)

            expiry = time.monotonic() + 5
            diagnostics = StringIO()
            supervisor = SimpleNamespace(returncode=0, communicate=lambda **_: (b'0', None))
            with patch.object(Path, 'read_text', read), \
                    patch('compiler_process.subprocess.Popen', return_value=supervisor) as launch, \
                    patch('compiler_process.os.close'), patch('compiler_process.os.write'), \
                    redirect_stderr(diagnostics):
                status = compiler_process.own_supervisor(
                    str(receipt), str(expiry), str(expiry + 5), '97', ['fixture'])

            self.assertEqual((status, receipt.exists(), launch.call_count), (1, False, 0))
            self.assertEqual(len(fdinfo_reads), 1)
            self.assertIn('procfs PID view does not match', diagnostics.getvalue())
            print(f'owner exit {status}, no receipt, no supervisor launch: {diagnostics.getvalue().strip()}',
                  flush=True)

    def test_deadline_claim_blocks_delayed_owner_publication(self):
        with tempfile.TemporaryDirectory() as scratch:
            receipt = Path(scratch) / 'actual.owner.json'
            clock = [10]
            link, replace = os.link, Path.replace

            def claim_at_deadline():
                if clock[0] == 10:
                    clock[0] = 12
                    self.assertTrue(claim_expired_receipt(receipt))

            def delayed_link(source, target):
                claim_at_deadline()
                return link(source, target)

            def delayed_replace(source, target):
                claim_at_deadline()
                return replace(source, target)

            supervisor = SimpleNamespace(returncode=0, communicate=lambda **_: (b'0', None))
            diagnostic = StringIO()
            with patch('compiler_process.time.monotonic', side_effect=lambda: clock[0]), \
                    patch('compiler_process.os.link', delayed_link), \
                    patch.object(Path, 'replace', delayed_replace), \
                    patch('compiler_process.subprocess.Popen', return_value=supervisor) as launch, \
                    patch('compiler_process.os.close'), patch('compiler_process.os.write'), \
                    redirect_stderr(diagnostic):
                status = compiler_process.own_supervisor(str(receipt), '11', '16', '97', ['fixture'])

            owner, proof = compiler_process.read_owner_receipt(receipt)
            self.assertEqual((owner['state'], proof), ('expired', 'expired'))
            self.assertEqual(status, 3)
            launch.assert_not_called()
            self.assertIn('deadline', diagnostic.getvalue())
            self.assertEqual(list(Path(scratch).iterdir()), [receipt])
            diagnostic = StringIO()
            with redirect_stderr(diagnostic):
                self.assertEqual(release_quarantine(receipt), 2)
            self.assertIn('Nothing to release', diagnostic.getvalue())
            self.assertEqual(compiler_process.read_owner_receipt(receipt), (owner, proof))
            print(f'delayed publication: exit {status}, expired receipt, zero supervisor launches',
                  flush=True)

    def test_owner_link_wins_before_deadline_claim(self):
        with tempfile.TemporaryDirectory() as scratch:
            receipt = Path(scratch) / 'actual.owner.json'
            clock = [10]
            acknowledged = []

            def launch(*args, **kwargs):
                clock[0] = 12
                acknowledged.append(receipt.read_bytes())
                self.assertFalse(claim_expired_receipt(receipt))
                self.assertEqual(receipt.read_bytes(), acknowledged[0])
                self.assertEqual(compiler_process.read_owner_receipt(receipt)[1], 'alive')
                return SimpleNamespace(returncode=0, communicate=lambda **_: (b'0', None))

            with patch('compiler_process.time.monotonic', side_effect=lambda: clock[0]), \
                    patch('compiler_process.subprocess.Popen', side_effect=launch) as spawn, \
                    patch('compiler_process.os.close'), patch('compiler_process.os.write'):
                status = compiler_process.own_supervisor(str(receipt), '11', '16', '97', ['fixture'])
            self.assertEqual(status, 0)
            spawn.assert_called_once()
            self.assertEqual(json.loads(acknowledged[0])['state'], 'owned')
            self.assertEqual(compiler_process.read_owner_receipt(receipt)[1], 'released')
            print('owner linked first: deadline claim lost, owned receipt survived until release', flush=True)


class ReceiptScope(unittest.TestCase):
    def test_reader_and_hand_command_require_receipt_scope(self):
        boot = '4b4571c9-b321-4901-aec3-e03648d605f4'
        old_boot = '74a385f2-c85a-4ec4-8c7d-cbd86b908bf9'
        for name, reader_boot, reader_ns, start, expected in (
                ('namespace mismatch', boot, 11, FileNotFoundError(), 'namespace differs'),
                ('namespace unreadable', boot, FileNotFoundError(), FileNotFoundError(), 'cannot read PID namespace'),
                ('earlier boot', old_boot, 10, 33, 'dead'),
                ('boot unreadable', FileNotFoundError(), 10, 33, 'cannot read boot ID'),
                ('same scope', boot, 10, 33, 'alive'),
                ('foreign procfs view', boot, 10, 34, 'procfs PID view does not match'),
                ('filtered procfs', boot, 10, FileNotFoundError(), 'cannot be read')):
            with self.subTest(name=name), tempfile.TemporaryDirectory() as scratch:
                receipt = Path(scratch) / 'actual.owner.json'
                recorded = {'pid': 123, 'start_time': 33, 'state': 'owned',
                            'boot_id': boot, 'pid_namespace': 10, 'acknowledged_at': time.time_ns()}
                receipt.write_text(json.dumps(recorded))
                read_text, stat = Path.read_text, os.stat

                def read(path, *args, **kwargs):
                    if path == Path('/proc/self/fdinfo/97'):
                        return 'Pid:\t456\n' if name == 'foreign procfs view' else 'Pid:\t123\n'
                    if path != Path('/proc/sys/kernel/random/boot_id'):
                        return read_text(path, *args, **kwargs)
                    if isinstance(reader_boot, OSError):
                        raise reader_boot
                    return reader_boot + '\n'

                def namespace(path, *args, **kwargs):
                    if path != '/proc/self/ns/pid':
                        return stat(path, *args, **kwargs)
                    if isinstance(reader_ns, OSError):
                        raise reader_ns
                    return SimpleNamespace(st_ino=reader_ns)

                start_patch = {'side_effect': start} if isinstance(start, OSError) else {'return_value': start}
                pidfd_patch = ({'return_value': 97} if name in
                               ('same scope', 'foreign procfs view', 'filtered procfs')
                               else {'side_effect': ProcessLookupError})
                diagnostic, output = StringIO(), StringIO()
                with patch.object(Path, 'read_text', read), patch('os.stat', namespace), \
                        patch('compiler_process.process_start_time', **start_patch) as lookup, \
                        patch('compiler_process.os.pidfd_open', **pidfd_patch), \
                        patch('compiler_process.os.fdopen', side_effect=lambda *_, **__: StringIO()), \
                        patch('compiler_process.os.close'), \
                        patch('compiler_process.poll', return_value=SimpleNamespace(
                            register=lambda *_: None, poll=lambda _: [])), \
                        patch.object(sys, 'argv', ['run.py', 'release', str(receipt)]), \
                        redirect_stderr(diagnostic), redirect_stdout(output):
                    status = main()
                    updated = json.loads(receipt.read_text())
                    if name == 'foreign procfs view':
                        self.assertEqual((status, updated), (2, recorded))
                    if expected in ('alive', 'dead'):
                        if expected == 'dead':
                            receipt.write_text(json.dumps(recorded))
                        self.assertEqual(compiler_process.read_owner_receipt(receipt)[1], expected)
                    else:
                        with self.assertRaisesRegex(ValueError, expected):
                            compiler_process.read_owner_receipt(receipt)
                    if name not in ('same scope', 'filtered procfs'):
                        lookup.assert_not_called()

                if expected == 'dead':
                    self.assertEqual(status, 0, diagnostic.getvalue())
                    self.assertEqual(updated['state'], 'released')
                    self.assertEqual(updated['released_by'], 'hand')
                    self.assertIn('Release recorded', output.getvalue())
                else:
                    self.assertEqual(status, 2)
                    self.assertEqual(updated, recorded)
                    self.assertIn('alive by identity' if expected == 'alive' else expected,
                                  diagnostic.getvalue())
                print(f'{name}: release exit {status}, receipt {updated["state"]}', flush=True)

    def test_legacy_receipt_without_scope_is_unverifiable(self):
        with tempfile.TemporaryDirectory() as scratch:
            receipt = Path(scratch) / 'actual.owner.json'
            original = {'pid': 123, 'start_time': 33, 'state': 'owned', 'acknowledged_at': time.time_ns()}
            for fields in ({}, {'pid_namespace': 10}, {'boot_id': 'boot'}):
                with self.subTest(fields=fields):
                    recorded = dict(original, **fields)
                    receipt.write_text(json.dumps(recorded))
                    with self.assertRaisesRegex(ValueError, 'lacks a recognized owner identity'):
                        compiler_process.read_owner_receipt(receipt)
                    diagnostic = StringIO()
                    with redirect_stderr(diagnostic):
                        self.assertEqual(release_quarantine(receipt), 2)
                    self.assertIn('lacks a recognized owner identity', diagnostic.getvalue())
                    self.assertEqual(json.loads(receipt.read_text()), recorded)


class CleanupReceipt(unittest.TestCase):
    def test_status_read_fault_waits_before_recording_uncertainty(self):
        with tempfile.TemporaryDirectory() as scratch:
            receipt = Path(scratch) / 'owner.json'
            waiting = Event()
            release = Event()
            outcomes = []

            def wait():
                waiting.set()
                if not release.wait(3):
                    raise AssertionError('test did not release supervisor')
                return 0

            def communicate(**kwargs):
                raise OSError('controlled status read failure')

            supervisor = SimpleNamespace(communicate=communicate, wait=wait,
                                         send_signal=lambda _: None, stdout=StringIO())
            expiry = time.monotonic() + 10
            diagnostics = StringIO()

            def own():
                try:
                    outcomes.append(compiler_process.own_supervisor(
                        str(receipt), str(expiry), str(expiry + 5), '97', ['fixture']))
                except Exception as error:
                    outcomes.append(str(error))

            with patch('compiler_process.subprocess.Popen', return_value=supervisor), \
                    patch('compiler_process.os.close') as close, patch('compiler_process.os.write'), \
                    redirect_stderr(diagnostics):
                owner = Thread(target=own)
                owner.start()
                try:
                    self.assertTrue(waiting.wait(0.5),
                                    f"status fault returned without reaping: {receipt.read_text()}")
                    self.assertEqual(json.loads(receipt.read_text())['state'], 'owned')
                    close.assert_not_called()
                finally:
                    release.set()
                    owner.join(4)
                self.assertFalse(owner.is_alive())
            self.assertEqual(outcomes, [1])
            self.assertEqual(json.loads(receipt.read_text())['state'], 'uncertain')
            self.assertIn('controlled status read failure', diagnostics.getvalue())

    def test_release_publication_observes_closed_capture_identities(self):
        with tempfile.TemporaryDirectory() as scratch:
            folder = Path(scratch)
            receipt = folder / 'owner.json'
            (folder / 'release-wait').touch()
            expiry = time.monotonic() + 5
            with (folder / 'capture-out').open('wb') as out, (folder / 'capture-err').open('wb') as err:
                owner = subprocess.run(
                    [sys.executable, '-B', __file__, '--owner-fixture', scratch, str(receipt),
                     str(expiry), str(expiry + 5), str(out.fileno()), 'fixture'],
                    pass_fds=(out.fileno(),), stderr=err, stdout=subprocess.PIPE,
                    timeout=5)
            self.assertEqual(owner.returncode, 1)
            observed = json.loads((folder / 'release-observed.json').read_text())
            self.assertEqual(observed, {'state': 'released', 'captures_open': [False, False]})
            self.assertEqual(json.loads(receipt.read_text())['state'], 'released')


def observe_owner_release(output, receipt, output_fd):
    replace = Path.replace
    descriptors = (int(output_fd), 2)
    identities = [(stat.st_dev, stat.st_ino) for stat in map(os.fstat, descriptors)]

    def capture_open(fd, identity):
        try:
            stat = os.fstat(fd)
        except OSError:
            return False
        return (stat.st_dev, stat.st_ino) == identity

    def observe(source, target):
        if target == Path(receipt):
            state = json.loads(source.read_text())['state']
            if state == 'released':
                captures = [capture_open(fd, identity) for fd, identity in zip(descriptors, identities)]
                (output / 'release-observed.json').write_text(json.dumps(
                    {'state': state, 'captures_open': captures}))
        return replace(source, target)

    return patch.object(Path, 'replace', observe)


class HeldSupervisor:
    pid = 123
    stdout = SimpleNamespace(close=lambda: None)

    def __init__(self, output):
        self.output = output

    def communicate(self, *, timeout):
        raise subprocess.TimeoutExpired('fixture', timeout)

    def send_signal(self, signum):
        assert signum == signal.SIGCONT

    def wait(self):
        (self.output / 'wait-ready').touch()
        while not (self.output / 'release-wait').exists():
            time.sleep(0.01)
        (self.output / 'wait-completed').touch()
        return 0


def run_shutdown_fixture(output, ownership_control=None):
    """Run the real verdict CLI with a supervisor whose wait is withheld."""
    if ownership_control:
        import run
        source = Path(ownership_control).read_text()
        exec(compile(source, run.__file__, 'exec'), vars(run))
    case = output / 'acceptance/case'
    popen = subprocess.Popen

    def launch(command, **kwargs):
        if command[3] == '--owner':
            fixture = [sys.executable, '-B', __file__, '--owner-fixture', str(output), *command[4:]]
            return popen(fixture, **kwargs)
        return HeldSupervisor(output)

    def suite(binary, reference, folder, tiny):
        if not tiny:
            return compare_case(binary, reference, case, {'input.ts': ''},
                                {'noEmit': True, 'incremental': True})
        return dict.fromkeys(('total', 'passed', 'failed', 'excluded', 'deferred',
                              'emit_enabled', 'no_emit'), 0) | {'timeouts': [], 'failures': []}

    argv = ['run.py', '--tsc', sys.executable, '--node-tsc', __file__, str(output)]
    with patch.object(sys, 'argv', argv), patch('run.CLEANUP_SECONDS', 0.1), \
            patch.dict(os.environ, {'TSC_TIMEOUT': '0.5'}), patch('run.subprocess.Popen', launch), \
            patch('run.driver_suite', side_effect=suite), patch('run.upstream_tree', return_value=None), \
            patch('run.baseline_suite', side_effect=lambda *args: suite(None, None, None, True)):
        return main()


class ShutdownOwnership(unittest.TestCase):
    def test_cli_exits_with_cleanup_acknowledged_and_still_pending(self):
        with tempfile.TemporaryDirectory() as scratch:
            output = Path(scratch) / 'verdict'
            case = output / 'acceptance/case'
            receipt = case / 'reference/actual.owner.json'
            argv = [sys.executable, '-B', str(Path(__file__).resolve()), '--shutdown-fixture', str(output)]
            control = os.environ.get('VERDICT_SHUTDOWN_CONTROL')
            if control:
                argv.append(control)

            with subprocess.Popen(argv, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True) as cli:
                try:
                    deadline = time.monotonic() + 4
                    while not all((output / name).exists() for name in ('wait-ready', 'cli-returned')):
                        self.assertLess(time.monotonic(), deadline, 'CLI did not report its bounded case')
                        time.sleep(0.01)
                    summary = json.loads((output / 'summary.json').read_text())
                    self.assertFalse(summary['success'])
                    self.assertEqual(summary['suites']['acceptance']['status'], 'harness_error')
                    error = summary['suites']['acceptance']['error']
                    self.assertIn('exceeded its deadline', error)
                    self.assertTrue((case / 'project/input.ts').is_file())
                    self.assertFalse((case / 'reference/project').exists())
                    self.assertFalse((case / 'candidate').exists())
                    try:
                        stdout, stderr = cli.communicate(timeout=1)
                    except subprocess.TimeoutExpired:
                        self.fail('CLI shutdown exceeded one second with the supervisor wait withheld')
                    self.assertEqual(cli.returncode, 2, (stdout, stderr))
                    self.assertFalse((output / 'wait-completed').exists())
                    self.assertTrue(receipt.is_file(), 'CLI exited without a cleanup owner receipt')
                    owner = json.loads(receipt.read_text())
                    self.assertEqual(owner['state'], 'owned')
                    self.assertIn(f'acknowledged at {owner["acknowledged_at"]} '
                                  f'by pid {owner["pid"]}; liveness not proven', error)
                    self.assertEqual(compiler_process.read_owner_receipt(receipt)[1], 'alive')
                    print(error, flush=True)
                    print('exit=2, wait still pending, owner acknowledged, scratch quarantined', flush=True)
                finally:
                    if output.is_dir():
                        (output / 'release-wait').touch()
                    if cli.poll() is None:
                        cli.communicate(timeout=4)
                    if receipt.exists():
                        deadline = time.monotonic() + 4
                        while json.loads(receipt.read_text())['state'] != 'released':
                            self.assertLess(time.monotonic(), deadline, 'owner did not finish cleanup')
                            time.sleep(0.01)
            self.assertTrue((output / 'wait-completed').exists())
            self.assertTrue((case / 'project/input.ts').is_file())
            observation = json.loads((output / 'release-observed.json').read_text())
            self.assertEqual(observation, {'state': 'released', 'captures_open': [False, False]})
            print('owner finished the wait and both captures were closed at release', flush=True)


class UpstreamSource(unittest.TestCase):
    """The baselines read a provided checkout at the pin, never one cloned from a mirror or the network."""

    def head(self, stdout):
        return patch('run.subprocess.run', return_value=subprocess.CompletedProcess([], 0, stdout=stdout))

    def test_neither_source_refuses_and_runs_nothing(self):
        # Mutant: the clone restored, which runs git against the network instead of refusing.
        environment = {key: value for key, value in os.environ.items()
                       if key not in ('STAGE3_VERDICT_UPSTREAM', 'ADAMIC_TYPESCRIPT_SOURCE')}
        with patch.dict(os.environ, environment, clear=True), self.head(PIN + '\n') as ran:
            with self.assertRaisesRegex(RuntimeError, 'ADAMIC_TYPESCRIPT_SOURCE.*never clones'):
                upstream_tree()
            ran.assert_not_called()

    def test_the_gate_inputs_checkout_is_read_and_the_verdicts_own_wins(self):
        with tempfile.TemporaryDirectory() as gate, tempfile.TemporaryDirectory() as own:
            with patch.dict(os.environ, {'ADAMIC_TYPESCRIPT_SOURCE': gate}), self.head(PIN + '\n') as ran:
                os.environ.pop('STAGE3_VERDICT_UPSTREAM', None)
                self.assertEqual(upstream_tree(), Path(gate).resolve())
                self.assertEqual(ran.call_args.args[0], ['git', '-C', str(Path(gate).resolve()), 'rev-parse', 'HEAD'])
            with patch.dict(os.environ, {'ADAMIC_TYPESCRIPT_SOURCE': gate, 'STAGE3_VERDICT_UPSTREAM': own}), \
                    self.head(PIN + '\n'):
                self.assertEqual(upstream_tree(), Path(own).resolve())

    def test_a_checkout_off_the_pin_refuses(self):
        with tempfile.TemporaryDirectory() as gate, patch.dict(os.environ, {'ADAMIC_TYPESCRIPT_SOURCE': gate}):
            os.environ.pop('STAGE3_VERDICT_UPSTREAM', None)
            with self.head('0' * 40 + '\n'), self.assertRaisesRegex(RuntimeError, 'not the pinned ' + PIN):
                upstream_tree()
            # Not a checkout at all: git answers nothing, and the refusal says so. The ceiling keeps git from finding
            # a checkout the temporary directory happens to sit inside.
            with patch.dict(os.environ, {'GIT_CEILING_DIRECTORIES': str(Path(gate).resolve().parent)}), \
                    self.assertRaisesRegex(RuntimeError, 'is at no commit'):
                upstream_tree()


if __name__ == '__main__':
    if sys.argv[1:2] == ['--paused-owner-fixture']:
        raise SystemExit(run_paused_owner_fixture(sys.argv[2:]))
    if len(sys.argv) > 2 and sys.argv[1] == '--owner-fixture':
        output = Path(sys.argv[2])
        with patch('compiler_process.subprocess.Popen', return_value=HeldSupervisor(output)), \
                observe_owner_release(output, sys.argv[3], sys.argv[6]):
            raise SystemExit(compiler_process.own_supervisor(*sys.argv[3:7], sys.argv[7:]))
    if len(sys.argv) > 2 and sys.argv[1] == '--shutdown-fixture':
        output = Path(sys.argv[2])
        code = run_shutdown_fixture(output, *sys.argv[3:])
        (output / 'cli-returned').touch()
        raise SystemExit(code)
    result = unittest.main(exit=False).result
    if result.skipped:
        print(f'Verdict tests must run without skips: {result.skipped}', file=sys.stderr)
    raise SystemExit(not result.wasSuccessful() or bool(result.skipped))
