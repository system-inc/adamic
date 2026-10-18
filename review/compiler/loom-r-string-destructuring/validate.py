import subprocess,time,json
from pathlib import Path
out=Path(__file__).resolve().parent
jobs=[
('guards',['go','test','./internal/ir','./internal/lower','./cmd/adamic-gate','./internal/oracle','-run','^(TestEveryMapAProgramHoldsIsSorted|TestNativeAgreementRowsNameRegisteredPrograms|TestEveryTestIsParallelOrSaysWhy|TestEveryDuplicatedProductDeclarationAgrees|TestRuntimeSetsAreEachUsed)$','-count=1','-timeout','90s','-v','-parallel','4']),
('lower',['go','test','./internal/lower','-run','^(TestStringDestructuring.*|Test.*Destructur.*)$','-count=1','-timeout','90s','-v','-parallel','4']),
('fixture',['go','test','./internal/oracle','-run',r'^TestNativeAgreesWithNode_097$/^internal$/^oracle$/^testdata$/^string_destructuring\.a$','-count=1','-timeout','90s','-v','-parallel','4']),
('readers',['go','test','./internal/ir','-run','^TestCallTargetReaders$','-count=1','-timeout','90s','-v']),
('flow',['go','test','./internal/flow','-run','^(TestFlowCorpusRemainder|TestFlowCorpusUnitsCoverEveryProgram|TestFlowCorpusSetupIsShared)$','-count=1','-timeout','90s','-v']),
('vet',['go','vet','./internal/lower','./internal/oracle'])]
rows=[]
for name,args in jobs:
 start=time.monotonic()
 with (out/(name+'.log.txt')).open('w') as log:
  try:code=subprocess.run(args,stdout=log,stderr=subprocess.STDOUT,timeout=180).returncode
  except subprocess.TimeoutExpired:code=124
 row=dict(name=name,command=args,exit=code,seconds=round(time.monotonic()-start,3));rows.append(row);(out/'validation.json').write_text(json.dumps(rows,indent=2)+'\n');print(json.dumps(row),flush=True)
 if code:break
