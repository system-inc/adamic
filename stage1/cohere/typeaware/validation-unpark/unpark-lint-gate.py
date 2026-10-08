import os,sys,json,time,subprocess,pathlib
root,label=sys.argv[1:];env=os.environ.copy();base='/tmp/unpark-'+label
profiles=base+'-profiles';pathlib.Path(profiles).mkdir(exist_ok=True)
env.update(GOCACHE='/tmp/unpark-go-cache',XDG_CACHE_HOME='/tmp/unpark-runtime-cache',GOMAXPROCS='4',GOFLAGS='-buildvcs=false',GOPROXY='https://proxy.golang.org|direct',ADAMIC_TYPESCRIPT_SOURCE='/workspace/wave-11-typescript',ADAMIC_LINT_BENCH='1',ADAMIC_LINT_PROFILE_DIR=profiles,ADAMIC_LINT_PROFILE_SNAPSHOTS=profiles)
args=['go','test','-json','-count=1','-timeout=90m','./stage1/cohere/lint'];start=time.monotonic();loads=[]
with open(base+'-package.jsonl','w') as log:
 p=subprocess.Popen(args,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
 while p.poll() is None:
  loads.append([round(time.monotonic()-start,3),*os.getloadavg()]);time.sleep(10)
result=dict(argv=args,exit=p.returncode,wall=round(time.monotonic()-start,3),nproc=len(os.sched_getaffinity(0)),loads=loads)
pathlib.Path(base+'-package-result.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps({k:v for k,v in result.items() if k!='loads'}),flush=True)
