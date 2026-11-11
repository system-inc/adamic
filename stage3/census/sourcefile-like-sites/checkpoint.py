import fcntl
_checkpoint_lock=open("/tmp/baobab-checkpoint.lock","a")
fcntl.flock(_checkpoint_lock,fcntl.LOCK_EX)
import pathlib,subprocess,sys,time
out=pathlib.Path(__file__).resolve().parent
repo=out.parents[2]
def run(args):
 for attempt in range(3):
  p=subprocess.run(args,cwd=repo,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,text=True)
  print('child exit='+str(p.returncode)+' attempt='+str(attempt+1)+' '+repr(args)+'\n'+p.stdout,flush=True)
  if p.returncode==0:return p.stdout.strip()
  transient=args[:2]==['git','push'] and any(token in p.stdout for token in ['HTTP 503','error: 503','HTTP 502','error: 502','HTTP 504','error: 504'])
  if not transient or attempt==2:sys.exit(p.returncode)
  time.sleep(attempt+1)
run(['git','add','-f','stage3/census/sourcefile-like-sites'])
branch='refs/heads/typescript/sourcefile-like-sites-checkpoint'
p=subprocess.run(['git','rev-parse','--verify',branch],cwd=repo,capture_output=True,text=True)
parent=p.stdout.strip() if p.returncode==0 else run(['git','rev-parse','HEAD'])
tree=run(['git','write-tree'])
commit=run(['git','commit-tree',tree,'-p',parent,'-m','baobab census checkpoint piece '+sys.argv[1]])
run(['git','update-ref',branch,commit])
run(['git','push','origin',branch+':'+branch])
