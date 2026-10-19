import concurrent.futures,json,re,subprocess,time
from pathlib import Path
root=Path.cwd();e=root/'review/compiler/loom-r-unions';consumers=json.loads((e/'consumers.json').read_text())
selected={
'bridge/tsgo':'TestBridgeUnitsCoverEveryPiece','cloud/reports/decode-ascii':'TestRuntimeFeaturesMatchNode','cmd/adamic':'TestUsageExitStatus','cmd/adamic-meter':'TestOptionalPositionCountsUTF16Surrogates','cmd/adamic-test262':'TestAdaptClassKeepsVars','internal/compilerinputs':'TestLoweringIsWhatTheCompilerLinks','internal/flow':'TestFlowCorpusRemainder','internal/fresh':'TestFreshCorpusRemainder','internal/fuzz':'TestOneSeedOneProgram','internal/ir':'TestCallTargetReaders','internal/leakcheck':'TestCountedBalance','internal/lower':'TestNullableStringAdmission','internal/lower/cycleledger':'TestOriginalCycleLedger','internal/native':'TestNoReaderCallingConvention','internal/native/runtimeproducts':'TestTheRuntimeBuildReachesOnlyItsDeclaredFiles','internal/oracle':'TestNullableStringFieldCheckMutant','stage1/cohere/css':'TestEachGapStandsWhereGapsMdSaysItDoes','stage1/cohere/cssnumbers':'TestCSSNumbersCompilerProductsKeyWhatTheCompilerLinks','stage1/cohere/cssstrings':'TestMultiPushGap','stage1/cohere/estree':'TestAcceptanceGrammar','stage1/cohere/formatfiles':'TestFormatfilesCompilerProductsKeyWhatTheCompilerLinks','stage1/cohere/gitignore':'TestEachGapStandsWhereGapsMdSaysItDoes','stage1/cohere/graphql':'TestEachGapStandsWhereGapsMdSaysItDoes','stage1/cohere/graphql/printer':'TestPrinterConstructorGap','stage1/cohere/json':'TestDocumentedStageZeroGaps','stage1/cohere/lint':'TestClosedComparatorGaps','stage1/cohere/lint/helpers':'TestKnownGapsAreExplicit','stage1/cohere/lint/helpers/comments':'TestJsxParserGapIsExplicit','stage1/cohere/lint/regex':'TestDynamicPatternGap','stage1/cohere/markdownblocks':'TestArrayGrowthWitness','stage1/cohere/markdowninline':'TestDelimiterExpressionMatchesNode','stage1/cohere/mediaquery':'TestEachGapStandsWhereGapsMdSaysItDoes','stage1/cohere/selector':'TestEachGapStandsWhereGapsMdSaysItDoes','stage1/cohere/static_single_assignment':'TestEachGapStandsWhereGapsMdSaysItDoes','stage1/cohere/suppression':'TestEachGapStandsWhereGapsMdSaysItDoes','stage1/cohere/tsprinter':'TestClassInterfaceMethodGap','stage1/cohere/typeaware':'TestInspectRequestRefusals','stage1/cohere/values':'TestEachGapStandsWhereGapsMdSaysItDoes','stage1/cohere/yaml':'TestBundledParserDifference','stage1/typescript/parser':'TestStrongAstParentGap','stage1/typescript/scanner':'TestGapStandsWhereGapsMdSays','stage3/fixtures':'TestFixturesEnums'}
for c in consumers:
 if '/lint/rules/' in c['package']:selected[c['package']]='TestCompileProfiles'
(e/'selected-consumers.json').write_text(json.dumps(selected,indent=2)+'\n')
def run(c):
 package=c['package'];name=package.replace('/','-');binary=e/(name+'.test');cmd=['go','test','-c','-o',str(binary),'./'+package];start=time.monotonic()
 with (e/(name+'-build.log')).open('w') as log:
  try:code=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,timeout=180).returncode
  except subprocess.TimeoutExpired:code=124
 build=dict(package=package,kind='build',command=cmd,seconds=round(time.monotonic()-start,3),exit=code);rows=[build]
 if code==0 and package in selected:
  # Focused driver already executes these exact guards; compile the package here without repeating them.
  already={'internal/flow','internal/fresh','internal/ir','internal/lower','internal/oracle'}
  if package not in already:
   test=selected[package];start=time.monotonic();cmd=[str(binary),'-test.run','^'+test+'$','-test.timeout=60s','-test.parallel=2','-test.v'];logpath=e/(name+'-consumer.log')
   with logpath.open('w') as log:
    try:code=subprocess.run(cmd,cwd=root/package,stdout=log,stderr=subprocess.STDOUT,timeout=65).returncode
    except subprocess.TimeoutExpired:code=124
   text=logpath.read_text();events=[]
   for state,n,seconds in re.findall(r'^\s*--- (PASS|FAIL|SKIP): (\S+) \(([0-9.]+)s\)',text,re.M):events.append(dict(test=n,status=state,seconds=float(seconds)))
   rows.append(dict(package=package,kind='test',test=test,command=cmd,seconds=round(time.monotonic()-start,3),exit=code,events=events))
 if binary.exists():binary.unlink()
 for row in rows:
  with (e/'consumer-results.jsonl').open('a') as f:f.write(json.dumps(row)+'\n')
 print(package,[(r['kind'],r['exit'],r['seconds']) for r in rows],flush=True)
 return rows
# Resume only failed or not-yet-recorded packages after storage/setup recovery.
previous=[]
if (e/'consumer-results.jsonl').exists():
 for line in (e/'consumer-results.jsonl').read_text().splitlines():
  try:previous.append(json.loads(line))
  except json.JSONDecodeError:pass
latest={(r['package'],r['kind']):r for r in previous}
completed=set()
already={'internal/flow','internal/fresh','internal/ir','internal/lower','internal/oracle'}
for c in consumers:
 package=c['package'];build=latest.get((package,'build'));test=latest.get((package,'test'))
 if build and build['exit']==0 and (package not in selected or package in already or test and test['exit']==0):completed.add(package)
retry=e/'estree-retry.json'
if retry.exists() and any(r['label']=='estree-consumer-retry' and r['exit']==0 for r in json.loads(retry.read_text())):completed.add('stage1/cohere/estree')
consumers=[c for c in consumers if c['package'] not in completed]
for c in consumers:
 for suffix in ['-build.log','-consumer.log']:
  p=e/(c['package'].replace('/','-')+suffix)
  if p.exists() and not p.with_suffix('.first.log').exists():p.rename(p.with_suffix('.first.log'))
print('remaining packages',len(consumers),flush=True)
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:rows=list(pool.map(run,consumers))
