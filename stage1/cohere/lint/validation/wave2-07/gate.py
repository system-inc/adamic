import subprocess,pathlib,os,json,time,datetime,threading
root=pathlib.Path('/workspace/scratch/lint-batch-wave2-07');out=pathlib.Path('/tmp/wave2-07-gate');out.mkdir(exist_ok=True)
profile=pathlib.Path('/tmp/wave2-07-profile-'+str(os.getpid()));profile.mkdir();assert not list(profile.iterdir())
env=os.environ|{'GOPROXY':'https://proxy.golang.org|direct','ADAMIC_TYPESCRIPT_SOURCE':'/workspace/scratch/typescript-6.0.3','ADAMIC_LINT_BENCH':'1','ADAMIC_PARSER_BENCH':'1','ADAMIC_LINT_PROFILE_DIR':str(profile),'ADAMIC_LINT_PROFILE_SNAPSHOTS':str(profile),'WASI_SYSROOT':'/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot','ADAMIC_GATE_UNCACHED':'1'}
env.pop('ADAMIC_LINT_RULES',None)
packages=['./stage1/cohere/lint','./stage1/cohere/lint/helpers','./stage1/cohere/lint/helpers/comments','./stage1/cohere/lint/helpers/property','./stage1/cohere/lint/helpers/next','./stage1/cohere/lint/shards','./stage1/typescript/parser']
meta={'head':subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip(),'inputs':{k:v for k,v in env.items() if k.startswith('ADAMIC_') or k in ['WASI_SYSROOT','GOPROXY']},'nproc':subprocess.check_output(['nproc'],text=True).strip(),'quota':pathlib.Path('/sys/fs/cgroup/cpu.max').read_text().strip(),'typescript_sha':subprocess.check_output(['git','rev-parse','HEAD'],cwd=env['ADAMIC_TYPESCRIPT_SOURCE'],text=True).strip(),'typescript_status':subprocess.check_output(['git','status','--porcelain','--ignored'],cwd=env['ADAMIC_TYPESCRIPT_SOURCE'],text=True),'packages':[]}
assert not meta['typescript_status'];done=threading.Event()
def samples():
 with (out/'resources.jsonl').open('w') as f:
  while not done.is_set():
   f.write(json.dumps({'time':datetime.datetime.now(datetime.timezone.utc).isoformat(),'load':pathlib.Path('/proc/loadavg').read_text().strip(),'cpu_stat':pathlib.Path('/sys/fs/cgroup/cpu.stat').read_text(),'memory':pathlib.Path('/sys/fs/cgroup/memory.current').read_text().strip()})+'\n');f.flush();done.wait(15)
threading.Thread(target=samples,daemon=True).start()
def save(): (out/'metadata.json').write_text(json.dumps(meta,indent=2))
save()
for package in packages:
 name=package.removeprefix('./').replace('/','-');log=out/(name+'.jsonl');start=time.monotonic();stamp=datetime.datetime.now(datetime.timezone.utc).isoformat();before=pathlib.Path('/proc/loadavg').read_text().strip();cmd=['go','test',package,'-count=1','-v','-json','-timeout=3h']
 print('START',package,stamp,flush=True)
 with log.open('w') as f:r=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
 row={'package':package,'command':cmd,'start':stamp,'end':datetime.datetime.now(datetime.timezone.utc).isoformat(),'wall':time.monotonic()-start,'exit':r.returncode,'load_before':before,'load_after':pathlib.Path('/proc/loadavg').read_text().strip(),'log':str(log)};meta['packages'].append(row);save();print('END',package,row['exit'],row['wall'],flush=True)
meta['typescript_status_after']=subprocess.check_output(['git','status','--porcelain','--ignored'],cwd=env['ADAMIC_TYPESCRIPT_SOURCE'],text=True);meta['complete']=True;save();done.set();print('DONE',flush=True)
