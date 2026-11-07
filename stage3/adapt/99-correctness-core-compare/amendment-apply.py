import os,sys
from pathlib import Path
from concurrent.futures import ThreadPoolExecutor
from child_runner import run,ChildExit
s=Path('/workspace/scratch/core-compare-amend');repo=Path('/workspace/adamic');base=s/'main'
env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE='/workspace/typescript-source',PYTHONDONTWRITEBYTECODE='1')
def main():
 run('worktree',['git','worktree','add','--detach',base,'f0271c6a7b2e92900eb7046d8246cfb89f52cfb6'],env=env)
 (base/'stage3/api/node_modules').symlink_to(repo/'stage3/api/node_modules',target_is_directory=True)
 def apply(label,checkout):return run('apply-'+label,['bash',checkout/'stage3/apply.sh','--build-product',s/label],cwd=checkout,env=env)
 with ThreadPoolExecutor(max_workers=2) as pool:
  jobs=[pool.submit(apply,'before',base),pool.submit(apply,'after',repo)]
  errors=[]
  for job in jobs:
   try:job.result()
   except ChildExit as e:errors.append(e)
  if errors:raise errors[0]
 print('Both full applies: childExitCode=0',flush=True)
try:main()
except ChildExit as e:
 print(str(e),flush=True);sys.exit(e.code if e.code>=0 else 128-e.code)
