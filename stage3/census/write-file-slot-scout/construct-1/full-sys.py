from pathlib import Path
import subprocess,os,time,json
r=Path('/workspace/scratch/write-file-construct-run')
while not (r/'census').exists():time.sleep(3)
for name,tree in [('control',r/'control-adapted'),('candidate',r/'lane/adapted-tree')]:
 if name=='control' and (r/'full-sys-control.exit').read_text().strip()=='0':continue
 if name=='candidate':
  while not (r/'lane/patch-set.md').exists() and not (r/'lane.exit').exists():time.sleep(5)
  assert (r/'lane/patch-set.md').exists()
 env=dict(os.environ,TMPDIR=str(r/'tmp'),GOTMPDIR=str(r/'tmp'),GOMEMLIMIT='3GiB',LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',CENSUS_MODULE_FORMAT='CommonJSBundle',LATENT_ONLY_FILE=str(tree/'src/compiler/sys.ts'),LATENT_TIMES=str(r/f'full-sys-{name}.times.jsonl'))
 for key in ['LATENT_SPECULATIVE','LATENT_RUN_ONLY','LATENT_ROOT_MANIFEST']:env.pop(key,None)
 with (r/f'full-sys-{name}.log').open('w') as log:p=subprocess.run([str(r/'census'),str(tree/'src/compiler'),str(r/f'full-sys-{name}.jsonl')],env=env,stdout=log,stderr=subprocess.STDOUT)
 (r/f'full-sys-{name}.exit').write_text(str(p.returncode)+'\n');assert p.returncode==0
