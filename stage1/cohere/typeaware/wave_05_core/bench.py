"""Isolated three-round aggregate native/Go comparison; logs go directly to files."""
from pathlib import Path
import json,os,statistics,subprocess,time
ROOT=Path(__file__).resolve().parents[4]
OUT=Path('/workspace/wave-05-core-benchmark');OUT.mkdir(exist_ok=True)
BIN=Path('/workspace/wave-05-core-validation')
results={}
for name,config,manifest in [('repository',ROOT/'tsconfig.json',Path('/workspace/wave-05-repository.manifest')),('compiler',Path('/workspace/wave-05-typescript/src/compiler/tsconfig.json'),Path('/workspace/wave-05-compiler.manifest'))]:
    rounds={'go':[],'native':[]}
    for at in range(3):
        outputs=[]
        for label in (['go','native'] if at%2==0 else ['native','go']):
            exe=BIN/('oracle' if label=='go' else 'native')
            prefix=OUT/f'{name}-{label}-{at}'
            begin=time.monotonic()
            with Path(str(prefix)+'.stdout').open('wb') as stdout,Path(str(prefix)+'.stderr').open('wb') as stderr:
                code=subprocess.run([str(exe),str(config),str(manifest),'--count'],cwd=ROOT,env=dict(os.environ,ADAMIC_TSGO_TIMING='1'),stdout=stdout,stderr=stderr).returncode
            elapsed=time.monotonic()-begin;assert code==0
            outputs.append(Path(str(prefix)+'.stdout').read_bytes());rounds[label].append(elapsed)
            print(name,label,at,elapsed,outputs[-1].strip().decode(),Path(str(prefix)+'.stderr').read_text().strip(),flush=True)
        assert outputs[0]==outputs[1]
    results[name]={'rounds':rounds,'medians':{label:statistics.median(values) for label,values in rounds.items()}}
    print(name,'median',results[name]['medians'],flush=True)
(OUT/'timings.json').write_text(json.dumps(results,indent=2)+'\n')
print('PASS timings, identical counts each round')
