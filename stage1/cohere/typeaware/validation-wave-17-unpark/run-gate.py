import collections,json,os,pathlib,statistics,subprocess,time
root=pathlib.Path('/workspace/adamic');out=pathlib.Path('/workspace/unpark-wave17');env=dict(os.environ,GOMAXPROCS='4',GOFLAGS='-buildvcs=false',GOPROXY='https://proxy.golang.org|direct',ADAMIC_LINT_BENCH='1',ADAMIC_TYPESCRIPT_SOURCE='/workspace/wave-17-typescript',ADAMIC_LINT_PROFILE_DIR=str(out/'profiles-final'),ADAMIC_LINT_PROFILE_SNAPSHOTS=str(out/'profiles-final'))
(out/'profiles-final').mkdir(exist_ok=True)
(out/'runner.pid').write_text(str(os.getpid()))
commands=[('lint',['go','test','-json','-count=1','-timeout=90m','./stage1/cohere/lint']),('legacy',['go','test','-json','-count=1','-timeout=30m','./stage1/cohere/typeaware','-run','^Test(Wave17(Fifth|Fourth|Next|Third)AgreementAndMutants|Wave17HeadJudgmentsAndNativeJSX|Wave17AgreementMutantAndNativeJSX|Wave17UnicodeUpper|CoverageAgreementAndMutants)$'])]
for name,command in commands:
 start=time.monotonic();samples=[];state={'running':name,'command':command,'start_load':os.getloadavg()};(out/'state.json').write_text(json.dumps(state))
 with (out/(name+'.jsonl')).open('wb') as log:
  p=subprocess.Popen(command,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
  while p.poll() is None:
   samples.append(os.getloadavg());time.sleep(1)
 wall=time.monotonic()-start;counts=collections.Counter();top=collections.Counter();skips=[];failed=[];mutants=[]
 for line in (out/(name+'.jsonl')).read_text().splitlines():
  try:event=json.loads(line)
  except ValueError:continue
  test=event.get('Test');action=event.get('Action')
  if test and action in ['pass','fail','skip']:
   counts[action]+=1
   if '/' not in test:top[action]+=1
   if action=='skip':skips.append(test)
   if action=='fail':failed.append(test)
   if action=='pass' and test.startswith('TestMutants/'):mutants.append(test)
 result={'command':command,'exit':p.returncode,'wall_seconds':wall,'counts':dict(counts),'top_level':dict(top),'skips':skips,'failed':failed,'mutants':mutants,'nproc':int(subprocess.check_output(['nproc'])),'load_start':state['start_load'],'load_end':os.getloadavg(),'load_1min_min_median_max':[min(s[0] for s in samples),statistics.median(s[0] for s in samples),max(s[0] for s in samples)]}
 (out/(name+'-result.json')).write_text(json.dumps(result,indent=2)+'\n');(out/(name+'-loads.json')).write_text(json.dumps(samples)+'\n')
 print(name,result,flush=True)
(out/'state.json').write_text(json.dumps({'finished':True}))
