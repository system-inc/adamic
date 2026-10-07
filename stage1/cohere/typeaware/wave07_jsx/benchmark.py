#!/usr/bin/env python3
"""Quiet per-rule process timings with complete Go byte comparison."""
import json
import pathlib
import statistics
import subprocess
import time
OUT = pathlib.Path('/workspace/wave-07-jsx')
ROOT = pathlib.Path(__file__).resolve().parents[4]
results = []
for corpus, config in [('repository', ROOT/'tsconfig.json'), ('compiler', pathlib.Path('/workspace/wave-07-typescript/src/compiler/tsconfig.json'))]:
    for rule in ['fragments', 'undef', 'adjacent']:
        samples = {'native': [], 'go': []}
        for iteration in range(3):
            outputs = []
            for language, executable in [('native', 'native-suite'), ('go', 'oracle')]:
                stem = OUT/f'timing-{corpus}-{rule}-{language}-{iteration}'
                started = time.monotonic_ns()
                with stem.with_suffix('.stdout').open('wb') as stdout, stem.with_suffix('.stderr').open('wb') as stderr:
                    run = subprocess.run([str(OUT/executable), str(config), str(OUT/(corpus+'.manifest')), rule], stdout=stdout, stderr=stderr)
                samples[language].append((time.monotonic_ns()-started)/1e9)
                assert run.returncode == 0
                if language == 'native': assert not stem.with_suffix('.stderr').read_bytes()
                outputs.append(stem.with_suffix('.stdout').read_bytes())
            assert outputs[0] == outputs[1]
        results.append(dict(corpus=corpus, rule=rule, samples=samples, median={key:statistics.median(value) for key,value in samples.items()}, bytes=len(outputs[0])))
(OUT/'timings.json').write_text(json.dumps(results, indent=2)+'\n')
for row in results: print(row['corpus'], row['rule'], row['median'], row['bytes'])
