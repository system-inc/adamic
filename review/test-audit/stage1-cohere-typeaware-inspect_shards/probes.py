import pathlib,subprocess,json,difflib,time,os
repo=pathlib.Path('/workspace/adamic');p=repo/'review/test-audit/stage1-cohere-typeaware-inspect_shards';tmp=pathlib.Path('/tmp/u145');base='stage1/cohere/typeaware/';runs=[]
for id,file,pattern in [('P2',base+'testdata/fact_cost.ts','^(TestInspectRequestRefusals|TestSixRuleAgreementAndMutants_033)$'),('P3',base+'testdata/sharded_suite.ts','^TestSixRuleAgreementAndMutants_000$')]:
 original=(repo/file).read_text();empty='export {};\n';(p/(id+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),empty.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 try:
  (repo/file).write_text(empty);cmd=['timeout','90',str(tmp/'adamic'),'build',file,'-o',str(tmp/id),'--tsgo',str(tmp/'checker.a')];started=time.monotonic()
  with (p/(id+'-build.log')).open('w') as f:r=subprocess.run(cmd,cwd=repo,stdout=f,stderr=subprocess.STDOUT)
  if r.returncode: raise RuntimeError('probe build '+id)
  build_seconds=time.monotonic()-started;cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/typeaware/','-run',pattern];started=time.monotonic()
  with (p/(id+'.log')).open('w') as f:r=subprocess.run(cmd,cwd=repo,env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE='/tmp/u145/typescript',ADAMIC_TYPEAWARE_BENCH='1'),stdout=f,stderr=subprocess.STDOUT)
  runs.append({'id':id,'file':file,'line':1,'change':'empty top-level native entry, export {}','command':cmd,'build_seconds':build_seconds,'seconds':time.monotonic()-started,'exit':r.returncode});(p/'entry-probe-runs.json').write_text(json.dumps(runs,indent=2))
 finally:(repo/file).write_text(original)
print('entry probes complete')
