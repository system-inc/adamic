import json,os,pathlib,subprocess,time
root=pathlib.Path(__file__).resolve().parents[3]
os.chdir(root)
work=root/'.split-pool-work'; work.mkdir(exist_ok=True)
evidence=root/'internal/native/split_pool_evidence'; evidence.mkdir(exist_ok=True)
replace={}
for name in ['native.go','tsgo.go','units.go','library.go','units_tsgo.go','stage1_profile.go']:
    original=work/('baseline-'+name)
    original.write_bytes(subprocess.check_output(['git','show','c9551640:internal/native/'+name]))
    replace[str(root/'internal/native'/name)]=str(original)
(work/'baseline.json').write_text(json.dumps({'Replace':replace}))
probe=work/'probe'; probe.mkdir(exist_ok=True)
(probe/'main.go').write_text((evidence/'probe.go.txt').read_text())
for label,overlay in [('baseline',['-overlay='+str(work/'baseline.json')]),('pool',[])]:
    subprocess.run(['go','build',*overlay,'-o',str(work/('probe-'+label)),str(probe)],check=True)
    subprocess.run(['go','test',*overlay,'./internal/oracle','-run=^$','-count=1'],check=True)

env=os.environ.copy(); env['GOCACHE']=subprocess.check_output(['go','env','GOCACHE'],text=True).strip(); env['ADAMIC_GATE_UNCACHED']='1'; env.pop('ADAMIC_NATIVE_JOBS',None)
meta={'cpus':len(os.sched_getaffinity(0)), 'cpu_max':pathlib.Path('/sys/fs/cgroup/cpu.max').read_text().strip(),'load':pathlib.Path('/proc/loadavg').read_text().strip(), 'flags':['-count=1','-timeout=30m'],'uncached':True}
(evidence/'oracle.jsonl').write_text('')
(evidence/'machine.json').write_text(json.dumps(meta,indent=2)+'\n')
for label,binary,split in [('off','probe-baseline','0'),('unpooled','probe-baseline','1'),('pooled','probe-pool','1')]:
 env.pop('ADAMIC_NATIVE_SPLIT',None)
 if split=='0': env['ADAMIC_NATIVE_SPLIT']='0'
 args=[str(work/binary),label]
 if split=='1':args.append('split')
 with (evidence/('probe-'+label+'.jsonl')).open('w') as log:
  result=subprocess.run(args,env=env,stdout=log,stderr=subprocess.STDOUT)
 if result.returncode:raise SystemExit('probe failed '+label)
 print('probe complete '+label,flush=True)
for round in range(1,4):
 for label,split,baseline in [('off','0',True),('unpooled','1',True),('pooled','1',False)]:
  env.pop('ADAMIC_NATIVE_SPLIT',None)
  if split=='0': env['ADAMIC_NATIVE_SPLIT']='0'
  # Reproduce CPU-sized jobs measured before the go-test fallback was applied.
  env['ADAMIC_NATIVE_JOBS']=str(len(os.sched_getaffinity(0)))
  # Runtime cache is cold for every sample; oracle result and unit caches are bypassed.
  cache=pathlib.Path(subprocess.check_output(['mktemp','-d','/tmp/adamic-gate/pool-oracle.XXXXXX'],text=True).strip());env['XDG_CACHE_HOME']=str(cache)
  args=['go','test']
  if baseline:args.append('-overlay='+str(work/'baseline.json'))
  args+=['./internal/oracle','-count=1','-timeout=30m']
  before=pathlib.Path('/proc/loadavg').read_text().strip();start=time.monotonic()
  with (evidence/('oracle-'+label+'-'+str(round)+'.log')).open('w') as log:
   result=subprocess.run(args,env=env,stdout=log,stderr=subprocess.STDOUT)
  row={'mode':label,'round':round,'seconds':time.monotonic()-start,'exit':result.returncode,'load_before':before,'load_after':pathlib.Path('/proc/loadavg').read_text().strip(),'command':args}
  with (evidence/'oracle.jsonl').open('a') as output:output.write(json.dumps(row)+'\n')
  print(json.dumps(row),flush=True)
