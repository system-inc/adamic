import subprocess,pathlib,json,time,sys
out=pathlib.Path(__file__).resolve().parent
root=out.parents[2]
cases=[
('./internal/fresh','^Test(Regex|NodeFSFile|FutureRegex|MethodKeepsArgument)'),
('./internal/javascript','^TestViewMixedUnionSelectsNumber$'),
('./internal/regexp','^Test(CaptureFactsCompilerTable|MatcherNodeControls|Parse)$'),
('./internal/leakcheck','^Test(ReportedLeakIsRejected|CountedBalance)$'),
('./internal/native','^Test(ParallelMapCompilerABI|PrototypeMixedCountedCycleProbe|PrototypeNodeResearch|PrototypeDebugSiteRefusals)$'),
('./internal/native','^TestParallel(Memory|Lifecycle|Map_numbers|Map_strings|Map_objects|Map_map|Map_fresh|Map_nested|Map_exception|Map_nested_exception|Map_million|Mutant_field_cache|Mutant_oversized_slot)$'),
('./internal/oracle','^Test(RuntimeSetsAreEachUsed|NativeAgreementRowsNameRegisteredPrograms|FamiliesOwnEveryCaseOnce|FamilyWrappersAreCurrent|AsyncFixturesAgreeWithNode|AsyncThrowReleasesFrame|ParallelNative_numbers|ParallelNative_readonly_forms)$'),
('./internal/compilerinputs','^Test(LoweringIsWhatTheCompilerLinks|LoweringIsACopy|CheckKeyTellsTheCompilerFromTscWhole)$'),
('./internal/fuzz','^Test(UndefinedNumbersShapes|JudgeReadsThePanicLine)$'),
('./stage1/typescript/parser','^Test(StrongAstParentGap|PushSpreadGap|ClassMethodInterfaceGap|ClosedOptionalFunctionValueGap|ClosedConditionalEmptyArrayGap)$'),
('./stage1/typescript/scanner','^Test(GapStandsWhereGapsMdSays|BigintGapStandsWhereGapsMdSays)$'),
('./stage3/fixtures','^Test(FixturesCycles|FixturesNamespaces|FixturePaths|FixtureShardManifest)$'),
('./stage3/census/latent/statecopy','^Test(UnknownMutableContainerFailsLoudly|UnknownForeignPointerFailsLoudly|PublishedCopyModeIsPrivate)$'),
]
results=[]
for i,(pkg,selector) in enumerate(cases):
 cmd=['timeout','120','go','test',pkg,'-run',selector,'-count=1','-parallel=4','-json','-timeout=90s']
 label='readers-'+str(i); start=time.monotonic()
 with (out/(label+'.log')).open('w') as log:r=subprocess.run(cmd,cwd=root,stdout=log,stderr=subprocess.STDOUT)
 tests=[]
 for line in (out/(label+'.log')).read_text().splitlines():
  try:e=json.loads(line)
  except:continue
  if e.get('Test') and e.get('Action') in ('pass','fail','skip'):tests.append({'test':e['Test'],'action':e['Action'],'seconds':e.get('Elapsed',0)})
 results.append({'command':cmd,'exit':r.returncode,'seconds':round(time.monotonic()-start,3),'tests':tests})
 (out/'readers-results.json').write_text(json.dumps(results,indent=2)+'\n')
 print(pkg,r.returncode,results[-1]['seconds'],flush=True)
 if r.returncode or any(t['seconds']>=60 for t in tests):sys.exit(1)
