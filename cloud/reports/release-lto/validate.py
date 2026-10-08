#!/usr/bin/env python3
"""Validate ThinLTO AST bytes against Go and Node; prove comparisons with compiled mutants."""
import hashlib
import json
from pathlib import Path
import shlex
import subprocess
import sys

repo = Path.cwd()
scratch = Path(sys.argv[1]).resolve()
clang = '/workspace/adamic-tools/llvm/bin/clang'
flags = ['-std=c11', '-Wall', '-Wextra', '-Werror', '-pedantic', '-Wno-unused-variable', '-Wno-unused-but-set-variable', '-Wno-unused-function', '-Wno-unused-parameter', '-Wno-self-assign', '-ffp-contract=off', '-fno-optimize-sibling-calls', '-O2', '-flto=thin', '-fuse-ld=lld', '-I', str(repo / 'internal/native/runtime')]
with (scratch / 'validation-commands.log').open('w') as commands:
    def run(args, stem, directory=repo):
        commands.write(shlex.join(map(str, args)) + '\n'); commands.flush()
        with (scratch / (stem + '.stdout')).open('wb') as out, (scratch / (stem + '.stderr')).open('wb') as err:
            subprocess.run(list(map(str, args)), cwd=directory, stdout=out, stderr=err, check=True)
    def build(source, binary, mode='thin'):
        compile_flags = flags if mode == 'thin' else [flag for flag in flags if flag not in ('-flto=thin', '-fuse-ld=lld')]
        run([clang, *compile_flags, '-o', scratch / binary, scratch / source, '-Xlinker', '--whole-archive', scratch / mode / 'runtime.a', '-Xlinker', '--no-whole-archive', '-lm'], binary + '-build')
    parser_source = repo / 'stage1/typescript/parser/main.ts'
    run([scratch / 'adamic', 'c', parser_source], 'ast-emit')
    (scratch / 'ast.c').write_bytes((scratch / 'ast-emit.stdout').read_bytes())
    build('ast.c', 'ast-thin')
    args = ['--manifest', scratch / 'compiler.txt', '--whole']
    run([scratch / 'ast-thin', *args], 'ast-thin')
    build('ast.c', 'ast-baseline', 'baseline')
    run([scratch / 'ast-baseline', *args], 'ast-baseline')
    run(['node', '--disable-warning=ExperimentalWarning', repo / 'oracle/node.mjs', parser_source, *args], 'ast-node')
    root = repo / 'cohere/TypeScript/tsc'
    virtual = root / 'adamic_parser_oracle.go'
    overlay = scratch / 'ast-overlay.json'
    overlay.write_text(json.dumps({'Replace': {str(virtual): str(repo / 'stage1/typescript/parser/testdata/oracle.go')}}))
    run(['go', 'build', '-overlay=' + str(overlay), '-o', scratch / 'ast-go', virtual], 'ast-go-build', root)
    run([scratch / 'ast-go', *args], 'ast-go')
    expected = (scratch / 'ast-go.stdout').read_bytes()
    for name in ['ast-thin', 'ast-node', 'ast-baseline']:
        if (scratch / (name + '.stdout')).read_bytes() != expected:
            raise RuntimeError('MISCOMPILE: full AST mismatch: ' + name)
    print('77-file AST byte parity PASS', len(expected), hashlib.sha256(expected).hexdigest(), flush=True)
    # Mutants are separate compiled executables. Neither changes the good binaries or runtime.
    for label, source, before, after, args, expected in [
        ('ast', 'ast.c', 'ADAMIC_STRING("SourceFile")', 'ADAMIC_STRING("XourceFile")', args, expected),
        ('service', 'service.c', r'ADAMIC_STRING("{\"status\":200,\"service\":\"orders\"}")', r'ADAMIC_STRING("{\"status\":201,\"service\":\"orders\"}")', ['/tmp/wasm-requests-profile/requests.jsonl', 'verify'], Path('/tmp/wasm-requests-profile/responses.jsonl').read_bytes() + b'7394547\n'),
    ]:
        text = (scratch / source).read_text()
        if text.count(before) != 1:
            raise RuntimeError('mutant site not unique: ' + label)
        mutant_source = label + '-mutant.c'
        (scratch / mutant_source).write_text(text.replace(before, after))
        binary = label + '-mutant'
        build(mutant_source, binary)
        run([scratch / binary, *args], binary)
        actual = (scratch / (binary + '.stdout')).read_bytes()
        if actual == expected:
            raise RuntimeError('mutant survived: ' + label)
        index = next(i for i, pair in enumerate(zip(actual, expected)) if pair[0] != pair[1])
        print('mutant caught only by byte comparison', label, 'first difference', index, 'lengths', len(actual), len(expected), flush=True)
