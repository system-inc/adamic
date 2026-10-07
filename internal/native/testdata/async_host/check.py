#!/usr/bin/env python3
"""Host bridge controls and mutants. Each command has file logs and a deadline."""
import os
from pathlib import Path
import subprocess
import tempfile

HERE = Path(__file__).resolve().parent
RUNTIME = HERE.parents[1] / 'runtime'
FLAGS = ['-std=c11', '-Wall', '-Wextra', '-Werror', '-pedantic', '-O1', '-g',
         '-pthread', '-fno-omit-frame-pointer']
MODES = ['resolve', 'reject', 'early', 'double', 'abandon', 'cancel', 'cancel-abandon', 'exit',
         'invalid', 'hooks', 'app', 'direct', 'stress', 'checkpoint', 'stale']
REFUSED = b'adamic: host completion refused: request retired or already completed\n'


def main():
    directory = Path(tempfile.mkdtemp(prefix='host-promises-'))
    print('artifacts:', directory, flush=True)

    def run(command, name, env=None, timeout=30):
        with (directory / (name + '.out')).open('wb') as out, \
             (directory / (name + '.err')).open('wb') as err:
            try:
                result = subprocess.run(command, stdout=out, stderr=err, env=env, timeout=timeout)
                status = result.returncode
            except subprocess.TimeoutExpired:
                status = 'timeout'
        return status, (directory / (name + '.out')).read_bytes(), (directory / (name + '.err')).read_bytes()

    def require_build(result):
        assert result[0] == 0, ('build failed', result)

    for sanitizer in ['address,undefined', 'thread']:
        label = 'asan' if sanitizer.startswith('address') else 'tsan'
        flags = FLAGS + ['-fsanitize=' + sanitizer, '-fno-sanitize-recover=all']
        environment = dict(os.environ, ASAN_OPTIONS='detect_leaks=1:halt_on_error=1',
                           UBSAN_OPTIONS='halt_on_error=1', TSAN_OPTIONS='halt_on_error=1:exitcode=66:history_size=4:report_atomic_races=1')
        common = []
        for source in sorted(RUNTIME.glob('*.c')):
            if source.name == 'async.c':
                continue
            target = directory / (label + '-' + source.stem + '.o')
            require_build(run(['clang', *flags, '-I', str(RUNTIME), '-c', str(source), '-o', str(target)],
                              label + '-' + source.stem + '-build'))
            common.append(str(target))

        def build(name, macro=None, replacement=None):
            target = directory / (label + '-' + name + '.o')
            command = ['clang', *flags, '-I', str(RUNTIME)]
            if macro:
                command.append('-D' + macro)
            source = RUNTIME / 'async.c'
            if replacement:
                mutant_dir = directory / (label + '-' + name + '-source')
                mutant_dir.mkdir()
                source = mutant_dir / 'async.c'
                source.write_bytes((RUNTIME / 'async.c').read_bytes())
                implementation = (RUNTIME / 'async_host_impl.h').read_text()
                old, new = replacement
                assert implementation.count(old) == 1, ('mutant seam must be unique', name)
                (mutant_dir / 'async_host_impl.h').write_text(implementation.replace(old, new, 1))
            require_build(run([*command, '-c', str(source), '-o', str(target)],
                              label + '-' + name + '-async-build'))
            binary = directory / (label + '-' + name)
            require_build(run([*command, str(HERE / 'host_promises.c'),
                               str(target), *common, '-lm', '-o', str(binary)], label + '-' + name + '-build'))
            return str(binary)

        control = build('control')
        for mode in MODES:
            result = run([control, mode], label + '-' + mode, environment)
            expected_error = REFUSED * (2 if mode == 'double' else 1 if mode in ['abandon', 'cancel-abandon', 'exit', 'stale'] else 0)
            count = 0 if mode in ['abandon', 'cancel', 'cancel-abandon', 'exit'] else 16 if mode == 'stress' else 2
            reaction = b'rejected\n' if mode in ['reject', 'invalid'] else b'fulfilled\n'
            expected = b'prefix\n' + reaction * count + (mode + ' clean\n').encode()
            assert result == (0, expected, expected_error), (label, mode, result)
            print(label, mode, 'clean', flush=True)

        for mode in ['resolve', 'reject', 'early']:
            witness = directory / 'oracle.mjs'
            witness.write_bytes((HERE / 'oracle.a').read_bytes())
            node = run(['node', str(witness), mode], label + '-node-' + mode)
            result = run([control, mode], label + '-oracle-' + mode, environment)
            assert node == result, (mode, node, result)
            print(label, mode, 'Node trace matches', flush=True)

        mutants = [('direct', 'ADAMIC_HOST_DIRECT_SETTLE_MUTANT', 'direct', b'ThreadSanitizer: data race'),
                   ('queue-lock', 'ADAMIC_HOST_QUEUE_LOCK_MUTANT', 'stress', b'ThreadSanitizer: data race')] \
                  if label == 'tsan' else [
                      ('checkpoint', 'ADAMIC_HOST_CHECKPOINT_MUTANT', 'checkpoint', b'Assertion'),
                      ('wake', 'ADAMIC_HOST_WAKE_MUTANT', 'resolve', None),
                      ('host-registry', 'ADAMIC_HOST_EXIT_REGISTRY_MUTANT', 'exit', b'LeakSanitizer: detected memory leaks'),
                      ('settle-cycle', 'ADAMIC_ASYNC_SETTLE_CYCLE_MUTANT', 'resolve', b'LeakSanitizer: detected memory leaks'),
                      ('cancel-cycle', 'ADAMIC_ASYNC_CANCEL_CYCLE_MUTANT', 'cancel-abandon', b'LeakSanitizer: detected memory leaks'),
                      ('exit-cycle', 'ADAMIC_ASYNC_EXIT_CYCLE_MUTANT', 'exit', b'LeakSanitizer: detected memory leaks')]
        for name, macro, mode, diagnostic in mutants:
            binary = build(name, macro)
            for attempt in range(3 if label == 'tsan' else 1):
                result = run([binary, mode], label + '-' + name + '-mutant-' + str(attempt), environment, timeout=3 if diagnostic is None else 30)
                if diagnostic is None:
                    assert result[0] == 'timeout', (name, result)
                    print('mutant', name, 'caught by bounded timeout (3s)', flush=True)
                else:
                    assert result[0] not in [0, 'timeout'] and diagnostic in result[2], (name, result)
                    print('mutant', name, 'attempt', attempt + 1, 'caught by', diagnostic.decode(), flush=True)
        if label == 'asan':
            for name, mode, old, new in [
                ('payload-copy', 'resolve', 'memcpy(entry->bytes, bytes, length);', "memset(entry->bytes, 'x', length);"),
                ('payload-status', 'resolve', 'entry->status = status;', 'entry->status = status + 1;'),
                ('utf8-validation', 'invalid', 'if (!host_valid_utf8(bytes, length))', 'if (false && !host_valid_utf8(bytes, length))'),
                ('token-reuse', 'stale', 'entry->identity = host_next_identity++;', 'entry->identity = 1;'),
                ('hooks-once', 'hooks', 'if (host_hooks_set || host_started || host_closed)', 'if (host_started || host_closed)'),
            ]:
                binary = build(name, replacement=(old, new))
                result = run([binary, mode], label + '-' + name + '-mutant', environment)
                assert result[0] not in [0, 'timeout'] and b'Assertion' in result[2], (name, result)
                assert b'ERROR: AddressSanitizer' not in result[2], (name, result)
                print('mutant', name, 'caught by contract assertion', flush=True)
    print('PASS', flush=True)


if __name__ == '__main__':
    main()
