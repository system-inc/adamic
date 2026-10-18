import subprocess,time,json,os
from pathlib import Path
out=Path(__file__).resolve().parent
jobs=[
 ('guards',['go','test','./internal/ir','./internal/lower','./cmd/adamic-gate','./internal/oracle','-run','^(TestEveryMapAProgramHoldsIsSorted|TestNativeAgreementRowsNameRegisteredPrograms|TestEveryTestIsParallelOrSaysWhy|TestEveryDuplicatedProductDeclarationAgrees|TestRuntimeSetsAreEachUsed)$','-count=1','-timeout','90s','-v','-parallel','4']),
 ('lower',['go','test','./internal/lower','-run','^(TestRecord.*|TestNamedRecord.*|TestPartialRecord.*|TestDetachedOwn.*|TestEntries.*|TestOptionalIndexing.*)$','-count=1','-timeout','90s','-v','-parallel','4']),
 ('oracle-fixtures',['go','test','./internal/oracle','-run',r'^TestNativeAgreesWithNode_[0-9]+$/^internal$/^oracle$/^testdata$/^(records_.*|detached_own_.*|json_replacer_duplicates_object\.a)$','-count=1','-timeout','90s','-v','-parallel','4']),
 ('oracle-checks',['go','test','./internal/oracle','-run','^(TestRecord.*|TestNamedRecord.*|TestPartialRecord.*|TestDetachedOwn.*|TestJSONReplacerDuplicates.*)$','-count=1','-timeout','90s','-v','-parallel','4']),
 ('native',['go','test','./internal/native','-run','^(TestJSONReplacerDuplicates.*Mutant|TestRecordsAgainstNode|TestRecordMutantsUnit0[0-5]|TestRecordReadMutants)$','-count=1','-timeout','90s','-v','-parallel','4']),
 ('readers',['go','test','./internal/ir','-run','^TestCallTargetReaders$','-count=1','-timeout','90s','-v']),
 ('flow',['go','test','./internal/flow','-run','^TestFlowCorpusRemainder$','-count=1','-timeout','90s','-v']),
 ('stage3',['go','test','./stage3/fixtures','-run','^TestFixturesRecords$','-count=1','-timeout','90s','-v']),
 ('vet',['go','vet','./internal/flow','./internal/fresh','./internal/ir','./internal/javascript','./internal/lower','./internal/native','./internal/oracle']),
]
results=[]
for name,args in jobs:
 start=time.monotonic()
 with (out/(name+'-final.log')).open('w') as log:
  try: result=subprocess.run(args,stdout=log,stderr=subprocess.STDOUT,timeout=180); code=result.returncode
  except subprocess.TimeoutExpired: code=124
 row=dict(name=name,command=args,exit=code,seconds=round(time.monotonic()-start,3));results.append(row)
 (out/'validation.json').write_text(json.dumps(results,indent=2)+'\n')
 print(json.dumps(row),flush=True)
 if code: break
