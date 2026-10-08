#!/usr/bin/env python3
"""Pinned, warmed timed regions; raw trials and profiling output stay outside the checkout."""
import argparse
import json
import hashlib
import math
import os
from pathlib import Path
import re
import shutil
import subprocess

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent
MARKER = 'console.log(`${run(8)}`);'


def command(args, *, cwd=ROOT, env=None):
    result = subprocess.run(list(map(str, args)), cwd=cwd, env=env,
                            capture_output=True, text=True, check=False)
    if result.returncode:
        raise RuntimeError(f'{args}: exit {result.returncode}\n{result.stdout}\n{result.stderr}')
    return result


def timed_source(source, iterations):
    if source.count(MARKER) != 1:
        raise ValueError('exactly one run marker required')
    tail = f'''run({iterations}); run({iterations}); run({iterations});
const start = performance.now();
const answer = run({iterations});
const finish = performance.now();
console.log(`${{answer}}`);
console.error(`scout36 elapsed ${{finish-start}}`);'''
    return 'import {performance} from "node:perf_hooks";\n' + source.replace(MARKER, tail)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--out', type=Path, required=True)
    parser.add_argument('--adamic', type=Path, required=True)
    parser.add_argument('--cpu', type=int, default=min(os.sched_getaffinity(0)))
    parser.add_argument('--rounds', type=int, default=7)
    parser.add_argument('--case', action='append', help='select named manifest rows')
    parser.add_argument('--profile-only', action='store_true', help='profile existing results without repeating release timings')
    parser.add_argument('--profile', action='store_true', help='separate GNU gprof -pg builds; never timed as release')
    args = parser.parse_args()
    out = args.out.resolve()
    if out == ROOT or ROOT in out.parents:
        parser.error('run output must be outside the checkout')
    if args.cpu not in os.sched_getaffinity(0) or args.rounds < 5:
        parser.error('choose an allowed CPU and at least five rounds')
    out.mkdir(parents=True, exist_ok=True)
    if command(['node', '--version']).stdout.strip() != 'v24.19.0':
        raise RuntimeError('this comparison requires Node v24.19.0')
    manifest = (HERE / 'cases.json').read_bytes()
    cases = json.loads(manifest)
    metadata = dict(manifest_sha256=hashlib.sha256(manifest).hexdigest(), cpu=args.cpu, affinity=sorted(os.sched_getaffinity(0)),
                    instrument='performance.now(): CLOCK_MONOTONIC native, Node perf_hooks; milliseconds',
                    warmups=3, rounds=args.rounds, clang=command(['clang', '--version']).stdout,
                    node=command(['node', '--version']).stdout.strip(),
                    revision=command(['git', 'rev-parse', 'HEAD']).stdout.strip(),
                    load_before=Path('/proc/loadavg').read_text().strip(),
                    cpu_quota=Path('/sys/fs/cgroup/cpu.max').read_text().strip(),
                    cpu_model=[line.split(':', 1)[1].strip() for line in Path('/proc/cpuinfo').read_text().splitlines()
                               if line.startswith('model name')][0])
    rows = []
    if args.case:
        names = {case['name'] for case in cases}
        if set(args.case) - names:
            parser.error('unknown benchmark name')
        cases = [case for case in cases if case['name'] in args.case]
    if args.profile_only:
        previous = json.loads((out / 'results.json').read_text())
        metadata, rows, cases = previous['metadata'], previous['rows'], []
        args.profile = True

    def build(source, binary, env=None):
        result = command([args.adamic, 'build', source, '-o', binary], env=env)
        (out / (binary.name + '.build.log')).write_text(result.stdout + result.stderr)

    def trial(argv):
        result = command(['taskset', '-c', args.cpu, *argv], cwd=out)
        match = re.fullmatch(r'scout36 elapsed ([0-9.eE+-]+)\n', result.stderr)
        if not match or not math.isfinite(float(match[1])) or float(match[1]) <= 0:
            raise RuntimeError(f'invalid timed-region report: {result.stderr!r}')
        return dict(ms=float(match[1]), checksum=result.stdout)

    for case in cases:
        name = case['name']
        source = (HERE / (name + '.a')).read_text()
        if not case['lowers']:
            result = subprocess.run([str(args.adamic), 'build', str(HERE / (name + '.a')),
                                     '-o', str(out / name)], capture_output=True, text=True)
            if result.returncode != 1 or case['reason'] not in result.stderr:
                raise RuntimeError(f'expected refusal changed: {result.stderr}')
            (out / (name + '.refusal.log')).write_text(result.stderr)
            rows.append(dict(name=name, refused=case['reason']))
            continue
        path, binary = out / (name + '.a'), out / name
        count = 64
        while True:
            path.write_text(timed_source(source, count))
            build(path, binary)
            native = trial([binary])
            node = trial(['node', '--disable-warning=ExperimentalWarning', ROOT / 'oracle/node.mjs', path])
            if native['checksum'] != node['checksum']:
                raise RuntimeError(f'calibration disagrees: {name}')
            # Bound the slow runtime; aim for 50 ms on the faster one when feasible.
            fast, slow = min(native['ms'], node['ms']), max(native['ms'], node['ms'])
            if fast >= 50 or slow >= 500 or count >= 1_000_000:
                break
            factor = max(2, min(16, math.ceil(min(50 / fast, 600 / slow))))
            count = min(1_000_000, count * factor)
        samples = {'native': [], 'node': []}
        expected = native['checksum']
        for repeat in range(args.rounds):
            order = ['native', 'node'] if repeat % 2 == 0 else ['node', 'native']
            for backend in order:
                argv = [binary] if backend == 'native' else ['node', '--disable-warning=ExperimentalWarning', ROOT / 'oracle/node.mjs', path]
                result = trial(argv)
                if result['checksum'] != expected:
                    raise RuntimeError(f'timed output disagrees: {name}/{backend}')
                samples[backend].append(result['ms'])
        best = {key: min(values) for key, values in samples.items()}
        row = dict(name=name, iterations=count, source=case['source'],
                   source_sha256=hashlib.sha256(source.encode()).hexdigest(), samples_ms=samples,
                   best_ms=best, native_over_node=best['native'] / best['node'], checksum=expected)
        rows.append(row)
        print(f"{name}: native {best['native']:.3f} ms / Node {best['node']:.3f} ms = {row['native_over_node']:.3f}x", flush=True)
        (out / 'results.json').write_text(json.dumps(dict(metadata=metadata, rows=rows), indent=2) + '\n')
    metadata['load_after'] = Path('/proc/loadavg').read_text().strip()
    (out / 'results.json').write_text(json.dumps(dict(metadata=metadata, rows=rows), indent=2) + '\n')

    if args.profile:
        # A distinct compiler version string keeps the -pg runtime archive out of the release cache.
        wrapper = out / 'profile-tools'
        wrapper.mkdir(exist_ok=True)
        real_clang = shutil.which('clang')
        script = wrapper / 'clang'
        script.write_text('#!/usr/bin/env python3\nimport subprocess,sys\n'
                          f'compiler={real_clang!r}\n'
                          'if sys.argv[1:]==["--version"]:\n'
                          ' subprocess.run([compiler,"--version"],check=True)\n'
                          ' print("scout36 GNU gprof -pg instrumentation")\n'
                          'else:\n sys.exit(subprocess.call([compiler,"-pg",*sys.argv[1:]]))\n')
        script.chmod(0o755)
        env = dict(os.environ, PATH=str(wrapper) + os.pathsep + os.environ['PATH'])
        for row in rows:
            if row.get('native_over_node', 0) <= 1:
                continue
            name = row['name']
            directory = out / ('profile-' + name)
            directory.mkdir(exist_ok=True)
            binary = directory / name
            build(out / (name + '.a'), binary, env)
            result = command(['taskset', '-c', args.cpu, binary], cwd=directory)
            (directory / 'run.log').write_text(result.stdout + result.stderr)
            if result.stdout != row['checksum']:
                raise RuntimeError(f'profile output differs: {name}')
            report = command(['gprof', '-b', binary, directory / 'gmon.out'], cwd=directory)
            (directory / 'gprof.txt').write_text(report.stdout + report.stderr)
            print('profiled ' + name, flush=True)


if __name__ == '__main__':
    main()
