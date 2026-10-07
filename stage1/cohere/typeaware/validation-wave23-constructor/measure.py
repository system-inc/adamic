#!/usr/bin/env python3
"""Three alternating whole-process runs, always checking complete diagnostic bytes."""
import hashlib
import json
import os
from pathlib import Path
import statistics
import subprocess
import sys
import time

binary, oracle, corpus, artifacts, repository = map(Path, sys.argv[1:])
rounds = []
for name, config, manifest in [
    ('compiler', corpus / 'src/compiler/tsconfig.json', artifacts.parent / 'compiler.manifest'),
    ('repository', repository / 'tsconfig.json', artifacts.parent / 'repository.manifest'),
]:
    for iteration in range(3):
        outputs = {}
        row = {'corpus': name, 'round': iteration + 1}
        for implementation in (['native', 'go'] if iteration % 2 == 0 else ['go', 'native']):
            executable = binary if implementation == 'native' else oracle
            stem = artifacts / f'{name}-{iteration + 1}-{implementation}'
            environment = os.environ.copy()
            if implementation == 'native':
                environment['ADAMIC_TSGO_TIMING'] = '1'
            with stem.with_suffix('.stdout').open('wb') as stdout, stem.with_suffix('.stderr').open('wb') as stderr:
                started = time.perf_counter_ns()
                result = subprocess.run([str(executable), str(config), str(manifest)], stdout=stdout, stderr=stderr, env=environment)
                row[implementation + '_ns'] = time.perf_counter_ns() - started
            if result.returncode:
                raise RuntimeError(f'{implementation} exited {result.returncode}: {stem}')
            outputs[implementation] = stem.with_suffix('.stdout').read_bytes()
            row[implementation + '_timing'] = stem.with_suffix('.stderr').read_text().strip()
        if outputs['native'] != outputs['go']:
            raise RuntimeError(f'{name} round {iteration + 1}: diagnostic bytes differ')
        row['bytes'] = len(outputs['native'])
        row['sha256'] = hashlib.sha256(outputs['native']).hexdigest()
        rounds.append(row)
summary = {}
for name in ['compiler', 'repository']:
    selected = [row for row in rounds if row['corpus'] == name]
    summary[name] = {implementation: statistics.median(row[implementation + '_ns'] for row in selected) / 1e9 for implementation in ['native', 'go']}
    summary[name]['native_over_go'] = summary[name]['native'] / summary[name]['go']
(artifacts / 'measurements.json').write_text(json.dumps({'rounds': rounds, 'medians_seconds': summary}, indent=2) + '\n')
print(json.dumps(summary, indent=2))
