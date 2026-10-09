from pathlib import Path
import subprocess,os,json,time
p=Path('/tmp/defend-conditions');root=Path('/workspace/adamic')
mutants=[('D1','internal/lower/library_object.go','call.ElementName = l.checker.TypeToString(element)','call.ElementName = "unknown"',145,'Change the declared element diagnostic label to unknown; preserve the check and exit code','TestEntriesProvenance'),('D2','internal/javascript/entries_provenance.go','adamicFieldReadiness.get(object)?.has(key) ? undefined : object[key]','false ? undefined : object[key]',21,'Flip the checked enumeration readiness condition to false, reading the stored value of an uninitialized field','TestEntriesRuntimeReadiness')]
regexes=[('core','^(TestEntriesProvenance|TestEntriesRuntimeReadiness|TestEntriesAcceptance|TestFractionalPowersReachRuntime|TestReviewProgramsNoLooseFiles|TestReviewProgramsRefuse|TestReviewProgramsSelfTest)$'),('general',r'^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^(entries_.*|library_object_.*|method_coverage_object_statics\.a|notyet_library_object_entries_const\.a|enums_modules)$/^main\.a$'),('review',r'^TestReviewProgramsAgreeWithNode$/^(smoke\.a|fxspptb_oct9_native_p24_tuple_from_object_entries\.a)$'),('counts','^TestCountsAreRecorded$')]
(p/'plan.json').write_text(json.dumps({'mutants':mutants,'regexes':regexes,'code_under_test':'Adamic enumeration lowering and JavaScript/native backends','oracle':'Node on successful original sources; self-authored exact checked failure contracts'},indent=2)+'\n')
for label,_ in regexes:
 es=[json.loads(s) for s in (p/(label+'-baseline.log')).read_text().splitlines() if s.startswith('{')]
 assert any(e['Action']=='pass' and not e.get('Test') for e in es),label
commands=[]
for mid,file,before,after,line,change,target in mutants:
 f=root/file;original=f.read_text();assert original.count(before)==1
 try:
  f.write_text(original.replace(before,after));(p/(mid+'.diff')).write_text(subprocess.check_output(['git','diff','--',file],cwd=root,text=True))
  env=os.environ.copy();env['ADAMIC_GATE_UNCACHED']='1';env['ADAMIC_BUILD_CACHE_DIR']=str(p/'cache'/mid)
  jobs=[('vet',['go','vet','./'+str(Path(file).parent)+'/'])]+[(label,['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',regex]) for label,regex in regexes]
  for label,cmd in jobs:
   start=time.monotonic()
   with open(p/(mid+'.'+label+'.log'),'w') as out:r=subprocess.run(cmd,cwd=root,env=env,stdout=out,stderr=subprocess.STDOUT)
   item={'mutant':mid,'label':label,'command':cmd,'cache':env['ADAMIC_BUILD_CACHE_DIR'],'exit':r.returncode,'wall':time.monotonic()-start};commands.append(item);(p/'commands.json').write_text(json.dumps(commands,indent=2)+'\n');print(item,flush=True)
   if label=='vet' and r.returncode:raise RuntimeError('vet failed')
 finally:f.write_text(original)
