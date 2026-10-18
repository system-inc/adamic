import pathlib, subprocess, json, time, sys
root=pathlib.Path.cwd(); out=root/'review/compiler/loom-r-views-fx-b-ac'
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

mode=sys.argv[1]
common=['-count=1','-json','-timeout','90s']
if mode=='focused':
 cmd=['go','test','./internal/lower','./internal/oracle','./internal/native','./internal/javascript','-run','^TestFX6|^TestScalarUnionView|^TestPredicateBodyProof$|^TestRuntimeSetsAreEachUsed$|^TestCountsAreRecordedTable$|^TestObjectMetadataCacheIsPerEmission$|^TestViewMixedUnionSelectsNumber$',*common]
elif mode=='readers':
 packages=json.loads((root/'review/compiler/loom-r-views-fx-b/reader-packages.json').read_text())
 cmd=['go','test',*packages,'-run','^TestCallTargetReaders$',*common]
elif mode=='fixtures':
 cmd=['go','test','./internal/oracle','-run','^TestNativeAgreesWithNode_[0-9]+$/internal/lower/testdata/(fx6|scalar_union_views)',*common]
elif mode=='vet':cmd=['go','vet','./internal/lower','./internal/native','./internal/javascript','./internal/oracle']
else:raise SystemExit(mode)
raise SystemExit(run(mode,cmd,240))
