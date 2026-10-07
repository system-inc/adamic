#!/usr/bin/env python3
"""Replay logged clang commands with a fresh archive for each workload independently."""
import json
import os
import resource
from pathlib import Path
import shlex
import subprocess
import sys
import time

scratch = Path(sys.argv[1]).resolve()
commands = [shlex.split(line) for line in (scratch / 'commands.log').read_text().splitlines()]
tag = sys.argv[2] if len(sys.argv) > 2 else "cold"
results = {}
with (scratch / (tag + '-commands.log')).open('w') as command_log:
    for mode in ['baseline', 'thin']:
        results[mode] = {}
        original = scratch / mode
        for workload in ['parse', 'service']:
            directory = scratch / (tag + '-' + mode + '-' + workload)
            directory.mkdir()  # Existing output would invalidate the cold-cache claim.
            runtime = [cmd for cmd in commands if '-c' in cmd and cmd[-1].startswith(str(original) + '/')]
            archive = next(cmd for cmd in commands if 'rcs' in cmd and cmd[2] == str(original / 'runtime.a'))
            link = next(cmd for cmd in commands if '-o' in cmd and cmd[cmd.index('-o') + 1] == str(original / workload))
            def run(cmd):
                replay = [arg.replace(str(original) + '/', str(directory) + '/') for arg in cmd]
                command_log.write(shlex.join(replay) + '\n'); command_log.flush()
                with (directory / 'build.log').open('ab') as out:
                    subprocess.run(replay, stdout=out, stderr=out, check=True)
            before = os.getloadavg()
            usage_start = resource.getrusage(resource.RUSAGE_CHILDREN)
            start = time.perf_counter()
            for cmd in runtime:
                run(cmd)
            run(archive)
            runtime_seconds = time.perf_counter() - start
            run(link)
            cold = time.perf_counter() - start
            usage_cold = resource.getrusage(resource.RUSAGE_CHILDREN)
            start = time.perf_counter()
            run(link)
            cached = time.perf_counter() - start
            usage_cached = resource.getrusage(resource.RUSAGE_CHILDREN)
            results[mode][workload] = {'cold_user_seconds': usage_cold.ru_utime - usage_start.ru_utime, 'cold_system_seconds': usage_cold.ru_stime - usage_start.ru_stime, 'cached_user_seconds': usage_cached.ru_utime - usage_cold.ru_utime, 'cached_system_seconds': usage_cached.ru_stime - usage_cold.ru_stime, 'cold_seconds': cold, 'runtime_seconds': runtime_seconds, 'cached_seconds': cached, 'load_before': before, 'load_after': os.getloadavg(), 'binary_bytes': (directory / workload).stat().st_size}
            print(mode, workload, json.dumps(results[mode][workload]), flush=True)
            (scratch / (tag + '-build-results.json')).write_text(json.dumps(results, indent=2) + '\n')
