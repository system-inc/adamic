import pathlib,subprocess,sys
out=pathlib.Path(__file__).resolve().parent
repo=out.parents[2]
def run(args):
 p=subprocess.run(args,cwd=repo,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,text=True)
 print('child exit='+str(p.returncode)+' '+repr(args)+'\n'+p.stdout,flush=True)
 if p.returncode:sys.exit(p.returncode)
 return p.stdout.strip()
run(['git','add','stage3/census/sourcefile-like-sites'])
branch='refs/heads/typescript/sourcefile-like-sites-checkpoint'
p=subprocess.run(['git','rev-parse','--verify',branch],cwd=repo,capture_output=True,text=True)
parent=p.stdout.strip() if p.returncode==0 else run(['git','rev-parse','HEAD'])
tree=run(['git','write-tree'])
commit=run(['git','commit-tree',tree,'-p',parent,'-m','baobab census checkpoint piece '+sys.argv[1]])
run(['git','update-ref',branch,commit])
run(['git','push','origin',branch+':'+branch])
