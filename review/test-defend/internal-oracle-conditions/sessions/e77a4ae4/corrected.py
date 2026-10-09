from pathlib import Path
import subprocess,os,json,time
p=Path('/tmp/defend-conditions');root=Path('/workspace/adamic');commands=json.loads((p/'commands.json').read_text())
regexes=[('general','^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^(entries_.*|library_object_.*|method_coverage_object_statics[.]a|notyet_library_object_entries_const[.]a|enums_modules)$/^main[.]a$'),('review','^TestReviewProgramsAgreeWithNode$/^(smoke[.]a|fxspptb_oct9_native_p24_tuple_from_object_entries[.]a)$')]
for mid in ['D1','D2']:
 subprocess.run(['git','apply',str(p/(mid+'.diff'))],cwd=root,check=True)
 try:
  env=os.environ.copy();env['ADAMIC_GATE_UNCACHED']='1';env['ADAMIC_BUILD_CACHE_DIR']=str(p/'cache'/mid)
  for label,regex in regexes:
   (p/(mid+'.'+label+'.log')).rename(p/(mid+'.'+label+'-prior-'+str(time.time_ns())+'.log'))
   cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',regex];start=time.monotonic()
   with open(p/(mid+'.'+label+'.log'),'w') as out:r=subprocess.run(cmd,cwd=root,env=env,stdout=out,stderr=subprocess.STDOUT)
   es=[json.loads(l) for l in (p/(mid+'.'+label+'.log')).read_text().splitlines() if l.startswith('{')];passed=[e['Test'] for e in es if e['Action'] in ['pass','skip'] and '/' in e.get('Test','')]
   assert len(passed)==(25 if label=='general' else 2),(mid,label,len(passed))
   item={'mutant':mid,'label':label+'-corrected','command':cmd,'cache':env['ADAMIC_BUILD_CACHE_DIR'],'exit':r.returncode,'wall':time.monotonic()-start};commands.append(item);(p/'commands.json').write_text(json.dumps(commands,indent=2)+'\n');print(item,flush=True)
 finally:subprocess.run(['git','apply','-R',str(p/(mid+'.diff'))],cwd=root,check=True)
