import sys,subprocess,os
from pathlib import Path
out=Path(__file__).resolve().parent;repo=out.parents[2]
def call(cmd):
 p=subprocess.run(cmd,cwd=repo);print('checkpoint command child_exit='+str(p.returncode),flush=True);p.check_returncode()
# Independent index creates append-only checkpoint commits without moving the final report branch.
env=dict(os.environ,GIT_INDEX_FILE='/workspace/scratch/require-row-checkpoint-index')
def git(args,capture=False):
 p=subprocess.run(['git',*args],cwd=repo,env=env,stdout=subprocess.PIPE if capture else None,text=True);print('checkpoint git child_exit='+str(p.returncode),flush=True);p.check_returncode();return p.stdout.strip() if capture else None
ref='refs/heads/typescript/require-row-census-checkpoint'
p=subprocess.run(['git','rev-parse','--verify',ref],cwd=repo,capture_output=True,text=True)
parent=p.stdout.strip() if p.returncode==0 else '71b126b06aa5840838330965f77b089fbffee3e4'
git(['read-tree',parent]);git(['add','--',str(out.relative_to(repo))]);tree=git(['write-tree'],True)
sha=git(['commit-tree',tree,'-p',parent,'-m','require row census checkpoint: '+sys.argv[1]],True)
git(['update-ref',ref,sha]);git(['push','origin',ref+':'+ref])
