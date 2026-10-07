#!/usr/bin/env python3
"""Compile real compiler mutants in Go overlays; leave the checkout unchanged."""
import json
import os
from pathlib import Path
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[4]
SOURCE = ROOT / 'internal/lower/moves.go'

def command(args, log, *, env=None, expected=0):
    with open(log, 'w') as stream:
        result = subprocess.run(args, cwd=ROOT, env=env, stdout=stream,
                                stderr=subprocess.STDOUT, timeout=600)
    if result.returncode != expected:
        raise RuntimeError(f'{args}: exit {result.returncode}; see {log}')

def main():
    directory = Path(tempfile.mkdtemp(prefix='adamic-move-mutants-'))
    original = SOURCE.read_text()
    # Only the literal provenance rule is broken. Items still have a sole source
    # use and work still satisfies every effect/result check.
    seam = '\t\tif element.Kind != ast.KindObjectLiteralExpression {\n'
    assert original.count(seam) == 1
    insertion = '''\t\tif ast.IsIdentifier(element) {
\t\t\tif source := proof.declaration(element); source != nil && source.Kind == ast.KindVariableDeclaration {
\t\t\t\telement = ast.SkipParentheses(source.AsVariableDeclaration().Initializer)
\t\t\t}
\t\t}
'''
    alias = directory / 'alias.go'
    alias.write_text(original.replace(seam, insertion + seam))
    overlay = directory / 'alias.json'
    overlay.write_text(json.dumps({'Replace': {str(SOURCE): str(alias)}}))
    compiler = directory / 'compiler'
    command(['go', 'build', '-overlay', str(overlay), '-o', str(compiler), './cmd/adamic'], directory/'alias-build.log')
    cfile = directory / 'alias.c'
    fixture = 'internal/oracle/testdata/moves/refused/race_alias.a'
    with cfile.open('w') as output, (directory/'alias-lower.log').open('w') as error:
        result = subprocess.run([str(compiler), 'c', fixture], cwd=ROOT,
                                stdout=output, stderr=error, timeout=60)
    if result.returncode != 0:
        raise RuntimeError(f'alias mutant did not lower: {directory}/alias-lower.log')
    binary = directory / 'alias'
    runtime = ROOT / 'internal/native/runtime'
    flags = ['-std=c11', '-Wall', '-Wextra', '-Werror', '-pedantic',
             '-Wno-unused-variable', '-Wno-unused-but-set-variable',
             '-Wno-unused-function', '-Wno-unused-parameter', '-Wno-self-assign',
             '-ffp-contract=off', '-fno-optimize-sibling-calls', '-pthread',
             '-O1', '-g', '-fsanitize=thread']
    command(['clang', *flags, '-I', str(runtime), str(cfile),
             *map(str, sorted(runtime.glob('*.c'))), '-lm', '-o', str(binary)],
            directory/'alias-clang.log')
    environment = os.environ.copy()
    environment.update(ADAMIC_THREADS='4', TSAN_OPTIONS='halt_on_error=1')
    command([str(binary)], directory/'alias-race.log', env=environment, expected=66)
    race = (directory/'alias-race.log').read_text()
    if 'WARNING: ThreadSanitizer: data race' not in race:
        raise RuntimeError('exit 66 without a TSan race is not a proof')
    print('aliased-element mutant:', next(line for line in race.splitlines()
                                          if 'SUMMARY: ThreadSanitizer:' in line))
    # Only later source uses are allowed; the old binding is still consumed.
    seam = '\t\t\tif child.Pos() > node.End() {\n'
    assert original.count(seam) == 1
    after = directory/'after.go'
    after.write_text(original.replace(seam, seam+'\t\t\t\treturn false // mutant: allow use after transfer\n'))
    overlay = directory/'after.json'
    overlay.write_text(json.dumps({'Replace': {str(SOURCE): str(after)}}))
    command(['go', 'test', '-overlay', str(overlay), './internal/oracle', '-run',
             '^TestMovesRefusals$/^after.a$', '-count=1', '-v'],
            directory/'after-test.log', expected=1)
    report = (directory/'after-test.log').read_text()
    if 'got <nil>' not in report:
        raise RuntimeError('use-after-move mutant failed for the wrong reason')
    print('use-after-move mutant: exact refusal expected; got <nil>, go test exit 1')
    command(['go','test','./internal/oracle','-run','^TestMovesPlainCounts$',
             '-count=1','-v'],directory/'shared-count-test.log')
    report = (directory/'shared-count-test.log').read_text()
    if 'plain 2050 shared 0' not in report or 'plain 1 shared 2049' not in report:
        raise RuntimeError('gratuitous-sharing mutant was not observed')
    print('shared-marking mutant: plain/shared graph counts 2050/0 became 1/2049')
    print('logs:', directory)

if __name__ == '__main__':
    main()
