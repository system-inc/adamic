import pathlib,subprocess,json,time,os,difflib,re
R=pathlib.Path('/workspace/adamic');E=R/'review/test-audit/cmd-adamic-test262-corpus';P=R/'cmd/adamic-test262';base=json.load(open(E/'base.json'));meta=json.load(open(E/'function-offsets.json'))
probes=[('P01','classify.go','classify','classified{}',['TestClassifyCorpus']),('P02','verdict.go','decide','verdict{}',['TestVerdictCorpus']),('P03','reason.go','compileClass','"", ""',['TestNormalizeReason']),('P04','rewrite.go','rewriteHarnessCalls','""',['TestRewriteHarnessCalls']),('P05','frontmatter.go','parseFrontmatter','frontmatter{}',['TestFrontmatterShapes']),('P06','run.go','runFilter','filterReport{}, nil',['TestMiniRunner','TestLargeCompilerOutputIsComplete','TestWorkerLazyFallback']),('P07','run.go','attempt','result{}',['TestEditCacheSeparation','TestLoweringSourceEdit']),('P08','run.go','Write','0, nil',['TestOutputOverflowIsReported']),('P09','cache.go','nodeHarnessSourceIdentity','""',['TestNodeHarnessIdentity']),('P10','cache.go','prepareCache','nil, "", "", nil',['TestRunnerLocationIdentity']),('P11','run.go','fallbackCompile','execution{}',['TestWorkerLazyFallback']),('P12','run.go','runCommandWithLimit','execution{}',['TestCompilerStartupMeasurement'])]
(E/'probes.json').write_text(json.dumps(probes,indent=2))
while not (E/'matrix.done').exists():time.sleep(1)
# switch all probes at entries so original import uses remain intact
for f,s in base.items():
 edits=[]
 for id,pf,name,value,rows in probes:
  if pf==f:
   fn=next(x for x in meta[f] if x['name']==name);edits.append((fn['body']+1,'\nif auditSelector()=="'+id+'" { return '+value+' }\n'))
 for pos,text in sorted(edits,reverse=True):s=s[:pos]+text+s[pos:]
 (P/f).write_text(s)
(P/'audit_selector.go').write_text('package main\nimport "os"\nfunc auditSelector() string {return os.Getenv("ADAMIC_MUTANT")}\n')
records=[]
for id,f,name,value,rows in probes:
 start=time.monotonic();env=os.environ.copy();env['ADAMIC_MUTANT']=id
 with (E/(id+'.log')).open('w') as log:rc=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./cmd/adamic-test262/','-run','^('+'|'.join(rows)+')$'],cwd=R,env=env,stdout=log,stderr=subprocess.STDOUT).returncode
 records.append(dict(id=id,seconds=time.monotonic()-start,exit=rc));(E/'probe-time.json').write_text(json.dumps(records,indent=2))
for f,s in base.items():(P/f).write_text(s)
(P/'audit_selector.go').unlink();(E/'probes.done').write_text('done')
