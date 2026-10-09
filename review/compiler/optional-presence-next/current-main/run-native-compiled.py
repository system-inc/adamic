import json,os,pathlib,signal,subprocess,time
out=pathlib.Path('review/compiler/optional-presence-next/current-main')
names=['TestRuntimeFeatureMismatchProgramFeature','TestRuntimeFeatureMismatchRuntimeFeature','TestRuntimeFeatureSetsNative','TestRuntimeFeatureSetsWASI','TestTSGoBuildSeesTheProgramsFeatures','TestRecordsAgainstNode']+['TestRecordMutantsUnit%02d'%i for i in range(6)]+['TestWASIUnit%02d'%i for i in range(39)]
with (out/'native-results.jsonl').open('w') as results:
 for name in names:
  env=dict(os.environ,ADAMIC_TEST_WASI='1')
  if name.startswith('TestWASIUnit'):env['PATH']='/workspace/adamic-tools/wasi-sdk/bin:'+env['PATH']
  args=['go','tool','test2json','-t','-p','github.com/system-inc/adamic/internal/native','/workspace/scratch/landing-native.test','-test.run=^'+name+'$','-test.count=1','-test.timeout=80s','-test.v=test2json'];start=time.monotonic();log=out/(name+'.log')
  with log.open('w') as f:
   child=subprocess.Popen(args,cwd='/workspace/adamic/internal/native',env=env,stdout=f,stderr=subprocess.STDOUT,start_new_session=True)
   try:status=child.wait(timeout=85)
   except subprocess.TimeoutExpired:os.killpg(child.pid,signal.SIGKILL);child.wait();status=124
  events=[]
  for line in log.read_text().splitlines():
   try:event=json.loads(line)
   except ValueError:continue
   if event.get('Test')==name and event.get('Action') in ['pass','fail','skip']:events.append(event)
  row=dict(name=name,command=args,status=status,seconds=round(time.monotonic()-start,3),events=events);results.write(json.dumps(row)+'\n');results.flush();print(name,status,row['seconds'],flush=True)
