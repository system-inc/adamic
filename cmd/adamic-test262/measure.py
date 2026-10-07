#!/usr/bin/env python3
"""Interleave the original serial binary and this runner on the same pinned corpus."""
import argparse
import json
import os
from pathlib import Path
import shlex
import subprocess
import sys
sys.path.insert(0, str(__import__("pathlib").Path(__file__).resolve().parents[2] / "internal" / "boundedrun"))
from python import run as bounded_run
import time

parser = argparse.ArgumentParser()
parser.add_argument('--before', required=True)
parser.add_argument('--after', required=True)
parser.add_argument('--test262', required=True)
parser.add_argument('--output', required=True)
arguments = parser.parse_args()
output = Path(arguments.output)
output.mkdir(parents=True, exist_ok=True)

def capture(command):
    return bounded_run(command, stdout=subprocess.PIPE, text=True, check=True).stdout.strip()

identity = {
    'commit': capture(['git', 'rev-parse', 'HEAD']),
    'test262': capture(['git', '-C', arguments.test262, 'rev-parse', 'HEAD']),
    'nproc': capture(['nproc']),
    'cpu.max': Path('/sys/fs/cgroup/cpu.max').read_text().strip(),
    'go': capture(['go', 'version']),
    'clang': capture(['clang', '--version']).splitlines()[0],
    'node': capture(['node', '--version']),
    'go build': 'go build -o <binary> ./cmd/adamic-test262',
    'native flags': '-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable '
                    '-Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter '
                    '-Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls '
                    '-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all',
}
filters = [
    ('abs', arguments.test262, 'built-ins/Math/abs'),
    ('padStart', arguments.test262, 'built-ins/String/prototype/padStart'),
    ('Math', arguments.test262, 'built-ins/Math'),
    ('sort', arguments.test262, 'built-ins/Array/prototype/sort'),
    ('mini', str(Path('cmd/adamic-test262/testdata/mini').resolve()), ''),
]
go_cache = capture(["go", "env", "GOCACHE"])
records = []
for label, corpus, directory in filters:
    for round_number in range(3):
        cache = output / ('cache-' + label + '-' + str(round_number))
        environment = dict(os.environ, XDG_CACHE_HOME=str(cache.resolve()),
                           ADAMIC_GATE_UNCACHED='0', GOMAXPROCS='4', GOCACHE=go_cache)
        for temperature in ['cold', 'warm']:
            pair = []
            logs = []
            for variant, binary in [('before', arguments.before), ('after', arguments.after)]:
                name = f'{label}-{round_number}-{temperature}-{variant}'
                command = [binary, '-adapt', '-json', '-test262', corpus,
                           '-work', str((output / ('work-' + label + '-' + variant)).resolve())]
                if variant == 'after':
                    command += ['-jobs', '4']
                command.append(directory)
                load_before = Path('/proc/loadavg').read_text().strip()
                start = time.monotonic()
                with (output / (name + '.json')).open('w') as stdout, (output / (name + '.log')).open('w') as stderr:
                    result = bounded_run(command, env=environment, stdout=stdout, stderr=stderr)
                seconds = time.monotonic() - start
                record = dict(identity, loop=label, round=round_number, temperature=temperature,
                              variant=variant, seconds=seconds, exit=result.returncode,
                              load_before=load_before, load_after=Path('/proc/loadavg').read_text().strip(),
                              cache=('no result cache in original runner' if variant == 'before' else temperature),
                              instrument='XDG_CACHE_HOME=' + shlex.quote(str(cache.resolve())) +
                                         ' ADAMIC_GATE_UNCACHED=0 GOMAXPROCS=4 GOCACHE=' + shlex.quote(go_cache) + ' ' + shlex.join(command))
                records.append(record)
                (output / 'measurements.json').write_text(json.dumps(records, indent=2) + '\n')
                print(name, f'{seconds:.3f}s', 'exit=' + str(result.returncode), flush=True)
                if result.returncode:
                    raise RuntimeError(name + ' failed; see its log')
                pair.append((output / (name + '.json')).read_bytes())
                logs.append((output / (name + '.log')).read_bytes())
            if pair[0] != pair[1] or logs[0] != logs[1]:
                raise RuntimeError(label + ' serial/parallel JSON or ordered table/progress differs')
print('all JSON, tables and progress logs byte-identical', flush=True)
