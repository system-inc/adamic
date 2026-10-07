"""Quiet alternating process measurements; complete diagnostics compared every round."""
import pathlib,subprocess,statistics,time,json
ROOT=pathlib.Path(__file__).resolve().parents[4];S=pathlib.Path('/workspace/wave-26-attributes');records=[]
for population,config,manifest in [('compiler','/workspace/wave-26-typescript/src/compiler/tsconfig.json','/workspace/wave-26-compiler.manifest'),('repository',str(ROOT/'tsconfig.json'),'/workspace/wave-26-repository.manifest')]:
    expected=(S/(population+'-go.stdout')).read_bytes();samples={'native':[],'go':[]}
    for round in range(3):
        for engine in (['native','go'] if round%2 else ['go','native']):
            name='cost-'+population+'-'+str(round)+'-'+engine;command=[str(S/('oracle' if engine=='go' else 'native')),config,manifest];started=time.monotonic()
            with (S/(name+'.stdout')).open('wb') as out,(S/(name+'.stderr')).open('wb') as err:result=subprocess.run(command,cwd=ROOT,stdout=out,stderr=err)
            seconds=time.monotonic()-started;assert result.returncode==0 and (S/(name+'.stdout')).read_bytes()==expected
            if engine=='native':assert (S/(name+'.stderr')).read_bytes()==b''
            samples[engine].append(seconds);records.append(dict(name=name,command=command,exit=result.returncode,seconds=seconds));(S/'cost-runs.json').write_text(json.dumps(records,indent=2)+'\n')
    print(population,'native',statistics.median(samples['native']),'Go',statistics.median(samples['go']),flush=True)
