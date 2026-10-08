#!/usr/bin/env python3
"""Linux LSan controls and cycle-preserving mutants against the production runtime."""
import os
import sys
sys.dont_write_bytecode = True
from pathlib import Path
import tempfile
from check import run, FLAGS

ROOT = Path(__file__).resolve().parent
RUNTIME = ROOT.parents[1] / 'internal/native/runtime'
def main():
    directory = Path(tempfile.mkdtemp(prefix='p286ycm-cycles-'))
    print('artifacts:', directory, flush=True)
    env = dict(os.environ, ASAN_OPTIONS='detect_leaks=1:halt_on_error=1', UBSAN_OPTIONS='halt_on_error=1')
    # Compile unrelated runtime units once. Each subprocess has a 30s bound, even on a busy gate.
    common = []
    for source in sorted(RUNTIME.glob('*.c')):
        if source.name == 'async.c':
            continue
        target = directory / (source.stem + '.o')
        result = run(['clang', *FLAGS, '-I', str(RUNTIME), '-c', str(source), '-o', str(target)],
                     source.stem + '-build', directory)
        assert result[0] == 0, result
        common.append(str(target))

    def build(name, macro=None):
        binary = directory / name
        target = directory / (name + '-async.o')
        command = ['clang', *FLAGS, '-I', str(RUNTIME)]
        if macro:
            command.append('-D' + macro)
        result = run([*command, '-c', str(RUNTIME / 'async.c'), '-o', str(target)],
                     name + '-async-build', directory)
        assert result[0] == 0, result
        result = run(['clang', *FLAGS, '-I', str(RUNTIME), str(ROOT / 'cycles.c'),
                      str(target), *common, '-lm', '-o', str(binary)], name + '-build', directory)
        assert result[0] == 0, result
        return binary
    control = build('control')
    for mode in ['settle', 'cancel', 'exit']:
        result = run([str(control), mode], mode + '-control', directory, env)
        assert result == (0, b'cycle clean\n', b''), result
        print(mode, 'ASan/UBSan/LSan clean', flush=True)
        mutant = build(mode, 'ADAMIC_ASYNC_' + mode.upper() + '_CYCLE_MUTANT')
        result = run([str(mutant), mode], mode + '-mutant', directory, env)
        assert result[0] != 0 and b'LeakSanitizer: detected memory leaks' in result[2], result
        print('mutant', mode, 'caught by LeakSanitizer', flush=True)
    mutant = build('next-link', 'ADAMIC_ASYNC_NEXT_LINK_MUTANT')
    result = run([str(mutant), 'cancel'], 'next-link-mutant', directory, env)
    assert result[0] != 0 and b'heap-use-after-free' in result[2], result
    print('next-link mutant caught by ASan with two observers', flush=True)
    oracle = directory / 'oracle.mjs'
    oracle.write_bytes((ROOT / 'cycles_oracle.a').read_bytes())
    node = run(['node', str(oracle)], 'fulfilled-node', directory)
    result = run([str(control), 'fulfilled'], 'fulfilled-control', directory, env)
    assert result == node == (0, b'prefix\nresume\nresume\ncycle clean\n', b''), (result, node)
    mutant = build('inline', 'ADAMIC_ASYNC_INLINE_MUTANT')
    result = run([str(mutant), 'fulfilled'], 'fulfilled-mutant', directory, env)
    assert result[0] == 0 and result[2] == b'' and result[1] != node[1], result
    print('fulfilled Node bytes match; inline mutant caught only by stdout comparison', flush=True)
    print('PASS', flush=True)
if __name__ == '__main__': main()
