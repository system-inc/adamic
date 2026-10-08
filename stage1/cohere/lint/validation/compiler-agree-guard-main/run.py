import os,subprocess,time,json,pathlib,threading,sys,gzip
repo=pathlib.Path('/workspace/adamic'); dest=repo/'stage1/cohere/lint/validation/compiler-agree-guard-main'
mode=sys.argv[1]; burners=[]; stop=threading.Event()
env=os.environ.copy(); profile=f'/workspace/scratch/lint-guard-{mode}-profile'
env.update(ADAMIC_TYPESCRIPT_SOURCE='/workspace/scratch/typescript-6.0.3',ADAMIC_LINT_BENCH='1',ADAMIC_LINT_PROFILE_DIR=profile,ADAMIC_LINT_PROFILE_SNAPSHOTS=profile,ADAMIC_GATE_UNCACHED='1')
def burner_sample(p):
 stat=pathlib.Path(f'/proc/{p.pid}/stat').read_text().split()
 return dict(pid=p.pid,alive=p.poll() is None,affinity=sorted(os.sched_getaffinity(p.pid)),cpu_seconds=(int(stat[13])+int(stat[14]))/os.sysconf('SC_CLK_TCK'))
def sample():
 with open(dest/f'{mode}-load.jsonl','w') as f:
  while not stop.is_set():
   f.write(json.dumps(dict(time=time.time(),load=pathlib.Path('/proc/loadavg').read_text().strip(),burners=[burner_sample(p) for p in burners]))+'\n');f.flush();stop.wait(10)
try:
 if mode=='loaded':
  for cpu in sorted(os.sched_getaffinity(0)):
   p=subprocess.Popen(['taskset','-c',str(cpu),'python3','-c','while True: pass'],stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL);burners.append(p)
 thread=threading.Thread(target=sample);thread.start()
 started=time.time(); before=pathlib.Path('/proc/loadavg').read_text().strip()
 command=['go','test','./stage1/cohere/lint','-count=1','-json','-v','-timeout=0']
 with open(dest/f'{mode}.jsonl','w') as f:
  result=subprocess.run(command,cwd=repo,env=env,stdout=f,stderr=subprocess.STDOUT)
 metadata=dict(command=command,inputs={k:env[k] for k in ['ADAMIC_TYPESCRIPT_SOURCE','ADAMIC_LINT_BENCH','ADAMIC_LINT_PROFILE_DIR','ADAMIC_LINT_PROFILE_SNAPSHOTS','ADAMIC_GATE_UNCACHED']},wall_seconds=time.time()-started,exit=result.returncode,nproc=len(os.sched_getaffinity(0)),cpu_max=pathlib.Path('/sys/fs/cgroup/cpu.max').read_text().strip(),load_before=before,load_after=pathlib.Path('/proc/loadavg').read_text().strip(),burner_pids=[p.pid for p in burners])
 (dest/f'{mode}-wall.json').write_text(json.dumps(metadata,indent=2)+'\n')
finally:
 stop.set()
 if 'thread' in locals():thread.join()
 for p in burners:p.terminate()
 for p in burners:p.wait()
