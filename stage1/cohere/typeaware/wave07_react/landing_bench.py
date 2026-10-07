#!/usr/bin/env python3
"""Alternate complete native/Go outputs after the landing gates finish."""
import hashlib
import json
import os
import pathlib
import statistics
import subprocess
import time

ROOT = pathlib.Path(__file__).resolve().parents[4]
LANDING = pathlib.Path(os.environ.get('ADAMIC_WAVE07_LANDING', '/workspace/wave-07-landing'))
OUT = LANDING / 'bench'
OUT.mkdir(parents=True, exist_ok=True)
records = []
units = [('original-trio', LANDING / 'original-trio',
          'wave07', 'wave07-oracle')]
for label, directory in [('timer','/workspace/wave-07-timer'),
    ('process','/workspace/wave-07-process'),('streams','/workspace/wave-07-streams-final'),
    ('rest','/workspace/wave-07-next-rest'),('promise','/workspace/wave-07-next-promise'),
    ('regex','/workspace/wave-07-next-regex')]:
    units.append((label,pathlib.Path(directory),'timer','oracle'))
for label, directory, native_name, oracle_name in units:
    for corpus, config in [('repository',ROOT/'tsconfig.json'),
        ('compiler',pathlib.Path('/workspace/wave-07-typescript/src/compiler/tsconfig.json'))]:
        manifest = (pathlib.Path('/workspace/wave-07-artifacts') if label == 'original-trio'
                    else directory) / (corpus+'.manifest')
        for iteration in range(3):
            order = [('go',directory/oracle_name),('native',directory/native_name)]
            if iteration % 2:
                order.reverse()
            expected = None
            for implementation, binary in order:
                stem = OUT / (label+'-'+corpus+'-'+str(iteration)+'-'+implementation)
                command = [str(binary),str(config),str(manifest)]
                with stem.with_suffix('.stdout').open('wb') as stdout, stem.with_suffix('.stderr').open('wb') as stderr:
                    started = time.perf_counter_ns()
                    result = subprocess.run(command,cwd=ROOT,stdout=stdout,stderr=stderr,
                                            env=dict(os.environ,ADAMIC_TSGO_TIMING='1'))
                    elapsed = time.perf_counter_ns()-started
                assert result.returncode == 0, (label,implementation,result.returncode)
                output = stem.with_suffix('.stdout').read_bytes()
                assert expected is None or output == expected, (label,corpus,'byte mismatch')
                expected = output
                records.append(dict(unit=label,corpus=corpus,round=iteration,
                    implementation=implementation,command=command,exit=result.returncode,
                    elapsed_ns=elapsed,stdout_sha256=hashlib.sha256(output).hexdigest()))
                (OUT/'records.json').write_text(json.dumps(records,indent=2)+'\n')
        values = {implementation:statistics.median(record['elapsed_ns']/1e9 for record in records
            if record['unit']==label and record['corpus']==corpus and record['implementation']==implementation)
            for implementation in ['native','go']}
        print(label+' '+corpus+': native '+format(values['native'],'.6f')+
              's / Go '+format(values['go'],'.6f')+'s',flush=True)
print('PASS all alternating output comparisons',flush=True)
