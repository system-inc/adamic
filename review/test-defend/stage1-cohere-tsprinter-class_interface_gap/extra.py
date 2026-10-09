import pathlib,subprocess,os,json,time
root=pathlib.Path('/workspace/adamic');out=root/'review/test-defend/stage1-cohere-tsprinter-class_interface_gap'
plan=json.load(open(out/'aimed-plan.json'))
pattern='^(TestClassInterfaceMethodGap|TestClosedOptionalBooleanFunctionGap|TestTrackedCorpusRoots|TestExpressionOwnershipStable|TestExpressionsShardPlantedDisagreement|TestTSPrinterCachedMetadataRelocation|TestExpressionUnitBoxSelection|TestExpressionUnitPlantedDisagreement|TestExpressionUnitUnionRejectsMissingAndRepeated|TestCorpusShardAssignment|TestCorpusShardCoverage|TestCorpusShardTransport|TestStatementsShardAssignmentStable|TestStatementsShardSelection|TestStatementsShardDisagreement|TestStatementsShardUnionRejectsMissingAndRepeatedIDs|TestTSCShardPlantedDisagreement|TestTSCShardUnionRejectsMissingAndRepeated)$'
for ident,file,line,old,new in plan:
 p=root/file;original=p.read_text();assert original.count(old)==1;p.write_text(original.replace(old,new))
 env=os.environ.copy();env.update(ADAMIC_BUILD_CACHE_DIR='/tmp/tsprinter-defense/cache/'+ident,ADAMIC_TYPESCRIPT_SOURCE='/tmp/u156-typescript',ADAMIC_TS_PRETTIER='/tmp/tsprinter-defense/prettier/node_modules/prettier')
 try:
  groups=[('extra',pattern)]
  if ident=='E1':groups.append(('expressions037','^TestExpressionsAgainstGoAndPrettier_037$'))
  for group,pat in groups:
   cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/tsprinter/','-run',pat];start=time.time()
   with (out/(ident+'-'+group+'.log')).open('w') as f:r=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
   with (out/'commands.jsonl').open('a') as f:f.write(json.dumps(dict(mutant=ident,group=group,command=cmd,seconds=time.time()-start,status=r.returncode,cache=env['ADAMIC_BUILD_CACHE_DIR']))+'\n')
   print(ident,group,r.returncode,round(time.time()-start,2),flush=True)
 finally:p.write_text(original)
