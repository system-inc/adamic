#!/usr/bin/env python3
"""Standalone Node differential and sanitizer proofs. All child output goes to files."""
import os
from pathlib import Path
import re
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parent
FLAGS = ['-std=c11', '-Wall', '-Wextra', '-Werror', '-pedantic', '-O1', '-g',
         '-pthread', '-fsanitize=address,undefined', '-fno-sanitize-recover=all']


def run(command, stem, directory, env=None):
    with (directory / (stem + '.out')).open('wb') as stdout, \
         (directory / (stem + '.err')).open('wb') as stderr:
        result = subprocess.run(command, stdout=stdout, stderr=stderr, env=env, timeout=30)
    return result.returncode, (directory / (stem + '.out')).read_bytes(), \
        (directory / (stem + '.err')).read_bytes()


def main():
    directory = Path(tempfile.mkdtemp(prefix='p286ycm-'))
    print('artifacts:', directory, flush=True)
    oracle = directory / 'oracle.cjs'
    oracle.write_bytes((ROOT / 'oracle.a').read_bytes())
    fixture = directory / 'input'
    fixture.write_text('payload\n')
    environment = dict(os.environ, ASAN_OPTIONS='detect_leaks=1:halt_on_error=1',
                       UBSAN_OPTIONS='halt_on_error=1')

    def build(name, macro=None):
        binary = directory / name
        command = ['clang', *FLAGS]
        if macro:
            command.append('-D' + macro)
        status, _, error = run([*command, str(ROOT / 'prototype.c'), '-o', str(binary)],
                               name + '-build', directory)
        if status:
            raise RuntimeError('build failed: ' + error.decode())
        return binary

    control = build('control')
    observations = {}
    for mode in ['success', 'throw', 'cancel']:
        node = run(['node', str(oracle), mode, str(fixture)], mode + '-node', directory)
        native = run([str(control), mode, str(fixture)], mode + '-native', directory, environment)
        assert node[0] == native[0] == 0, (mode, node, native)
        assert node[1] == native[1], (mode, 'Node stdout mismatch')
        assert node[2] == b'', (mode, 'Node stderr')
        counts = re.fullmatch(rb'counts allocations=(\d+) frees=(\d+) retains=(\d+) releases=(\d+)\n', native[2])
        assert counts, (mode, native[2])
        values = tuple(map(int, counts.groups()))
        assert values[0] == values[1] and values[0] + values[2] == values[3], values
        assert values == {"success": (7, 7, 12, 19), "throw": (8, 8, 12, 20),
                          "cancel": (6, 6, 7, 13)}[mode], (mode, values)
        observations[mode] = native
        print(mode, 'Node bytes and exit match;', native[2].decode().strip(), flush=True)

    for name, macro, mode, diagnostic in [
        ('inline', 'MUTANT_INLINE', 'success', None),
        ('checkpoint', 'MUTANT_CHECKPOINT', 'success', None),
        ('borrow', 'MUTANT_BORROW', 'throw', b'heap-use-after-free'),
        ('leak', 'MUTANT_LEAK', 'throw', b'LeakSanitizer: detected memory leaks'),
        ('cancel-leak', 'MUTANT_LEAK', 'cancel', b'LeakSanitizer: detected memory leaks'),
    ]:
        binary = build(name, macro)
        mutant = run([str(binary), mode, str(fixture)], name + '-mutant', directory, environment)
        if diagnostic:
            assert mutant[0] != 0 and diagnostic in mutant[2], (name, mutant)
            caught = diagnostic.decode()
        else:
            assert mutant[0] == 0, (name, 'ordering mutant must finish cleanly', mutant)
            assert mutant[2] == observations[mode][2], (name, 'ordering mutant changed ownership')
            assert mutant[1] != observations[mode][1], (name, 'mutant survived Node comparison')
            caught = 'Node stdout comparison'
        print('mutant', name, 'caught by', caught, flush=True)
    binary = build('counts', 'MUTANT_COUNTS')
    mutant = run([str(binary), 'throw', str(fixture)], 'counts-mutant', directory, environment)
    assert mutant[0] == 0 and mutant[1] == observations['throw'][1], mutant
    assert mutant[2] == b'counts allocations=8 frees=8 retains=13 releases=21\n', mutant
    assert mutant[2] != observations['throw'][2], 'count mutant survived'
    print('mutant counts caught by exact count comparison; Node and sanitizers pass', flush=True)
    print('PASS', flush=True)


if __name__ == '__main__':
    main()
