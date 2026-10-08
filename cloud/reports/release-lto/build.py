#!/usr/bin/env python3
"""Build unchanged generated C with production release flags and separate runtime archives."""
import json
from pathlib import Path
import shlex
import subprocess
import sys
import time

repository = Path.cwd()
scratch = Path(sys.argv[1]).resolve()
compiler = '/workspace/adamic-tools/llvm/bin/clang'
archiver = '/workspace/adamic-tools/llvm/bin/llvm-ar'
flags = ['-std=c11', '-Wall', '-Wextra', '-Werror', '-pedantic', '-Wno-unused-variable', '-Wno-unused-but-set-variable', '-Wno-unused-function', '-Wno-unused-parameter', '-Wno-self-assign', '-ffp-contract=off', '-fno-optimize-sibling-calls', '-O2']
results = {}
with (scratch / 'commands.log').open('w') as commands:
    def run(arguments, log):
        commands.write(shlex.join(map(str, arguments)) + '\n')
        commands.flush()
        start = time.perf_counter()
        with log.open('ab') as output:
            subprocess.run(list(map(str, arguments)), stdout=output, stderr=output, check=True)
        return time.perf_counter() - start
    for mode in ['baseline', 'thin']:
        directory = scratch / mode
        directory.mkdir(exist_ok=True)
        compile_flags = flags + (['-flto=thin'] if mode == 'thin' else [])
        start = time.perf_counter()
        objects = []
        for source in sorted((repository / 'internal/native/runtime').glob('*.c')):
            obj = directory / (source.stem + '.o')
            run([compiler, *compile_flags, '-c', source, '-o', obj], directory / 'runtime-build.log')
            objects.append(obj)
        library = directory / 'runtime.a'
        run([archiver, 'rcs', library, *objects], directory / 'runtime-build.log')
        runtime_seconds = time.perf_counter() - start
        results[mode] = {'runtime_seconds': runtime_seconds, 'archive_bytes': library.stat().st_size, 'workloads': {}}
        for workload in ['parse', 'service']:
            arguments = [compiler, *compile_flags, *(['-fuse-ld=lld'] if mode == 'thin' else []), '-I', repository / 'internal/native/runtime', '-o', directory / workload, scratch / (workload + '.c'), '-Xlinker', '--whole-archive', library, '-Xlinker', '--no-whole-archive', '-lm']
            first = run(arguments, directory / (workload + '-build.log'))
            cached = run(arguments, directory / (workload + '-build.log'))
            results[mode]['workloads'][workload] = {'cold_seconds': runtime_seconds + first, 'first_link_seconds': first, 'cached_seconds': cached, 'binary_bytes': (directory / workload).stat().st_size}
        (scratch / 'build-results.json').write_text(json.dumps(results, indent=2) + '\n')
print(json.dumps(results, indent=2))
