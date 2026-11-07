import os,time,subprocess,json
from pathlib import Path
s=Path('/workspace/scratch/core-extend')
while not all((s/n/'patch-set.md').exists() for n in ['before','after']):
 print('waiting for complete applies',flush=True);time.sleep(30)
commands=[('source-comparison',['node',str(s/'compare-source.cjs')]),('callers-adapted',['node',str(s/'callers.cjs'),str(s/'before')]),('node-witnesses',['node',str(s/'witness.cjs')]),('provenance',['python3',str(s/'provenance.py')])]
for name,args in commands:
 with (s/(name+'.json')).open('w') as log:r=subprocess.run(args,stdout=log,stderr=subprocess.STDOUT,timeout=1000)
 print(name,r.returncode,flush=True)
 if r.returncode:raise SystemExit(r.returncode)
