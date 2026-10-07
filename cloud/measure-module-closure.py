#!/usr/bin/env python3
"""Interleave fresh module caches; retain exact logs, flags and archive counts."""
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import time

REPO = Path(__file__).resolve().parent.parent
SCRATCH = Path('/tmp/adamic-gate/closure-proof')
REPORT = REPO / 'cloud/reports/setup-closure'
BEFORE = SCRATCH / 'before.py'
AFTER = REPO / 'cloud/setup-modules.py'


def command(args, env):
    return subprocess.check_output(args, cwd=REPO, env=env, timeout=30).decode().strip()


def remove_cache(root):
    if not root.resolve().is_relative_to(SCRATCH):
        raise ValueError('scratch path required')
    for directory, _, _ in os.walk(root):
        Path(directory).chmod(0o755)
    shutil.rmtree(root)



def main():
    REPORT.mkdir(parents=True, exist_ok=True)
    measurements = json.loads((REPORT / 'measurements.json').read_text()) if (REPORT / 'measurements.json').exists() else []
    SCRATCH.mkdir(parents=True, exist_ok=True)
    if not BEFORE.exists():
        BEFORE.write_bytes(subprocess.check_output(['git', 'show', '5794c87661b030d24cee457bbec4900f260e025d:cloud/setup-modules.py'], cwd=REPO))
    repeat = 3
    for loop in range(1, 4):
        for kind, helper in [('before', BEFORE), ('after', AFTER)]:
            root = SCRATCH / f'{kind}-{loop}'
            if any(item['loop'] == loop and item['mode'] == kind for item in measurements):
                if (root / 'cache').exists() and not (kind == 'after' and loop == 3):
                    remove_cache(root / 'cache')
                continue
            root.mkdir(exist_ok=True)
            assert not (root / 'cache').exists(), 'fresh module cache required' 
            env = dict(os.environ, GOMODCACHE=str(root / 'cache'))
            env.pop('ADAMIC_GATE_UNCACHED', None)
            flags = dict(commit=command(['git', 'rev-parse', 'HEAD'], env), nproc=command(['nproc'], env),
                         cpu_max=Path('/sys/fs/cgroup/cpu.max').read_text().strip(),
                         go=command(['go', 'version'], env), clang=command(['clang', '--version'], env).splitlines()[0],
                         node=command(['node', '--version'], env), cached=False,
                         load_before=Path('/proc/loadavg').read_text().strip())
            args = [sys.executable, str(helper), str(REPO), str(root / 'tools')]
            start = time.monotonic()
            with (REPORT / f'{kind}-{loop}.log').open('wb') as output:
                result = subprocess.run(args, cwd=REPO, env=env, stdout=output, stderr=subprocess.STDOUT, timeout=600)
            elapsed = time.monotonic() - start
            flags['load_after'] = Path('/proc/loadavg').read_text().strip()
            zips = sorted(str(p.relative_to(root / 'cache')) for p in (root / 'cache').rglob('*.zip'))
            measurement = dict(loop=loop, mode=kind, seconds=elapsed, exit=result.returncode,
                               count=len(zips), archives=zips, instrument=args, build_flags=flags)
            measurements.append(measurement)
            (REPORT / 'measurements.json').write_text(json.dumps(measurements, indent=2) + '\n')
            print(kind, loop, f'{elapsed:.3f}s', len(zips), 'archives, exit', result.returncode, flush=True)
            if result.returncode:
                raise RuntimeError('measurement failed; see log')
            if elapsed > 300:
                repeat = 1
            if kind == 'after' and (loop == 3 or repeat == 1):
                (REPORT / 'fresh-cache-path').write_text(str(root / 'cache') + '\n')
                stamps = list((root / 'tools').glob('modules-closure-*'))
                selected = json.loads(stamps[0].read_text())['modules']
                actual = set(zips)
                expected = set()
                for name in selected:
                    path, version = name.rsplit('@', 1)
                    escaped = ''.join('!' + c.lower() if c.isupper() else c for c in path)
                    expected.add('cache/download/' + escaped + '/@v/' + version + '.zip')
                assert actual == expected, (actual - expected, expected - actual)
                print('PASS: fresh cache contains exactly selected closure archives', flush=True)
            else:
                remove_cache(root / 'cache')
        if repeat == 1:
            break


if __name__ == '__main__':
    main()
