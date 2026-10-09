import subprocess,pathlib,json,time,gzip,shutil
root=pathlib.Path('/workspace/adamic');p=root/'review/test-defend/internal-flow';plan=json.loads((p/'plan.json').read_text());result=[]
assigned=['TestFlowSingleAssignment____oracle_testdata_timsort_a_63f1305dfb95','TestFlowMutationRanges____oracle_testdata_timsort_a_63f1305dfb95','TestFlowGraphPaths____oracle_testdata_timsort_a_63f1305dfb95','TestFlowLiveness____oracle_testdata_timsort_a_63f1305dfb95']
extra=['TestFlowCorpusUnitsCoverEveryProgram','TestFlowCorpusSetupIsShared','TestFlowCorpusRemainder','TestFlowProgram_testdata_joins_a_091c4a59e83f','TestFlowProgram_testdata_mutations_a_fe30ed94ac7e','TestDebuggerHasNoFlowInstruction']
bounded='^('+'|'.join(assigned+extra)+')$'
def run(mid,pattern,label):
 cmd=f"source /workspace/adamic-tools/env.sh; ADAMIC_BUILD_CACHE_DIR=/tmp/defend-flow/cache/{mid}-{label} timeout 120 go test -json -count=1 -timeout 90s ./internal/flow/ -run '{pattern}'"
 start=time.monotonic();logpath=p/(mid+'-'+label+'.log')
 with logpath.open('w') as log:process=subprocess.run(['bash','-c',cmd],cwd=root,stdout=log,stderr=subprocess.STDOUT)
 wall=time.monotonic()-start;fails=[];passes=[];lines={};timeout=False;panic=False;final=None
 with logpath.open() as log:
  for s in log:
   try:e=json.loads(s)
   except:continue
   name=e.get('Test','');output=e.get('Output','');top=name.split('/')[0]
   if 'test timed out after' in output:timeout=True
   if output.startswith('panic:'):panic=True
   if e.get('Action')=='fail' and name and '/'not in name:fails.append(name)
   if e.get('Action')=='pass' and name and '/'not in name:passes.append(name)
   if '.go:' in output and 'build 'not in output and len(lines.get(top,[]))<3:lines.setdefault(top,[]).append(output.strip())
   if e.get('Action') in ['pass','fail'] and 'Test'not in e:final=e
 with logpath.open('rb') as src,gzip.open(str(logpath)+'.gz','wb') as dest:shutil.copyfileobj(src,dest)
 logpath.unlink()
 r=dict(command=cmd,log=logpath.name+'.gz',wall_seconds=wall,exit=process.returncode,timeout=timeout,panic=panic,failed=fails,passed=passes,failing_lines={n:lines.get(n,[]) for n in fails},assigned_lines={n:lines.get(n,[]) for n in assigned},package_result=final)
 print(mid,label,round(wall,2),'failed',len(fails),'timeout',timeout,flush=True);return r
for recipe in plan:
 mid=recipe['mutant'];f=root/recipe['file'];original=f.read_text();assert original.count(recipe['old'])==1
 try:
  f.write_text(original.replace(recipe['old'],recipe['new']))
  with (p/(mid+'-vet.log')).open('w') as log:v=subprocess.run(['bash','-c','source /workspace/adamic-tools/env.sh; timeout 90 go vet ./internal/flow/'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  assert v.returncode==0
  full=run(mid,'.','whole');runs=[full]
  if full['timeout'] or full['panic'] or full['exit']==124:runs.append(run(mid,bounded,'bounded'))
  result.append(dict(recipe,runs=runs));(p/'matrix.json').write_text(json.dumps(result,indent=2))
 finally:f.write_text(original)
