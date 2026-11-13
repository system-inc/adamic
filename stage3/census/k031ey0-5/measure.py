import json,os,signal,subprocess,sys,time,datetime,socket
from pathlib import Path
base=Path('/workspace/scratch/k031ey0-5'); out=base/'evidence'; tree=base/'adapted'
def box():
 return dict(utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),hostname=socket.gethostname(),cpus=len(os.sched_getaffinity(0)),load=os.getloadavg(),cpu_max=Path('/sys/fs/cgroup/cpu.max').read_text().strip())
def run(name,cmd,env=None):
 start=time.monotonic(); before=box()
 with (out/(name+'.log')).open('wb') as log:
  p=subprocess.Popen(cmd,stdout=log,stderr=log,env=env,start_new_session=True)
  stop=None
  try: rc=p.wait(timeout=1780)
  except subprocess.TimeoutExpired:
   stop='elapsed bound'; os.killpg(p.pid,signal.SIGTERM)
   try: rc=p.wait(timeout=10)
   except subprocess.TimeoutExpired: os.killpg(p.pid,signal.SIGKILL); rc=p.wait()
  finally:
   try: os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError: pass
 status=dict(command=cmd,exit_code=rc,seconds=time.monotonic()-start,stopping_reason=stop,before=before,after=box())
 (out/(name+'-status.json')).write_text(json.dumps(status,indent=2)+'\n'); print(name,'child_exit='+str(rc),round(status['seconds'],3),flush=True)
 if rc: raise SystemExit(rc if rc>0 else 128-rc)
 return status
if sys.argv[1] in ('group','range'):
 group=int(sys.argv[2]); roots=sorted((tree/'src/compiler').rglob('*.ts')); assert len(roots)==79,len(roots)
 (out/'roots.json').write_text(json.dumps([str(r.relative_to(tree)) for r in roots],indent=2)+'\n')
 start,end=((group-1)*20,min(group*20,79)) if sys.argv[1]=='group' else (int(sys.argv[2])-1,int(sys.argv[3]))
 assert 0 <= start < end <= 79
 assert sys.argv[1]!='range' or end-start <= 10
 for i in range(start,end):
  for side in ('parent','head'):
   name=f'{side}-root-{i+1:02}'; dest=out/(name+'.jsonl')
   run(name,[str(base/(side+'-census')),str(roots[i]),str(dest)])
   rows=[json.loads(l) for l in dest.read_text().splitlines()]; assert len(rows)==2
   assert {k:v for k,v in rows[0].items() if k!='seconds'}=={k:v for k,v in rows[1].items() if k!='seconds'}
elif sys.argv[1]=='latent':
 side=sys.argv[2]; env=dict(os.environ,LATENT_FULL='0',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(out/(side+'-latent-times.jsonl')),LATENT_PREPASS_TIMES=str(out/(side+'-prepass-times.tsv')))
 run(side+'-latent',[str(base/(side+'-entry-census')),str(tree/'src/tsc/tsc.ts'),str(out/(side+'-latent.jsonl'))],env)
