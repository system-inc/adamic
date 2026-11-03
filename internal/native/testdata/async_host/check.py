#!/usr/bin/env python3
"""Host bridge controls and mutants. Each command has file logs and a deadline."""
import argparse
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

# A mutant's diagnostic is a list of shapes, each a list of byte strings that must all appear in its
# stderr. A racing mutant can trip a different report depending on the schedule: the queue-lock mutant
# is usually caught as the named data race, and sometimes as a heap-use-after-free in host_find, which
# its unlocked queue also causes (Loom's Hot, Nov 2, #7kfk2rn). Only those known shapes count. Any
# other report means the check passed for the wrong reason, so it doesn't.
RACE = [[b'ThreadSanitizer: data race']]
QUEUE_LOCK = [[b'ThreadSanitizer: data race'], [b'ThreadSanitizer: heap-use-after-free', b'host_find']]


def caught(stderr, shapes):
    return any(all(part in stderr for part in shape) for shape in shapes)


def check_caught():
    # The rule above, checked on fixed reports before any mutant runs, so loosening it fails here.
    race = b'WARNING: ThreadSanitizer: data race (pid=1)\n  Write of size 8 at 0x1 by thread T1:\n'
    freed = b'WARNING: ThreadSanitizer: heap-use-after-free (pid=1)\n    #0 host_find host_promises.c:76\n'
    elsewhere = b'WARNING: ThreadSanitizer: heap-use-after-free (pid=1)\n    #0 adamic_release heap.c:40\n'
    unrelated = b'WARNING: ThreadSanitizer: signal-unsafe call inside of a signal (pid=1)\n'
    assert caught(race, QUEUE_LOCK) and caught(freed, QUEUE_LOCK), 'a known queue-lock shape was refused'
    assert not caught(elsewhere, QUEUE_LOCK), 'a use after free outside host_find was taken'
    assert not caught(unrelated, QUEUE_LOCK), 'an unrelated report was taken'
    assert caught(race, RACE) and not caught(freed, RACE), 'the direct mutant takes only its race'


def main():
    check_caught()
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

    parser = argparse.ArgumentParser()
    parser.add_argument('--sanitizer', choices=['address,undefined', 'thread'])
    args = parser.parse_args()
    for sanitizer in ([args.sanitizer] if args.sanitizer else ['address,undefined', 'thread']):
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

        mutants = [('direct', 'ADAMIC_HOST_DIRECT_SETTLE_MUTANT', 'direct', RACE),
                   ('queue-lock', 'ADAMIC_HOST_QUEUE_LOCK_MUTANT', 'stress', QUEUE_LOCK)] \
                  if label == 'tsan' else [
                      ('checkpoint', 'ADAMIC_HOST_CHECKPOINT_MUTANT', 'checkpoint', [[b'Assertion']]),
                      ('wake', 'ADAMIC_HOST_WAKE_MUTANT', 'resolve', None),
                      ('host-registry', 'ADAMIC_HOST_EXIT_REGISTRY_MUTANT', 'exit', [[b'LeakSanitizer: detected memory leaks']]),
                      ('settle-cycle', 'ADAMIC_ASYNC_SETTLE_CYCLE_MUTANT', 'resolve', [[b'LeakSanitizer: detected memory leaks']]),
                      ('cancel-cycle', 'ADAMIC_ASYNC_CANCEL_CYCLE_MUTANT', 'cancel-abandon', [[b'LeakSanitizer: detected memory leaks']]),
                      ('exit-cycle', 'ADAMIC_ASYNC_EXIT_CYCLE_MUTANT', 'exit', [[b'LeakSanitizer: detected memory leaks']])]
        for name, macro, mode, diagnostic in mutants:
            binary = build(name, macro)
            for attempt in range(3 if label == 'tsan' else 1):
                result = run([binary, mode], label + '-' + name + '-mutant-' + str(attempt), environment, timeout=3 if diagnostic is None else 30)
                if diagnostic is None:
                    assert result[0] == 'timeout', (name, result)
                    print('mutant', name, 'caught by bounded timeout (3s)', flush=True)
                else:
                    assert result[0] not in [0, 'timeout'] and caught(result[2], diagnostic), (name, result)
                    shape = next(shape for shape in diagnostic if all(part in result[2] for part in shape))
                    print('mutant', name, 'attempt', attempt + 1, 'caught by', b' + '.join(shape).decode(), flush=True)
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
