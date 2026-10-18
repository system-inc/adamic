import pathlib, subprocess, json, time, sys
root=pathlib.Path.cwd(); out=root/'review/compiler/loom-r-views-fx-b'
def run(label,cmd,limit=600):
 start=time.monotonic()
 with (out/(label+'.log')).open('w') as f:
  try: code=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,timeout=limit).returncode
  except subprocess.TimeoutExpired: code=124
 tests=[]
 for line in (out/(label+'.log')).read_text().splitlines():
  try: x=json.loads(line)
  except ValueError: continue
  if x.get('Action') in ('pass','fail','skip') and x.get('Test'): tests.append({k:x[k] for k in ('Package','Test','Action','Elapsed') if k in x})
 result=dict(command=cmd,exit=code,seconds=round(time.monotonic()-start,3),tests=tests)
 (out/(label+'.json')).write_text(json.dumps(result,indent=2)+'\n'); print(label,code,result['seconds'],len(tests),flush=True)
 return code
if __name__=='__main__':
 mode=sys.argv[1]
 if mode=='focused':
  sys.exit(run(mode,['go','test','./internal/lower','./internal/oracle','./internal/ir','./internal/javascript','./internal/native','-run',r'^TestFX6|^TestScalarUnionView|^TestPredicateBodyProof$|^TestCheckedView|^TestInterfaceCast|^TestRequiredView|^TestNarrowedField|^TestViewField|^TestDefaultTagged|^TestCallTargetReaders$|^TestViewMixedUnionSelectsNumber$|^TestObjectMetadataCacheIsPerEmission$|^TestRuntimeSetsAreEachUsed$|TestNativeAgreesWithNode_[0-9]+$/internal/lower/testdata/(fx6|scalar_union_views)','-count=1','-json','-timeout','90s'],180))
 if mode=='counts': sys.exit(run(mode,['go','test','./internal/oracle','-run','TestCountsAreRecorded','-count=1','-json','-timeout','540s','-args','-update-counts'],600))
 if mode=='vet':sys.exit(run(mode,['go','vet','./internal/lower','./internal/oracle','./internal/native','./internal/javascript','./internal/flow','./internal/ir'],180))
