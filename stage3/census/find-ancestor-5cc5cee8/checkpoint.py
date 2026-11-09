import subprocess,json,sys
from pathlib import Path
out=Path(__file__).resolve().parent;repo=out.parents[2]
event=sys.argv[1]
subprocess.run(['git','add','-f','--',str(out)],cwd=repo,check=True)
tree=subprocess.check_output(['git','write-tree'],cwd=repo,text=True).strip()
sha=subprocess.check_output(['git','commit-tree',tree,'-p','c1e9379e'],cwd=repo,input='census: find ancestor bar '+event+' checkpoint\n',text=True).strip()
ref='refs/heads/typescript/find-ancestor-bar-census-checkpoint-'+event
with (out/('checkpoint-'+event+'.log')).open('w') as f:
 rc=subprocess.run(['git','push','origin',sha+':'+ref],cwd=repo,stdout=f,stderr=subprocess.STDOUT).returncode
(out/('checkpoint-'+event+'.json')).write_text(json.dumps(dict(sha=sha,ref=ref,exit=rc),indent=2)+'\n')
print(event,'push_child_exit='+str(rc),sha,flush=True)
raise SystemExit(rc)
