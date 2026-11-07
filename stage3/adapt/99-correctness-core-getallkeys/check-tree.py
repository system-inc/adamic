"""Check the 79 roots, ordinary build, and full guarded core census."""
import json,os,pathlib,sys,time
from supervise import run,SCRATCH
side=sys.argv[1];tree=pathlib.Path('/workspace/getallkeys-'+side)
while not (tree/'patch-set.md').exists():time.sleep(1)
print(side+' full adaptation finished',flush=True)
roots=sorted(str(p) for p in (tree/'src/compiler').rglob('*.ts') if not p.name.endswith('.d.ts'))
assert len(roots)==79,len(roots)
status=run('types-'+side,[str(SCRATCH/'adamic'),'types',*roots])
if status['exit_code']:raise SystemExit(1)
(tree/'node_modules').symlink_to('/workspace/adamic/stage3/typescript/node_modules',target_is_directory=True)
status=run('build-'+side,['npm','run','build'],cwd=str(tree))
if status['exit_code']:raise SystemExit(1)
env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')}
env.update(LATENT_FULL='1',LATENT_ONLY_FILE=str(tree/'src/compiler/core.ts'),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(SCRATCH/('site-'+side+'-times.jsonl')))
status=run('site-'+side,[str(SCRATCH/'latent'),str(tree/'src/compiler'),str(SCRATCH/('site-'+side+'.jsonl'))],env=env,watch=[SCRATCH/('site-'+side+'.jsonl'),SCRATCH/('site-'+side+'-times.jsonl')])
if status['exit_code']:raise SystemExit(1)
