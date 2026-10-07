#!/usr/bin/env python3
"""Search the same 300 move seeds against isolated compiler mutants."""
import argparse
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile
import time

ROOT = Path(__file__).resolve().parents[4]
SOURCE = ROOT / 'internal/lower/moves.go'

def command(args, log, *, env=None, expected=0):
    started = time.monotonic()
    with log.open('w') as output:
        result = subprocess.run(args, cwd=ROOT, env=env, stdout=output,
                                stderr=subprocess.STDOUT, timeout=1800)
    if result.returncode != expected:
        raise RuntimeError(f'{args}: exit {result.returncode}, see {log}')
    return time.monotonic() - started

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--fuzzer', required=True, help='built cmd/adamic-fuzz from this branch')
    args = parser.parse_args()
    directory = Path(tempfile.mkdtemp(prefix='adamic-fuzz-moves-mutants-'))
    original = SOURCE.read_text()
    seam = '\t\tif element.Kind != ast.KindObjectLiteralExpression {\n'
    assert original.count(seam) == 1
    insertion = '''\t\tif ast.IsIdentifier(element) {
\t\t\tif source := proof.declaration(element); source != nil && source.Kind == ast.KindVariableDeclaration {
\t\t\t\telement = ast.SkipParentheses(source.AsVariableDeclaration().Initializer)
\t\t\t}
\t\t}
'''
    alias = original.replace(seam, insertion+seam)
    seam = '\t\t\tif value.Kind != ast.KindNumericLiteral && value.Kind != ast.KindTrueKeyword && value.Kind != ast.KindFalseKeyword {\n'
    assert original.count(seam) == 1
    nested = original.replace(seam, '\t\t\tif false { // mutant: treat nested graph as flat\n')
    evidence = []
    for name, mutated in [('alias', alias), ('nested-flat', nested)]:
        replacement = directory/(name+'.go')
        replacement.write_text(mutated)
        overlay = directory/(name+'.json')
        overlay.write_text(json.dumps({'Replace': {str(SOURCE): str(replacement)}}))
        environment = os.environ.copy()
        environment['GOFLAGS'] = environment.get('GOFLAGS', '')+' -overlay='+str(overlay)
        environment['ADAMIC_FUZZ_MOVES_SEEDS'] = '300'
        log = directory/(name+'-search.log')
        seconds = command(['go', 'test', './internal/fuzz', '-run', '^TestMovesCampaign$',
                           '-count=1', '-timeout', '30m', '-v'], log, env=environment, expected=1)
        report = log.read_text()
        match = re.search(r'seed (\d+) after .*: finding compiler accepted a program it must refuse', report)
        if not match:
            raise RuntimeError(f'{name}: search failed for another reason, see {log}')
        seed = int(match[1])
        assert 1 <= seed <= 300
        print(f'{name}: found seed {seed} within 300, {seconds:.3f}s', flush=True)
        fixture = directory/(name+'.a')
        with fixture.open('w') as output:
            subprocess.run([args.fuzzer, '-only-moves', '-seed', str(seed), '-print'],
                           cwd=ROOT, stdout=output, check=True)
        # The exact refusal already catches the regression. Remove only its
        # expectation comments to execute this same generated graph as a second,
        # independent witness of why accepting it is unsound.
        source = fixture.read_text()
        shape = re.search(r'// moves-shape: (.*)', source).group(1)
        fixture.write_text('\n'.join(line for line in source.splitlines()
                            if not line.startswith('// moves-refuse:')
                            and not line.startswith('// moves-fix:'))+'\n')
        compiler = directory/(name+'-compiler')
        command(['go', 'build', '-o', str(compiler), './cmd/adamic'], directory/(name+'-build.log'), env=environment)
        cfile = directory/(name+'.c')
        with cfile.open('w') as output, (directory/(name+'-lower.log')).open('w') as error:
            result = subprocess.run([str(compiler), 'c', str(fixture)], cwd=ROOT,
                                    stdout=output, stderr=error, timeout=60)
        if result.returncode:
            raise RuntimeError(f'{name}: witness failed to lower')
        runtime = ROOT/'internal/native/runtime'
        binary = directory/name
        flags = ['-std=c11', '-Wall', '-Wextra', '-Werror', '-pedantic',
                 '-Wno-unused-variable', '-Wno-unused-but-set-variable',
                 '-Wno-unused-function', '-Wno-unused-parameter', '-Wno-self-assign',
                 '-ffp-contract=off', '-fno-optimize-sibling-calls', '-pthread',
                 '-O1', '-g', '-fsanitize=thread']
        command(['clang', *flags, '-I', str(runtime), str(cfile),
                 *map(str, sorted(runtime.glob('*.c'))), '-lm', '-o', str(binary)],
                directory/(name+'-clang.log'))
        node_log = directory/(name+'-node.log')
        command(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(fixture)], node_log)
        sequential = os.environ.copy()
        sequential.update(ADAMIC_THREADS='1', TSAN_OPTIONS='halt_on_error=1')
        one_log = directory/(name+'-one.log')
        command([str(binary)], one_log, env=sequential)
        if node_log.read_bytes() != one_log.read_bytes():
            raise RuntimeError(f'{name}: sequential witness disagrees with Node')
        concurrent = sequential.copy()
        concurrent.pop('ADAMIC_THREADS')
        race_log = directory/(name+'-race.log')
        command([str(binary)], race_log, env=concurrent, expected=66)
        race = race_log.read_text()
        if 'WARNING: ThreadSanitizer: data race' not in race:
            raise RuntimeError(f'{name}: exit 66 without a TSan race is no proof')
        summary = next(line for line in race.splitlines() if 'SUMMARY: ThreadSanitizer:' in line)
        print(f'{name}: {shape}; one thread matches Node; {summary}', flush=True)
        evidence.append({'mutant': name, 'seed': seed, 'shape': shape,
                         'seed_limit': 300, 'seconds_including_build': seconds,
                         'detector': 'compiler accepted a program it must refuse',
                         'sequential_stdout': node_log.read_text(),
                         'tsan_exit': 66, 'tsan_summary': summary})
    # Prove the fuzzer's exact matcher and boundary-preserving shrinker too.
    checks = [
        ('substring', ROOT/'internal/fuzz/moves.go',
         '!present || actual != expected', '!present || !strings.Contains(actual, what)',
         'TestMovesRunnerCanFail', 'false success'),
        ('shrink-boundary', ROOT/'internal/fuzz/shrink.go',
         'if s.current.RefusalFix != "" && !strings.Contains(source, "parallelMap(") {',
         'if false {', 'TestMovesShrinkKeepsBoundary', 'lost the boundary'),
    ]
    for name, source, seam, change, test, witness in checks:
        original = source.read_text()
        assert original.count(seam) == 1
        replacement = directory/(name+'.go')
        replacement.write_text(original.replace(seam, change))
        overlay = directory/(name+'.json')
        overlay.write_text(json.dumps({'Replace': {str(source): str(replacement)}}))
        log = directory/(name+'-test.log')
        command(['go', 'test', '-overlay', str(overlay), './internal/fuzz', '-run',
                 '^'+test+'$', '-count=1', '-v'], log, expected=1)
        if witness not in log.read_text():
            raise RuntimeError(f'{name}: check failed for another reason')
        print(f'{name}: assertion caught mutant, go test exit 1', flush=True)
        evidence.append({'mutant': name, 'detector': test, 'exit': 1, 'witness': witness})
    (directory/'evidence.json').write_text(json.dumps(evidence, indent=2)+'\n')
    print('logs:', directory, flush=True)

if __name__ == '__main__':
    main()
