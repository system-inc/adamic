"""Restore only core's original whole owner in an isolated accepted source copy."""
import os,pathlib,shutil
from supervise import run,SCRATCH
before=pathlib.Path('/workspace/getallkeys-before');after=pathlib.Path('/workspace/getallkeys-after');mutant=SCRATCH/'mutant'
shutil.copytree(after/'src',mutant/'src')
shutil.copyfile(before/'src/compiler/core.ts',mutant/'src/compiler/core.ts')
env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')}
env.update(LATENT_FULL='1',LATENT_ONLY_FILE=str(mutant/'src/compiler/core.ts'),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(SCRATCH/'site-mutant-times.jsonl'))
status=run('site-mutant',[str(SCRATCH/'latent'),str(mutant/'src/compiler'),str(SCRATCH/'site-mutant.jsonl')],env=env,watch=[SCRATCH/'site-mutant.jsonl',SCRATCH/'site-mutant-times.jsonl'])
if status['exit_code']:raise SystemExit(1)
