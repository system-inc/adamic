"""Interleave native output modes and Node, each with stdout a fresh regular file."""
import os
import statistics
import subprocess
import tempfile
import time

commands = {
    'buffer': ['/tmp/stdout-word-buffer'],
    'lines': ['/tmp/stdout-word-lines'],
    'node': ['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', 'bench/word_count.ts'],
}
times = {name: [] for name in commands}
expected = None
for round in range(9):
    order = list(commands)
    if round % 2:
        order.reverse()
    for name in order:
        with tempfile.TemporaryFile() as output:
            started = time.perf_counter()
            child = subprocess.run(commands[name], stdout=output, stderr=subprocess.PIPE, timeout=30)
            elapsed = time.perf_counter() - started
            output.seek(0)
            data = output.read()
        assert child.returncode == 0 and child.stderr == b'', (name, child.returncode, child.stderr)
        if expected is None:
            expected = data
        assert data == expected, (name, data, expected)
        times[name].append(elapsed)
        print(round, name, f'{elapsed:.6f}', len(data), flush=True)
for name, measurements in times.items():
    print(name, 'best', f'{min(measurements):.6f}', 'median', f'{statistics.median(measurements):.6f}')
print('bytes', expected)
