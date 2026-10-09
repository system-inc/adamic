import pathlib,subprocess,os,time,json
p=pathlib.Path('review/test-audit/stage1-cohere-json-audit');names=['TestUpstreamNumericSeparatorGap','TestExternalComparisonCatchesThreePrinterMutants','TestJSONGoToolchainStable','TestNativeChunkPlan','TestNativeChunkDifferenceNamesCase','TestDocumentedStageZeroGaps','TestClosedEmptyFallbackGap','TestProfileSnapshotsAgree','TestCachedWidthMutantIsCaught','TestCorpusPinDiagnostics','TestCorpusPinRejectsMissingProvisionedInputs','TestCorpusPinBeforeSampling'];rs=[];env=os.environ.copy();env['ADAMIC_JSON_PRETTIER']='/tmp/u102/prettier'
for n in names:
 for i in range(3):
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/json/','-run','^'+n+'$'];t=time.monotonic()
  with (p/('timing-'+n+'-'+str(i)+'.log')).open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env=env)
  rs.append(dict(test=n,trial=i,command=cmd,exit=r.returncode,wall_seconds=time.monotonic()-t));pathlib.Path('/tmp/u102/timing.json').write_text(json.dumps(rs,indent=2));print(n,i,r.returncode,round(rs[-1]['wall_seconds'],2),flush=True)
  if r.returncode:raise SystemExit('Red scoped baseline; stop')
