from pathlib import Path
import subprocess,os,time,signal,json,sys
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/debug-messages');raw=s/'raw'
def run(name,args,cwd=repo,env=None,limit=1490):
 start=time.monotonic()
 with (raw/(name+'.log')).open('w') as log:
  p=subprocess.Popen([str(a) for a in args],cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  timed=False
  try:
   while p.poll() is None:
    if time.monotonic()-start>limit:
     timed=True;os.killpg(p.pid,signal.SIGTERM)
     try:p.wait(timeout=10)
     except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);p.wait()
     break
    time.sleep(1)
  finally:
   if p.poll() is None:
    os.killpg(p.pid,signal.SIGKILL);p.wait()
 result={'command':[str(a) for a in args],'cwd':str(cwd),'exit':p.returncode,'seconds':time.monotonic()-start,'timed_out':timed}
 (raw/(name+'.json')).write_text(json.dumps(result,indent=2)+'\n');print(name,json.dumps(result),flush=True)
 if p.returncode:raise SystemExit(p.returncode if p.returncode>0 else 1)
 return result
if __name__=='__main__':
 mode=sys.argv[1];env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE='/workspace/scratch/debug-sites/TypeScript')
 if mode=='apply':run('after-apply',['bash','stage3/apply.sh',s/'after'],env=env)
 elif mode=='build':
  side=sys.argv[2];tree=s/side;modules=tree/'node_modules'
  if not modules.exists():modules.symlink_to(repo/'stage3/typescript/node_modules',target_is_directory=True)
  run(side+'-build',['npm','run','build'],cwd=tree,env=env)
 elif mode=='compiler':run('adamic-build',['go','build','-o',s/'adamic','./cmd/adamic'],env=env)
 elif mode=='types':
  side=sys.argv[2];roots=sorted((s/side/'src/compiler').rglob('*.ts'));assert len(roots)==79
  run(side+'-types',[s/'adamic','types',*roots],env=env)
