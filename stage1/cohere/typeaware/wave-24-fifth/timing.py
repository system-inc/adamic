#!/usr/bin/env python3
"""Alternating full-process counts; logs never use a shell pipe."""
import argparse,json,re,statistics,subprocess,time
from pathlib import Path
p=argparse.ArgumentParser(description=__doc__)
p.add_argument('artifacts',type=Path)
p.add_argument('--repository',type=Path,required=True)
p.add_argument('--compiler',type=Path,required=True)
a=p.parse_args();d=a.artifacts.resolve();result={}
for corpus,config in [('repository',a.repository/'tsconfig.json'),('compiler',a.compiler/'src/compiler/tsconfig.json')]:
    samples={'native':[],'go':[]};truth=None
    for repeat in range(3):
        for label,executable in [('native','native'),('go','oracle')]:
            stem=d/f'timing-{corpus}-{label}-{repeat}'
            with stem.with_suffix('.stdout').open('wb') as out,stem.with_suffix('.stderr').open('wb') as err:
                start=time.perf_counter_ns()
                run=subprocess.run([str(d/executable),str(config),str(d/f'{corpus}.manifest'),'--count'],stdout=out,stderr=err)
                elapsed=time.perf_counter_ns()-start
            assert run.returncode==0
            stderr=stem.with_suffix('.stderr').read_bytes()
            assert not stderr or (label=='go' and re.fullmatch(rb'cohere: load_ns=\d+ rule_ns=\d+ run_ns=\d+\n',stderr))
            count=stem.with_suffix('.stdout').read_bytes()
            if truth is None:truth=count
            assert count==truth,'count mismatch'
            samples[label].append(elapsed)
    median={label:statistics.median(values) for label,values in samples.items()}
    result[corpus]={'samples_ns':samples,'median_ns':median,'native_over_go':median['native']/median['go']}
(d/'timings.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result))
