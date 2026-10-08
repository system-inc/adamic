#!/usr/bin/env python3
"""Compare an identical merged baseline with only unused compiler units omitted."""
from pathlib import Path
import json, os, subprocess, tempfile
repo = Path(__file__).resolve().parents[4]
proof = Path(tempfile.mkdtemp(prefix='regex-runtime-landing-sizes-'))
env = dict(os.environ, GOFLAGS='-buildvcs=false', GOPROXY='https://proxy.golang.org|direct')
files = ['parser', 'sets', 'bytecode', 'v8', 'runtime']
overlay = proof / 'without-compiler.json'
overlay.write_text(json.dumps({'Replace': {str(repo / ('internal/native/runtime/regexp_compile_' + name + '.c')): '' for name in files}}))
rows = []
for version in ['before', 'after']:
    driver = proof / ('adamic-' + version)
    args = ['go', 'build'] + (['-overlay=' + str(overlay)] if version == 'before' else []) + ['-o', str(driver), './cmd/adamic']
    with (proof / (version + '-driver.log')).open('w') as log:
        subprocess.run(args, cwd=repo, env=env, stdout=log, stderr=subprocess.STDOUT, check=True)
    output = proof / version
    output.mkdir()
    fixtures = [('hello', 'internal/load/testdata/0.1/compile/01_hello.ts'), ('request', 'cmd/adamic/testdata/wasi/request.a')]
    if version == 'after': fixtures.append(('dynamic_gap', 'internal/oracle/testdata/regexp_dynamic/dynamic_gap.a'))
    for name, fixture in fixtures:
        binary = output / (name + '.wasm')
        with (output / (name + '.build.log')).open('w') as log:
            subprocess.run([str(driver), 'build', '--target', 'wasm32-wasi', str(repo / fixture), '-o', str(binary)], cwd=repo, env=env, stdout=log, stderr=subprocess.STDOUT, check=True)
        packed = subprocess.check_output(['node', '-e', "const fs=require('fs'),z=require('zlib');const b=fs.readFileSync(process.argv[1]);console.log(z.brotliCompressSync(b,{params:{[z.constants.BROTLI_PARAM_QUALITY]:11}}).length)", str(binary)], env=env)
        rows.append(dict(version=version, name=name, raw=binary.stat().st_size, brotli=int(packed)))
for name in ['hello', 'request']:
    assert (proof / 'before' / (name + '.wasm')).read_bytes() == (proof / 'after' / (name + '.wasm')).read_bytes(), name + ' WASI bytes differ'
(proof / 'sizes.json').write_text(json.dumps(rows, indent=2) + '\n')
print(json.dumps(rows, indent=2), flush=True)
print('byte-identical hello/request; proof directory: ' + str(proof), flush=True)
