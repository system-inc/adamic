"""Build release artifacts and interleave pinned sort-only user CPU observations."""
from pathlib import Path
import json
import os
import re
import statistics
import subprocess
import sys

root = Path(__file__).resolve().parents[4]
folder = Path(__file__).resolve().parent
output = Path(sys.argv[1] if len(sys.argv) > 1 else '/tmp/adamic-typed-sort-bench')
output.mkdir(exist_ok=True, parents=True)
with (output / 'build.log').open('w') as log:
    subprocess.run(['go', 'build', '-o', str(output / 'adamic'), './cmd/adamic'], cwd=root, stdout=log, stderr=subprocess.STDOUT, check=True)
    for name in ['typed', 'numbers']:
        emitted = subprocess.run([str(output / 'adamic'), 'c', str(folder / f'{name}.a')], cwd=root, stdout=subprocess.PIPE, stderr=log, check=True).stdout.decode()
        pattern = r'(?m)^([ \t]*)(adamic_(?:typed_array|array)_sort\([^\n]+;)$'
        def instrument(match):
            indent, call = match.groups()
            return '\n'.join(indent + line for line in [
                'struct rusage sort_started, sort_finished;',
                'if (getrusage(RUSAGE_SELF, &sort_started) != 0) { abort(); }',
                call,
                'if (getrusage(RUSAGE_SELF, &sort_finished) != 0) { abort(); }',
                'fprintf(stderr, "sort_user=%.6f\\n", (double)(sort_finished.ru_utime.tv_sec - sort_started.ru_utime.tv_sec) + (double)(sort_finished.ru_utime.tv_usec - sort_started.ru_utime.tv_usec) / 1000000.0);',
            ])
        emitted, count = re.subn(pattern, instrument, emitted)
        assert count == 1, (name, count)
        source = output / f'{name}.c'
        source.write_text('#include <sys/resource.h>\n#include <stdio.h>\n#include <stdlib.h>\n' + emitted)
        subprocess.run(['go', 'run', str(folder / 'build.go'), str(source), str(output / name)], cwd=root, stdout=log, stderr=subprocess.STDOUT, check=True)
    subprocess.run([str(output / 'adamic'), 'js', str(folder / 'typed.a')], cwd=root, stdout=(output / 'typed.mjs').open('w'), stderr=log, check=True)
if '--build-only' in sys.argv:
    print(f'built release artifacts in {output}')
    sys.exit(0)
cpu = min(os.sched_getaffinity(0))
commands = {'typed': [str(output / 'typed')], 'numbers': [str(output / 'numbers')], 'node': ['node', str(folder / 'node.a')]}
measurements = {name: [] for name in commands}
expected = None
for round in range(7):
    order = list(commands)
    if round % 2:
        order.reverse()
    for name in order:
        observed = subprocess.run(['taskset', '-c', str(cpu), *commands[name]], stdout=subprocess.PIPE, stderr=subprocess.PIPE, check=True)
        (output / f'{round}-{name}.stdout').write_bytes(observed.stdout)
        (output / f'{round}-{name}.stderr').write_bytes(observed.stderr)
        if expected is None:
            expected = observed.stdout
        assert observed.stdout == expected and b'ordered true' in observed.stdout, (round, name, observed.stdout)
        user = re.fullmatch(rb'sort_user=([0-9.]+)\n', observed.stderr)
        assert user, observed.stderr
        measurements[name].append(float(user[1]))
print(json.dumps({'cpu': cpu, 'seed': 42, 'length': 1000000, 'rounds': measurements,
    'median_user': {name: statistics.median(times) for name, times in measurements.items()},
    'best_user': {name: min(times) for name, times in measurements.items()},
    'stdout': expected.decode(), 'load': os.getloadavg()}, indent=2))
