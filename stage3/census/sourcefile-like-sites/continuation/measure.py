import os,sys,json,pathlib,subprocess,time,signal,gzip,shutil
out=pathlib.Path(__file__).resolve().parents[1]
def run(label,root,names,bound=1500,checkpoint=False):
 dest=out/'continuation'/label/'raw';dest.mkdir(parents=True,exist_ok=True)
 for name in names:
  stem=name.replace('/','__').removesuffix('.ts');artifact=lambda suffix:dest/(stem+suffix)
  status=artifact('.status.json')
  if status.exists():print('retained existing '+str(status),flush=True);continue
  command=['/workspace/scratch/baobab-latent',str(root),str(artifact('.jsonl'))]
  env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=str(root/name),LATENT_TIMES=str(artifact('.times.jsonl')))
  for key in ['LATENT_SPECULATIVE','LATENT_RUN_ONLY','LATENT_RETRY_FILES']:env.pop(key,None)
  start=time.monotonic();timed_out=False
  with artifact('.log').open('wb') as log:
   child=subprocess.Popen(command,stdout=log,stderr=subprocess.STDOUT,env=env,start_new_session=True)
   try:child.wait(timeout=bound)
   except subprocess.TimeoutExpired:
    timed_out=True;os.killpg(child.pid,signal.SIGTERM)
    try:child.wait(timeout=10)
    except subprocess.TimeoutExpired:os.killpg(child.pid,signal.SIGKILL);child.wait()
  times=artifact('.times.jsonl');progress=[]
  if times.exists():
   for line in times.read_text().splitlines():
    try:progress.append(json.loads(line))
    except ValueError:pass
  record=dict(file=name,command=command,child_exit=child.returncode,wall_seconds=time.monotonic()-start,timed_out=timed_out,bound_seconds=bound,max_concurrent_files=1,completed_timing_units=len(progress),last_timing_units=progress[-5:])
  status.write_text(json.dumps(record,indent=2)+'\n')
  for suffix in ['.jsonl','.times.jsonl','.log']:
   p=artifact(suffix)
   if p.exists() and p.stat().st_size>1000000:
    with p.open('rb') as src,gzip.open(str(p)+'.gz','wb') as target:shutil.copyfileobj(src,target)
    p.unlink()
  print(json.dumps(record),flush=True)
  if checkpoint:
   for command in [['python3',str(out/'collect.py')],['node',str(out/'classify.cjs')],['python3',str(out/'report.py')],['python3',str(out/'checkpoint.py'),'long-'+stem]]:
    log=out/'continuation'/(stem+'-'+pathlib.Path(command[1]).stem+'.log')
    with log.open('wb') as destlog:p=subprocess.run(command,stdout=destlog,stderr=subprocess.STDOUT,cwd=out.parents[2])
    print('child exit='+str(p.returncode)+' '+repr(command),flush=True)
    if p.returncode:raise SystemExit(p.returncode)
if __name__=='__main__':
 label=sys.argv[1];root=pathlib.Path(sys.argv[2]);run(label,root,sys.argv[3:],checkpoint=label=='baseline')
