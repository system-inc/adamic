#!/usr/bin/env python3
"""Time production runtime/cache/link work without Go frontend build overhead."""
import json
import os
from pathlib import Path
import resource
import subprocess
import sys
import time

s = Path(sys.argv[1]).resolve()
results = {}
for workload in ['parse', 'service']:
    results[workload] = {}
    for mode in ['baseline', 'thin']:
        cache = s / ('build-cache-' + workload + '-' + mode)
        if cache.exists():
            raise RuntimeError('cold cache must not exist: ' + str(cache))
        env = dict(os.environ, XDG_CACHE_HOME=str(cache))
        results[workload][mode] = {}
        for state in ['cold', 'cached']:
            stem = s / ('build-' + workload + '-' + mode + '-' + state)
            binary = str(stem) + '.bin'
            command = [str(s / 'build-shipped')]
            if mode == 'thin':
                command.append('-release')
            command += [str(s / (workload + '.c')), binary]
            before = resource.getrusage(resource.RUSAGE_CHILDREN)
            load = os.getloadavg()
            start = time.monotonic()
            with Path(str(stem) + '.json').open('wb') as out, Path(str(stem) + '.stderr').open('wb') as err:
                subprocess.run(command, stdout=out, stderr=err, env=env, check=True)
            wall = time.monotonic() - start
            after = resource.getrusage(resource.RUSAGE_CHILDREN)
            results[workload][mode][state] = dict(wall=wall, user=after.ru_utime-before.ru_utime, system=after.ru_stime-before.ru_stime, bytes=Path(binary).stat().st_size, load_before=load, load_after=os.getloadavg(), command=command)
            print(workload, mode, state, json.dumps(results[workload][mode][state]), flush=True)
            (s / 'build-times.json').write_text(json.dumps(results, indent=2) + '\n')
