import json,os,re,subprocess,time
from pathlib import Path
out=Path(__file__).resolve().parent
results=[]
def run(label,command,expected=0,minimum=1):
 start=time.monotonic();log=out/(label+'.log')
 with log.open('w') as stream:
  completed=subprocess.run(command,stdout=stream,stderr=subprocess.STDOUT,timeout=240)
 result=dict(label=label,command=command,exit=completed.returncode,seconds=round(time.monotonic()-start,3))
 results.append(result);(out/'checks.json').write_text(json.dumps(results,indent=2))
 print(label,result['exit'],result['seconds'],flush=True)
 text=log.read_text();events=[]
 for line in text.splitlines():
  if line.startswith('{'):
   try:events.append(json.loads(line))
   except json.JSONDecodeError:pass
 tests=[r for r in events if r.get('Action') in ['pass','fail'] and 'Test' in r]
 assert completed.returncode==expected,(label,completed.returncode)
 assert len(tests)>=minimum,(label,'matched no tests')
 assert not any(r.get('Elapsed',0)>=60 for r in tests),(label,'test crossed 60 seconds')
 if expected==1:
  assert '[build failed]' not in text and any(r['Action']=='fail' for r in tests),(label,'not a test catcher')
 return tests
standard=['-count=1','-timeout','90s','-parallel','4','-json']
paths='internal/oracle/testdata/(scout_|taste_stage3_representations)'
run('counts-regeneration',['go','test','./internal/oracle','-run','^TestCountsAreRecorded_[0-9]+$/'+paths,*standard,'-args','-update-counts'])
run('counts-table',['go','test','./internal/oracle','-run','^TestCountsAreRecordedTable$',*standard])
run('lower-focus',['go','test','./internal/lower','-run','^Test(NullableStringAdmission|ScoutClassBooleanKeepsUnsafeViewsRefused|NullableObjectSharedStorageStaysNotYet|NullableObjectUnionKeepsStrongCyclesRefused|TasteBoxedScalarFieldNowLowers|TasteRepresentationLimitsStayExplicit)$',*standard])
run('oracle-focus',['go','test','./internal/oracle','-run','^Test(NullableStringFieldCheckMutant|NullableObjectStorage.*)$',*standard])
run('fixtures',['go','test','./internal/oracle','-run','^TestNativeAgreesWithNode_[0-9]+$/'+paths,*standard])
run('regex-adapter',['go','test','./internal/oracle','-run','^TestNativeAgreesWithNode_[0-9]+$/internal/oracle/testdata/regexp_replace/callback.a$',*standard])
run('stage1-gaps',['go','test','./stage1/cohere/selector','./stage1/cohere/values','-run','^TestEachGapStandsWhereGapsMdSaysItDoes$/gaps/(3_optional_boolean|2_boolean_or_undefined_field)',*standard])
run('flow-consumers',['go','test','./internal/flow','-run','^TestFlowCorpusRemainder$',*standard])
run('fresh-consumers',['go','test','./internal/fresh','-run','^TestFreshCorpusRemainder$',*standard])
run('native-consumers',['go','test','./internal/native','-run','^TestInheritanceMemoryPlans$',*standard])
run('guards-ir',['go','test','./internal/ir','-run','^TestEveryMapAProgramHoldsIsSorted$',*standard])
run('guards-runtime',['go','test','./internal/oracle','-run','^TestRuntimeSetsAreEachUsed$',*standard])
run('guards-registration',['go','test','./internal/lower','-run','^TestNativeAgreementRowsNameRegisteredPrograms$',*standard])
run('guards-parallel',['go','test','./cmd/adamic-gate','-run','^TestEveryTestIsParallelOrSaysWhy$',*standard])
run('readers',['go','test','./internal/ir','-run','^TestCallTargetReaders$',*standard])
run('vet',['go','vet','./internal/lower','./internal/native','./internal/oracle','./stage1/cohere/selector','./stage1/cohere/values'],minimum=0)
for c in json.loads((out/'mutants/cases.json').read_text()):
 run('mutant-'+c['topic'],['go','test','-overlay',c['overlay'],*c['packages'],'-run',c['pattern'],*standard],1)
