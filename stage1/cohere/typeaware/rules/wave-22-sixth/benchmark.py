#!/usr/bin/env python3
"""Three alternating full-output runs; compare bytes before recording medians."""
import json, pathlib, statistics, subprocess, time
ROOT=pathlib.Path(__file__).resolve().parents[5]
WORK=pathlib.Path('/workspace/wave-22-sixth-work')
results={}
for corpus,config in [('repository',ROOT/'tsconfig.json'),('compiler',pathlib.Path('/workspace/wave-22-typescript-pinned/src/compiler/tsconfig.json'))]:
    samples={'native':[],'go':[]}
    for round in range(3):
        outputs=[]
        for implementation,binary in [('go',WORK/'oracle'),('native',WORK/'native')]:
            label=f'benchmark-{corpus}-{round}-{implementation}';before=time.perf_counter()
            with (WORK/(label+'.stdout')).open('wb') as out,(WORK/(label+'.stderr')).open('wb') as err:
                process=subprocess.run([str(binary),str(config),str(WORK/(corpus+'.manifest'))],cwd=ROOT,stdout=out,stderr=err)
            elapsed=time.perf_counter()-before
            assert process.returncode==0
            samples[implementation].append(elapsed);outputs.append((WORK/(label+'.stdout')).read_bytes())
            if implementation=='native':assert not (WORK/(label+'.stderr')).read_bytes()
        assert outputs[0]==outputs[1]
    result={'seconds':samples,'median_seconds':{name:statistics.median(values) for name,values in samples.items()}}
    result['native_over_go']=result['median_seconds']['native']/result['median_seconds']['go'];results[corpus]=result
    print(corpus,json.dumps(result),flush=True)
(WORK/'timings.json').write_text(json.dumps(results,indent=2)+'\n')
