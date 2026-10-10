#!/usr/bin/env python3
"""Measure the actual parallel_files items/work preparation using runtime snapshots."""
import argparse
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import tempfile
import time

parser = argparse.ArgumentParser()
parser.add_argument('--before-root', required=True)
parser.add_argument('--after-root', required=True)
parser.add_argument('--before-compiler', required=True)
parser.add_argument('--after-compiler', required=True)
parser.add_argument('--threads', default='1,2,4')
parser.add_argument('--rounds', type=int, default=5)
parser.add_argument('--output', required=True)
args = parser.parse_args()
roots = {version: Path(getattr(args, version + '_root')).resolve() for version in ('before', 'after')}
threads = [int(n) for n in args.threads.split(',')]
report = {'load_start': os.getloadavg(), 'samples': [], 'rounds': args.rounds}
footer = r'''
#include <time.h>
#include <stdio.h>
static double adamic_profile_started;
static double adamic_profile_now(void) {
 struct timespec value;
 clock_gettime(CLOCK_MONOTONIC, &value);
 return (double)value.tv_sec + (double)value.tv_nsec / 1e9;
}
void adamic_test_share_begin(void) { adamic_profile_started = adamic_profile_now(); }
void adamic_test_share_end(void) {
 fprintf(stderr, "items/work share seconds %.9f\n", adamic_profile_now() - adamic_profile_started);
}
'''
with tempfile.TemporaryDirectory(prefix='adamic-share-profile-') as temporary:
    temporary = Path(temporary)
    binaries = {}
    for version in ('before', 'after'):
        root = roots[version]
        snapshot = temporary / (version + '-runtime')
        shutil.copytree(root / 'internal/native/runtime', snapshot)
        parallel = snapshot / 'parallel.c'
        source = parallel.read_text()
        seam = 'adamic_share(items);\n\tadamic_share(work);'
        assert source.count(seam) == 1
        source = source.replace(seam, 'extern void adamic_test_share_begin(void); extern void adamic_test_share_end(void);\n\tadamic_test_share_begin();\n\t' + seam + '\n\tadamic_test_share_end();')
        parallel.write_text(source)
        generated = subprocess.check_output([getattr(args, version + '_compiler'), 'c', str(root / 'bench/parallel_files.a')], cwd=root)
        program = temporary / (version + '.c')
        program.write_text('#define _POSIX_C_SOURCE 200809L\n' + generated.decode() + footer)
        binary = temporary / version
        subprocess.run(['go', 'run', './internal/native/testdata/parallel/build_measure.go', '-source', str(program), '-runtime', str(snapshot), '-o', str(binary)], cwd=roots['after'], check=True)
        binaries[version] = binary
    report['load_measurement_start'] = os.getloadavg()
    expected = None
    for iteration in range(args.rounds):
        for n in threads:
            versions = ('before', 'after') if iteration % 2 == 0 else ('after', 'before')
            for version in versions:
                start = time.perf_counter()
                run = subprocess.run([binaries[version]], env=dict(os.environ, ADAMIC_THREADS=str(n)), capture_output=True, check=True)
                wall = time.perf_counter() - start
                if expected is None:
                    expected = run.stdout
                assert run.stdout == expected
                measurements = re.findall(rb'items/work share seconds ([0-9.]+)', run.stderr)
                assert len(measurements) == 1, run.stderr
                sample = {'round': iteration + 1, 'version': version, 'threads': n, 'share': float(measurements[0]), 'wall': wall}
                report['samples'].append(sample)
                print(json.dumps(sample), flush=True)
report['load_end'] = os.getloadavg()
report['best'] = []
for n in threads:
    entry = {'threads': n}
    for version in ('before', 'after'):
        entry[version] = min(sample['share'] for sample in report['samples'] if sample['threads'] == n and sample['version'] == version)
    report['best'].append(entry)
Path(args.output).write_text(json.dumps(report, indent=2) + '\n')
print(json.dumps({'load_start': report['load_measurement_start'], 'load_end': report['load_end'], 'best': report['best']}, indent=2))
