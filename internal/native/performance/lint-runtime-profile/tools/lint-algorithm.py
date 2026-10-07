import sys,subprocess,time,json,os
from pathlib import Path
binary=Path(sys.argv[1]);d=Path('/workspace/lint-algorithm');d.mkdir(exist_ok=True)
commands={'Go':[str(binary/'oracle')],'native':[str(binary/'scanner')],'Node':['node','--disable-warning=ExperimentalWarning','/workspace/lint-runtime-drivers/oracle/node.mjs',str(binary/'main.ts')]}
results=[]
for size in (100,200,400,800):
 source=d/f'private_{size}.a';source.write_text('class C {\n'+''.join(f'#p{i}=0;\n' for i in range(size))+''.join(f'm{i}(){{return this.#p{i};}}\n' for i in range(size))+'}\n')
 for rule in ('react/no-find-dom-node','no-unused-private-class-members'):
  slug=rule.replace('/','_')
  manifest=d/f'{size}-{slug}.txt';manifest.write_text(str(source)+'\t'+rule+'\n');samples={name:[] for name in commands}
  for trial in range(5):
   for name,command in commands.items():
    label=f'{size}-{slug}-{trial}-{name}'
    with (d/(label+'.stdout')).open('wb') as out,(d/(label+'.stderr')).open('wb') as err:
     start=time.perf_counter();r=subprocess.run(command+['--manifest',str(manifest),'--count'],stdout=out,stderr=err,check=True);elapsed=time.perf_counter()-start
    assert (d/(label+'.stdout')).read_bytes()==b'0\n',label
    assert not (d/(label+'.stderr')).stat().st_size,label
    samples[name].append(elapsed)
  results.append({'private_names':size,'rule':rule,'samples':samples,'best':{name:min(times) for name,times in samples.items()}})
  (d/'results.json').write_text(json.dumps({'load':Path('/proc/loadavg').read_text().strip(),'driver':str(binary),'results':results},indent=2)+'\n')
  print(size,rule,results[-1]['best'],flush=True)
