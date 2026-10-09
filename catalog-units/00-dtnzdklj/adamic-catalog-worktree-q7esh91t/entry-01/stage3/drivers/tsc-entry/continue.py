#!/usr/bin/env python3
"""Continue body-only stopping-site exploration in a disposable source copy."""
import argparse
import json
import os
from pathlib import Path
import re
import subprocess

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('source', type=Path, help='disposable src directory, never a repository source')
parser.add_argument('compiler', type=Path)
parser.add_argument('output', type=Path, help='new evidence directory')
parser.add_argument('--limit', type=int, default=15)
args = parser.parse_args()
here = Path(__file__).resolve().parent
source = args.source.resolve()
compiler = args.compiler.resolve()
out = args.output.resolve()
if out.exists():
    parser.error('output must be new')
if source.is_relative_to(here.parents[2]) or (source / '.git').exists():
    parser.error('source must be a disposable src copy')
out.mkdir(parents=True)
stops = []
for ordinal in range(1, args.limit + 1):
    streams = []
    exits = []
    for split in (0, 1):
        stem = out / f'{ordinal:02d}-split-{split}'
        env = dict(os.environ, ADAMIC_NATIVE_SPLIT=str(split), ADAMIC_NATIVE_JOBS=subprocess.check_output(['nproc'], text=True).strip())
        command = [str(compiler), 'build', str(source / 'tsc/tsc.ts'), '-o', str(out / f'tsc-{split}')]
        with stem.with_suffix('.stdout').open('wb') as stdout, stem.with_suffix('.stderr').open('wb') as stderr:
            result = subprocess.run(command, env=env, stdout=stdout, stderr=stderr)
        stem.with_suffix('.exit').write_text(str(result.returncode) + '\n')
        streams.append(stem.with_suffix('.stderr').read_text())
        exits.append(result.returncode)
    if streams[0] != streams[1] or exits[0] != exits[1]:
        raise RuntimeError('split results differ')
    if result.returncode == 0:
        (out / 'completed.json').write_text(json.dumps({'ordinal': ordinal}) + '\n')
        break
    first = streams[0].splitlines()[0]
    match = re.match(r'(.+):(\d+):(\d+): (.+)', first)
    if not match:
        raise RuntimeError(f'unlocated stopping message: {first}')
    file, line, column, message = match.groups()
    stop = {'ordinal': ordinal, 'file': 'src/' + Path(file).relative_to(source).as_posix(),
            'line': int(line), 'column': int(column), 'message': message,
            'outside_compiler': not Path(file).is_relative_to(source / 'compiler'),
            'split_0_exit': exits[0], 'split_1_exit': exits[1]}
    stops.append(stop)
    (out / 'stops.json').write_text(json.dumps(stops, indent=2) + '\n')
    print(json.dumps(stop), flush=True)
    if ordinal == args.limit:
        break
    command = ['node', str(here / 'stub.cjs'), file, line, column, str(out / f'{ordinal:02d}-replacement.json')]
    with (out / f'{ordinal:02d}-stub.stdout').open('wb') as stdout, (out / f'{ordinal:02d}-stub.stderr').open('wb') as stderr:
        result = subprocess.run(command, stdout=stdout, stderr=stderr)
    stop['replacement_exit'] = result.returncode
    if result.returncode:
        (out / 'stops.json').write_text(json.dumps(stops, indent=2) + '\n')
        break
