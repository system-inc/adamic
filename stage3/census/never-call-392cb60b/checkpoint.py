import sys,subprocess,pathlib,json,time
repo=pathlib.Path('/workspace/never-call-report'); label=sys.argv[1]; folder='stage3/census/never-call-392cb60b'
def run(cmd,**kw):
 r=subprocess.run(cmd,cwd=repo,**kw)
 print('child exit code:',r.returncode,cmd[0:2],flush=True)
 if r.returncode:sys.exit(r.returncode)
 return r
run(['git','add',folder])
tree=subprocess.check_output(['git','write-tree'],cwd=repo,text=True).strip()
sha=subprocess.check_output(['git','commit-tree',tree,'-p','bdbc73864f82dbc0c51434decf0ffe96d4782767'],input=f'never-call census checkpoint: {label}\n',cwd=repo,text=True).strip()
ref='refs/heads/typescript/never-call-census-checkpoint-'+label
with open('/tmp/never-call-checkpoint-'+label+'.log','wb') as f:run(['git','push','origin',sha+':'+ref],stdout=f,stderr=subprocess.STDOUT)
print(json.dumps(dict(label=label,sha=sha,ref=ref)),flush=True)
