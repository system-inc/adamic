import pathlib,subprocess,json,time,os
p=pathlib.Path('review/test-audit/internal-native-arguments_length')
names='TestNoReaderCallingConvention TestBorrowChainDeclarations TestBorrowChainTargets TestPassThroughsAreNotConsumers TestCaseMappingMatchesNode TestCaseTablesMatchNodesUnicode TestInheritanceMemoryPlans TestClosureConventionDropCount TestClosureConventionRuntimeDropCount TestClosureConventionRuntimeFeaturesIgnoreLiterals TestClosureConventionWrongOrder TestParserHasNoUnusedOptionalMethodThunks TestOptionalMethodThunksMatchNode TestArithmeticIsNeverFused'.split()
(p/'names.json').write_text(json.dumps(names))
pat='^('+ '|'.join(names)+')$'
with open(p/'bounded-baseline.log','w') as f:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverpkg=./internal/native','-coverprofile='+str(p/'coverage.out'),'./internal/native/','-run',pat],stdout=f,stderr=subprocess.STDOUT)
print('bounded baseline',r.returncode,flush=True)
if r.returncode:raise SystemExit(1)
subprocess.run(['go','tool','cover','-func='+str(p/'coverage.out')],stdout=open(p/'reached-functions.txt','w'))
for name in names:
 for i in range(3):
  with open(p/f'timing-{name}-{i}.log','w') as f:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run','^'+name+'$'],stdout=f,stderr=subprocess.STDOUT)
  print(name,i,r.returncode,flush=True)
