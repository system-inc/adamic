import os, pathlib, subprocess, json, time, gzip, shutil, sys, fcntl
root=pathlib.Path('/workspace/scratch/each-node-8n7eet3')
report=root/'report'
dest=report/'stage3/census/each-node-34437289'
dest.mkdir(parents=True,exist_ok=True)
env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE=str(root/'typescript'))
def run(name,args,cwd,extra=None):
 e=dict(env); e.update(extra or {})
 started=time.time()
 with (root/(name+'.log')).open('w') as log:
  code=subprocess.run([sys.executable,str(root/'measure.py'),str(root/(name+'.time')),*map(str,args)],cwd=cwd,env=e,stdout=log,stderr=subprocess.STDOUT).returncode
 receipt={'command':list(map(str,args)),'cwd':str(cwd),'exit':code,'wall_seconds':time.time()-started}
 (root/(name+'.receipt.json')).write_text(json.dumps(receipt,indent=2)+'\n')
 print(name,'child_exit='+str(code),flush=True)
 return code
def checkpoint(label):
 lock=(root/"checkpoint.lock").open("w"); fcntl.flock(lock,fcntl.LOCK_EX)
 for p in root.iterdir():
  if p.is_file() and p.name!='control.py':
   data=p.read_bytes()
   target=dest/p.name
   if len(data)>1000000:
    target=dest/(p.name+'.gz'); data=gzip.compress(data,mtime=0)
   target.write_bytes(data)
 (dest/'CHECKPOINT.md').write_text('Task 8n7eet3 site 2. Report in progress. Last completed step: '+label+'\n')
 subprocess.run(['git','add','--','stage3/census/each-node-34437289'],cwd=report,check=True)
 tree=subprocess.check_output(['git','write-tree'],cwd=report,text=True).strip()
 ref='refs/heads/typescript/each-node-census-checkpoint'
 old=subprocess.run(['git','rev-parse','--verify',ref],cwd=report,capture_output=True,text=True)
 parent=old.stdout.strip() if old.returncode==0 else '34437289896454f4076396d7bf913e70f6b9a826'
 commit=subprocess.check_output(['git','-c','user.name=Codex','-c','user.email=codex@openai.com','commit-tree',tree,'-p',parent,'-m','Census checkpoint: '+label],cwd=report,text=True).strip()
 subprocess.run(['git','update-ref',ref,commit],cwd=report,check=True)
 code=run('push-checkpoint-'+label,['git','push','origin',ref+':'+ref],report)
 if code: raise SystemExit(code)
if __name__=='__main__':
 action=sys.argv[1]
 if action=='apply':
  name=sys.argv[2]
  code=run('apply-'+name,['bash','stage3/apply.sh',str(root/('adapted-'+name))],root/name)
  checkpoint('apply-'+name)
  sys.exit(code)
 elif action=='checkpoint': checkpoint(sys.argv[2])
