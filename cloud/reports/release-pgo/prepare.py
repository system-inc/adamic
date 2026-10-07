#!/usr/bin/env python3
"""Build held-out PGO variants from the committed ThinLTO unit's exact snapshots."""
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import time

s = Path(sys.argv[1]).resolve()
lto = Path(sys.argv[2]).resolve()
s.mkdir(parents=True, exist_ok=True)
compiler = '/workspace/adamic-tools/llvm/bin/clang'
archiver = '/workspace/adamic-tools/llvm/bin/llvm-ar'
profdata = '/workspace/adamic-tools/llvm/bin/llvm-profdata'
rows = sorted(lto.joinpath('compiler.txt').read_text().splitlines())
if len(rows) != 77 or len(set(rows)) != 77:
    raise RuntimeError('expected 77 distinct corpus files')
train, held = rows[1::2], rows[::2]
if len(train) != 38 or len(held) != 39 or set(train) & set(held):
    raise RuntimeError('invalid held-out split')
for name, values in [('train', train), ('held-out', held)]:
    (s / (name + '.txt')).write_text('\n'.join(values) + '\n')
    (s / (name + '-files.json')).write_text(json.dumps([dict(name='src/compiler/' + p.split('/src/compiler/')[1], sha256=hashlib.sha256(Path(p).read_bytes()).hexdigest()) for p in values], indent=2) + '\n')
for file in ['parse.c', 'ast.c']:
    shutil.copyfile(lto / file, s / file)
runtime = s / 'runtime'
runtime.mkdir(exist_ok=True)
for p in (lto / 'thin').iterdir():
    if p.suffix in ['.c', '.h']:
        shutil.copyfile(p, runtime / p.name)
flags = json.loads((lto / 'thin-parse-build.json').read_text())['Compile']
profile = str(s / 'training.profdata')
variants = {
    'o2': [f for f in flags if f != '-flto=thin'],
    'thin': flags,
    'generate': flags + ['-fprofile-instr-generate'],
    'pgo': flags + ['-fprofile-use=' + profile, '-Wno-profile-instr-out-of-date'],
    'split': flags + ['-fprofile-use=' + profile, '-Wno-profile-instr-out-of-date', '-fsplit-machine-functions'],
}
commands = (s / 'commands.jsonl').open('w')
builds = {}
def run(args, stem, env=None):
    commands.write(json.dumps(list(map(str, args))) + '\n'); commands.flush()
    with (s / (stem + '.stdout')).open('wb') as out, (s / (stem + '.stderr')).open('wb') as err:
        subprocess.run(list(map(str, args)), stdout=out, stderr=err, env=env, check=True)

def build(mode):
    start = time.monotonic()
    dest = s / mode; dest.mkdir(exist_ok=True)
    objects = []
    for source in sorted(runtime.glob('*.c')):
        obj = dest / (source.stem + '.o')
        run([compiler, *variants[mode], '-c', source, '-o', obj], mode + '-compile-' + source.stem)
        objects.append(obj)
    archive = dest / 'runtime.a'
    run([archiver, 'rcs', archive, *objects], mode + '-archive')
    link = variants[mode] + ([] if mode == 'o2' else ['-fuse-ld=lld'])
    run([compiler, *link, '-I', runtime, '-o', dest / 'parse', s / 'parse.c', '-Xlinker', '--whole-archive', archive, '-Xlinker', '--no-whole-archive', '-lm'], mode + '-link')
    run(['/workspace/adamic-tools/llvm/bin/llvm-size', '--format=sysv', dest / 'parse'], mode + '-sections')
    sections = (s / (mode + '-sections.stdout')).read_text()
    text = sum(int(line.split()[1]) for line in sections.splitlines() if line.split() and line.split()[0].startswith('.text'))
    builds[mode] = dict(wall=time.monotonic()-start, bytes=(dest/'parse').stat().st_size, text=text, compile_flags=variants[mode], link_flags=link)
    (s / 'builds.json').write_text(json.dumps(builds, indent=2) + '\n')
    print('built', mode, json.dumps(builds[mode]), flush=True)

for mode in ['o2', 'thin', 'generate']:
    build(mode)
env = dict(os.environ, LLVM_PROFILE_FILE=str(s / 'training.profraw'))
run(['taskset', '-c', '3', s / 'generate/parse', '--manifest', s / 'train.txt', '--count'], 'training-run', env)
if (s / 'training-run.stdout').read_bytes() != b'0\n' or (s / 'training-run.stderr').read_bytes():
    raise RuntimeError('training output differs')
run([profdata, 'merge', s / 'training.profraw', '-o', s / 'training.profdata'], 'merge')
run([profdata, 'show', '--all-functions', '--counts', s / 'training.profdata'], 'profile-functions')
print('trained on exactly 38 files; profile', hashlib.sha256((s/'training.profdata').read_bytes()).hexdigest(), flush=True)
for mode in ['pgo', 'split']:
    build(mode)
# All generated-C variants must keep the parse driver's exact output and status.
for mode in ['o2', 'thin', 'pgo', 'split']:
    run(['taskset', '-c', '3', s / mode / 'parse', '--manifest', s / 'held-out.txt', '--count'], mode + '-held-count')
    if (s / (mode + '-held-count.stdout')).read_bytes() != b'0\n' or (s / (mode + '-held-count.stderr')).read_bytes():
        raise RuntimeError('MISCOMPILE: held-out parse output differs: ' + mode)
print('all four held-out parse outputs PASS', flush=True)
commands.close()
