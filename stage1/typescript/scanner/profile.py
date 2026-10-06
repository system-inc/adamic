#!/usr/bin/env python3
"""Profile TestProfileArtifacts snapshots; all program output goes to files."""
import argparse
import collections
import json
import os
from pathlib import Path
import re
import subprocess
import time


def summarize(path):
    # DWARF inline records can split one function among main.c and adamic.h,
    # or add a same-name call edge. Collapse those records, not their costs twice.
    symbols = {}
    files = {}
    source_file = "???"
    source_line = 0
    line_self = collections.Counter()
    own = collections.Counter()
    edges = collections.Counter()
    calls = collections.Counter()
    function = callee = None
    pending = False
    count = 0
    total = None
    for line in path.read_text().splitlines():
        if line.startswith('summary:'):
            total = int(line.split()[1])
        elif line.startswith(('fl=', 'fi=', 'fe=', 'cfl=', 'cfi=')):
            value = line.split('=', 1)[1]
            match = re.match(r'\((\d+)\)(?: (.*))?$', value)
            if match:
                identifier, name = match.groups()
                if name is not None:
                    files[identifier] = name
                value = files[identifier]
            if not line.startswith('c'):
                source_file = Path(value).name
        elif line.startswith(('fn=', 'cfn=')):
            value = line.split('=', 1)[1]
            match = re.match(r'\((\d+)\)(?: (.*))?$', value)
            if match:
                identifier, name = match.groups()
                if name is not None:
                    symbols[identifier] = name
                value = symbols[identifier]
            if line.startswith('fn='):
                function = value
            else:
                callee = value
        elif line.startswith('calls='):
            pending = True
            count = int(line[6:].split()[0])
        elif function and line and line[0] in '0123456789+-*':
            position = line.split()[0]
            if position != '*':
                source_line = source_line + int(position) if position[0] in '+-' else int(position)
            cost = int(line.split()[-1])  # The recorded event is Ir alone.
            if pending:
                edges[function, callee] += cost
                calls[function, callee] += count
                pending = False
            else:
                own[function] += cost
                line_self[source_file, source_line, function] += cost
    if total is None or sum(own.values()) != total:
        raise RuntimeError("callgrind self costs do not sum to summary")
    inclusive = own.copy()
    for (caller, target), cost in edges.items():
        if caller != target:
            inclusive[caller] += cost
    def rows(counter):
        return [dict(function=name, instructions=value,
                     percent=100 * value / total,
                     self=own[name], inclusive=inclusive[name])
                for name, value in counter.most_common()
                if name not in ('(below main)',) and not name.startswith('0x')]
    result = dict(total=total, inclusive=rows(inclusive), self=rows(own), self_all=dict(own),
                  line_self=[dict(file=f, line=n, function=fn, instructions=value)
                             for (f, n, fn), value in line_self.most_common()],
                  edges=[dict(caller=a, callee=b, instructions=value, calls=calls[a, b])
                         for (a, b), value in edges.most_common()])
    path.with_suffix('.json').write_text(json.dumps(result, indent=2) + '\n')
    return result


def run(command, directory, stem, env=None):
    with (directory / (stem + '.stdout')).open('wb') as out, (directory / (stem + '.stderr')).open('wb') as err:
        started = time.perf_counter()
        subprocess.run(command, stdout=out, stderr=err, env=env, check=True)
        duration = time.perf_counter() - started
    return duration, (directory / (stem + '.stdout')).read_bytes()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('directory', type=Path)
    parser.add_argument('--valgrind', default='valgrind')
    parser.add_argument('--summarize-only', action='store_true')
    args = parser.parse_args()
    directory = args.directory.resolve()
    profile = directory / 'callgrind.out'
    if not args.summarize_only:
        manifest = ['--manifest', str(directory / 'compiler.txt'), '--count']
        repo = Path(__file__).resolve().parents[3]
        commands = dict(native=[str(directory / 'scanner'), *manifest],
                        Go=[str(directory / 'oracle'), *manifest],
                        Node=['node', '--disable-warning=ExperimentalWarning', str(repo / 'oracle/node.mjs'), str(directory / 'main.ts'), *manifest])
        samples = {name: [] for name in commands}
        want = None
        for round_number in range(1, 6):
            for name, command in commands.items():
                duration, answer = run(command, directory, f'round-{round_number}-{name}')
                if want is None:
                    want = answer
                if answer != want or (directory / f'round-{round_number}-{name}.stderr').stat().st_size:
                    raise RuntimeError(f'{name} count/stderr differs')
                samples[name].append(duration)
        tokens = int(want)
        measurements = dict(tokens=tokens, rounds=samples,
                            best={name: dict(seconds=min(values), tokens_per_second=tokens / min(values))
                                  for name, values in samples.items()},
                            loadavg=Path('/proc/loadavg').read_text().strip())
        _, answer = run([str(directory / 'counted'), *manifest], directory, 'counted')
        if answer != want:
            raise RuntimeError('counted build count differs')
        _, answer = run([args.valgrind, '--tool=callgrind', '--callgrind-out-file=' + str(profile),
                         str(directory / 'profiled'), *manifest], directory, 'callgrind')
        if answer != want:
            raise RuntimeError('profiled build count differs')
        measurements['counts'] = (directory / 'counted.stderr').read_text().strip()
    else:
        measurements = {}
    result = summarize(profile)
    measurements['instructions'] = result['total']
    if 'tokens' in measurements:
        measurements['instructions_per_token'] = result['total'] / measurements['tokens']
        (directory / 'measurements.json').write_text(json.dumps(measurements, indent=2) + '\n')
    for mode in ('inclusive', 'self'):
        print(f'\nTop 20 {mode} instructions (same-name inline records collapsed):')
        for row in result[mode][:20]:
            print(f'{row["instructions"]:>14,} {row["percent"]:6.2f}% {row["function"]}')
    print('\n' + json.dumps(measurements, indent=2))


if __name__ == '__main__':
    main()
