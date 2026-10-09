#!/usr/bin/env python3
"""Time saved native snapshots, Node sources and Go; verify every answer byte."""
import argparse
import json
from pathlib import Path
import statistics
import subprocess
import time


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('directory', type=Path)
    parser.add_argument('cases', type=Path)
    parser.add_argument('--rounds', type=int, default=5)
    args = parser.parse_args()
    directory = args.directory.resolve()
    repo = Path(__file__).resolve().parents[3]
    inputs = ['--cases', str(args.cases.resolve())]
    expected = (directory / 'expected.txt').read_bytes()
    count = args.cases.read_bytes().count(b'\n')
    node = ['node', '--disable-warning=ExperimentalWarning', str(repo / 'oracle/node.mjs')]
    commands = {
        'baseline-native': [str(directory / 'baseline'), *inputs],
        'native': [str(directory / 'final'), *inputs],
        'baseline-Node': [*node, str(directory / 'baseline-source/stage1/cohere/json/main.ts'), *inputs],
        'Node': [*node, str(repo / 'stage1/cohere/json/main.ts'), *inputs],
        'Go': [str(directory / 'go-cohere'), *inputs],
    }
    samples = {name: [] for name in commands}
    loads = []
    for round_number in range(args.rounds):
        for name, command in commands.items():
            stdout = directory / (name + '.timed.stdout')
            stderr = directory / (name + '.timed.stderr')
            with stdout.open('wb') as out, stderr.open('wb') as err:
                started = time.perf_counter()
                subprocess.run(command, stdout=out, stderr=err, check=True)
                elapsed = time.perf_counter() - started
            if stdout.read_bytes() != expected or stderr.stat().st_size:
                raise RuntimeError(f'{name} bytes or stderr differ from Go')
            samples[name].append(elapsed)
            print(f'round {round_number + 1} {name}: {elapsed:.6f}s, {count / elapsed:.2f} texts/s; exact Go bytes', flush=True)
        loads.append(Path('/proc/loadavg').read_text().strip())
    result = dict(texts=count, answer_bytes=len(expected), rounds=samples,
                  median_texts_per_second={name: count / statistics.median(values) for name, values in samples.items()},
                  loadavg=loads)
    # .txt keeps measurement metadata out of the repository's JSON formatter corpus.
    (directory / 'measurements.txt').write_text(json.dumps(result, indent=2) + '\n')
    print(json.dumps(result, indent=2))


if __name__ == '__main__':
    main()
