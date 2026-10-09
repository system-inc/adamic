import pathlib,subprocess,json,os,time
root=pathlib.Path('/workspace/adamic');p=root/'review/test-defend/internal-oracle-enums_open'
mutants=[('D1','internal/native/runtime/input.c','O_WRONLY | O_CREAT | O_TRUNC | O_CLOEXEC','O_WRONLY | O_CREAT | O_CLOEXEC','drop file truncation option'),('D2','internal/lower/class_static.go','checked = !l.provenModuleReads[expression]','checked = l.provenModuleReads[expression]','flip class-base readiness proof condition')]
records=[]
rx='^(TestGateCacheNodeInputs|TestCheckedViewUntaggedOwnClassData|TestClassWrongOutput103|TestClassWrongOutput107|TestClassWrongOutput108|TestClassWrongOutput106|TestNumericEnumNeverPathsPinned|TestNumericEnumNeverPinned|TestImportCycleRuntimeCalls|TestImportCycleLoadTimeReads|TestImportedNonliteralConstCaseIsNotYet|TestInputAgreesWithNode|TestInterfaceCastImportedConstruction|TestModuleNamespaceReadsMatchNode|TestModuleNamespaceReadinessMutants|TestModuleNamespaceLiveBindingMutant|TestNamespaceLiveExportBoundary|TestNamespaceRuledMutants|TestNamespaceSemanticMutants|TestNamespaceStateMutants|TestFileWritesLandInNodesOrder|TestParserNamespaceReceiver|TestParserNamespaceClass|TestParserCallableNamespace|TestParserCallableNamespaceMutant|TestParserNamespaceClassRegistrationMutant|TestParserNamespaceReceiverMutant|TestViewFieldInheritedStaticReadiness)$'
for ident,file,old,new,why in mutants:
 target=root/file;original=target.read_text();assert original.count(old)==1;line=original[:original.index(old)].count('\n')+1
 try:
  target.write_text(original.replace(old,new))
  (p/(ident+'.diff')).write_bytes(subprocess.check_output(['git','diff','--',file],cwd=root))
  validation=['clang','-std=c11','-Wall','-Wextra','-Werror','-pedantic','-fsyntax-only','-I','internal/native/runtime',file] if file.endswith('.c') else ['go','vet','./internal/lower/']
  with (p/(ident+'-validation.log')).open('w') as log:v=subprocess.run(validation,cwd=root,stdout=log,stderr=subprocess.STDOUT)
  assert v.returncode==0,ident
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',rx];env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/oracle-defense/cache/'+ident+'-narrow';env['ADAMIC_GATE_UNCACHED']='1';t=time.monotonic()
  with (p/(ident+'-narrow.log')).open('w') as log:r=subprocess.run(cmd,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
  events=[json.loads(s) for s in (p/(ident+'-narrow.log')).read_text().splitlines() if s.startswith('{')];failed=sorted({e['Test'] for e in events if e['Action']=='fail' and 'Test' in e and '/' not in e['Test']});passed=sorted({e['Test'] for e in events if e['Action']=='pass' and 'Test' in e and '/' not in e['Test']});skipped=sorted({e['Test'] for e in events if e['Action']=='skip' and 'Test' in e and '/' not in e['Test']})
  record=dict(mutant=ident,file_line=file+':'+str(line),change=why,old=old,new=new,rows_failed=failed,rows_passed=passed,rows_skipped=skipped,exit=r.returncode,wall=time.monotonic()-t,command='ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' ADAMIC_GATE_UNCACHED=1 '+' '.join(cmd)+' > '+ident+'-narrow.log 2>&1')
  records.append(record);(p/'attempts.json').write_text(json.dumps(records,indent=2));print(ident,record['wall'],failed,flush=True)
 finally:target.write_text(original)
