import sys,subprocess,time,json,os,platform
from pathlib import Path
root=Path('/workspace/lint-runtime-drivers')
d=Path(sys.argv[1]); mode=sys.argv[2]
manifest=['--manifest',str(d/'compiler.txt')]
commands={'Go':[str(d/'oracle')],'native':[str(d/'scanner')],'Node':['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(d/'main.ts')]}
def run(name,cmd):
 with (d/(name+'.stdout')).open('wb') as out,(d/(name+'.stderr')).open('wb') as err:
  start=time.perf_counter();subprocess.run(cmd,stdout=out,stderr=err,check=True);elapsed=time.perf_counter()-start
 return elapsed,(d/(name+'.stdout')).read_bytes()
if mode=='bench':
 samples={n:[] for n in commands};want=None;before=Path('/proc/loadavg').read_text().strip()
 for r in range(5):
  for n,c in commands.items():
   elapsed,answer=run(f'round{r}-{n}',c+manifest+['--count'])
   if want is None:want=answer
   assert answer==want,(n,answer,want)
   assert not (d/f'round{r}-{n}.stderr').stat().st_size
   samples[n].append(elapsed)
 _,answer=run('counted',[str(d/'counted')]+manifest+['--count']);assert answer==want
 result={'findings':int(want),'samples':samples,'best':{n:min(v) for n,v in samples.items()},'load_before':before,'load_after':Path('/proc/loadavg').read_text().strip(),'nproc':subprocess.check_output(['nproc']).decode().strip(),'cpu_max':Path('/sys/fs/cgroup/cpu.max').read_text().strip(),'machine':platform.uname()._asdict(),'counts':(d/'counted.stderr').read_text()}
 (d/'timing.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
elif mode=='parity':
 want=None
 for n,c in commands.items():
  elapsed,answer=run('full-'+n,c+manifest)
  if want is None:want=answer
  assert answer==want,n
  print(n,len(answer),elapsed)
 if (d/'sanitized').exists():
  elapsed,answer=run('full-sanitized',[str(d/'sanitized')]+manifest);assert answer==want;print('sanitized',len(answer),elapsed)
