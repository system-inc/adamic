import pathlib,subprocess,json,time,os,signal,re,datetime,shutil
root=pathlib.Path('/workspace/adamic');out=root/'review/test-defend/deletion-set/stage1-typescript-parser';skip='^(TestJsxMemberNameRejection|TestWholeCompilerAgrees)($|_|/)'
mutants=json.loads((out/'mutant-list.json').read_text());matrix=[]
def witness(name):
 return bool(re.search(r'Mutant|PlantedFailure|CatchesMutant',name)) or name=='TestExpressionsAgree'
# Stop only after a completed ordinary top-level failure. Witness and panic outputs alone never settle a replay.
for m in mutants:
 if m['stale']:matrix.append(m);continue
 patch=out/m['diff'];files=re.findall(r'^\+\+\+ b/(.*)$',patch.read_text(),re.M);original={f:(root/f).read_bytes() for f in files}
 try:
  subprocess.run(['git','apply',str(patch)],cwd=root,check=True)
  panics=[];attempts=[]
  while True:
   pattern=skip if not panics else '^('+ '|'.join(['TestJsxMemberNameRejection','TestWholeCompilerAgrees']+panics)+')($|_|/)'
   cache='/tmp/deletion-parser/cache/'+m['mutant']+'-'+str(len(attempts))
   cmd=f"source /workspace/adamic-tools/env.sh; ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR={cache} go test -json -count=1 -timeout 30m ./stage1/typescript/parser/ -skip '{pattern}'"
   logpath=out/(m['mutant']+'-'+str(len(attempts))+'.log');start=time.monotonic();failed=[];seenpanics=[];passed=[];witnessfails=[];events=[];stopped=False
   with logpath.open('w') as log:
    proc=subprocess.Popen(['bash','-c',cmd],cwd=root,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
    with logpath.open() as reader:
     while proc.poll() is None:
      for s in reader:
       try:e=json.loads(s)
       except:continue
       events.append(e);name=e.get('Test','');top=name.split('/')[0];output=e.get('Output','')
       if output.startswith('panic:') and top and top not in seenpanics:seenpanics.append(top)
       if e.get('Action')=='pass' and name and '/'not in name:passed.append(name)
       if e.get('Action')=='fail' and name and '/'not in name:
        failed.append(name)
        if witness(name):witnessfails.append(name)
        elif name not in seenpanics:
         stopped=True;os.killpg(proc.pid,signal.SIGTERM);break
      if stopped:break
      time.sleep(.5)
     proc.wait()
     for s in reader:
      try:e=json.loads(s)
      except:continue
      events.append(e);name=e.get('Test','');output=e.get('Output','');top=name.split('/')[0]
      if output.startswith('panic:') and top and top not in seenpanics:seenpanics.append(top)
      if e.get('Action')=='fail' and name and '/'not in name and name not in failed:failed.append(name)
   ordinary=[n for n in failed if not witness(n) and n not in seenpanics]
   attempts.append(dict(command=cmd,log=logpath.name,wall_seconds=time.monotonic()-start,returncode=proc.returncode,stopped_after_catch=stopped,failed=failed,witness_failures=witnessfails,panicking_tests=seenpanics,still_caught_by=ordinary,passed=passed,failing_lines=[e['Output'].strip() for e in events if e.get('Test','').split('/')[0] in failed and '.go:' in e.get('Output','')]))
   if ordinary or (proc.returncode==0):break
   new=[n for n in seenpanics if n not in panics]
   if not new:break
   panics.extend(new)
  result=dict(m,still_caught_by=attempts[-1]['still_caught_by'],attempts=attempts,panicking_tests=panics,complete_pass=attempts[-1]['returncode']==0,broken=not attempts[-1]['still_caught_by'] and attempts[-1]['returncode']!=0)
  matrix.append(result);(out/'matrix.json').write_text(json.dumps(matrix,indent=2));print(m['mutant'],attempts[-1]['wall_seconds'],result['still_caught_by'],flush=True)
 finally:
  for f,data in original.items():(root/f).write_bytes(data)
shutil.copy('/tmp/deletion-parser/gather.py',out/'gather.py');shutil.copy('/tmp/deletion-parser/replay.py',out/'replay.py')
