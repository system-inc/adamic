import json, os, pathlib, subprocess, threading, time, tempfile
root=pathlib.Path('/workspace/adamic')
evidence=root/'stage1/cohere/typeaware/wave_10_leaf/evidence/unified-landing'
evidence.mkdir(parents=True,exist_ok=True)
profiles=tempfile.mkdtemp(prefix='lint-wave10-profiles-',dir='/workspace/scratch')
env=os.environ.copy()
env.update(GOMAXPROCS='4', GOFLAGS='-buildvcs=false', ADAMIC_LINT_BENCH='1', ADAMIC_TYPESCRIPT_SOURCE='/workspace/typescript-lint-landing', ADAMIC_LINT_PROFILE_DIR=profiles, ADAMIC_LINT_PROFILE_SNAPSHOTS=profiles)
command=['go','test','-json','-count=1','-timeout=90m','./stage1/cohere/lint']
load=[]; done=threading.Event()
def sample():
 while not done.is_set():
  load.append({'seconds':time.monotonic()-started,'load':os.getloadavg()})
  done.wait(10)
inputs={'command':command,'inputs':{key:env[key] for key in ['GOMAXPROCS','GOFLAGS','ADAMIC_LINT_BENCH','ADAMIC_TYPESCRIPT_SOURCE','ADAMIC_LINT_PROFILE_DIR','ADAMIC_LINT_PROFILE_SNAPSHOTS']},'nproc':subprocess.check_output(['nproc'],text=True).strip(),'typescript_sha':subprocess.check_output(['git','-C',env['ADAMIC_TYPESCRIPT_SOURCE'],'rev-parse','HEAD'],text=True).strip(),'typescript_status_before':subprocess.check_output(['git','-C',env['ADAMIC_TYPESCRIPT_SOURCE'],'status','--short'],text=True),'branch_sha':subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip()}
(evidence/'all-inputs.json').write_text(json.dumps(inputs,indent=2)+'\n')
started=time.monotonic(); worker=threading.Thread(target=sample); worker.start()
with (evidence/'lint-all.jsonl').open('w') as output:
 result=subprocess.run(command,cwd=root,env=env,stdout=output,stderr=subprocess.STDOUT)
done.set();worker.join();wall=time.monotonic()-started
counts={'pass':0,'fail':0,'skip':0}; top=counts.copy(); skips=[]; failures=[]
for line in (evidence/'lint-all.jsonl').read_text().splitlines():
 try:event=json.loads(line)
 except ValueError:continue
 action=event.get('Action');test=event.get('Test')
 if test and action in counts:
  counts[action]+=1
  if '/' not in test:top[action]+=1
  if action=='skip':skips.append(test)
  if action=='fail':failures.append(test)
summary={'exit':result.returncode,'wall_seconds':wall,'all_tests':counts,'top_level':top,'skips':skips,'failures':failures,'load_samples':load,'nproc':inputs['nproc'],'typescript_status_after':subprocess.check_output(['git','-C',env['ADAMIC_TYPESCRIPT_SOURCE'],'status','--short'],text=True)}
(evidence/'all-inputs-results.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps({key:value for key,value in summary.items() if key!='load_samples'}))
raise SystemExit(result.returncode)
