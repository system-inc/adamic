from pathlib import Path
import subprocess,json,difflib,time,os
p=Path('review/test-audit/stage1-cohere-lint-profile_compilation_main');root=Path.cwd();tmp=Path('/workspace/u111-tmp/checks');tmp.mkdir(exist_ok=True);base=lambda f:subprocess.check_output(['git','show','HEAD:'+f],text=True)
plans=[]
def add(mid,f,old,new,kind,rows):
 a=base(f);assert a.count(old)==1,(mid,a.count(old));b=a.replace(old,new);d=''.join(difflib.unified_diff(a.splitlines(True),b.splitlines(True),fromfile='a/'+f,tofile='b/'+f));(p/(mid+'.diff')).write_text(d);plans.append(dict(id=mid,file=f,line=a[:a.index(old)].count('\n')+1,change=new.strip() or 'drop statement',before=old,after=new,kind=kind,rows=rows));return b
f='stage1/cohere/lint/lint_test.go';a=base(f);start=a.index('func difference(');end=a.index('\nfunc manifest(',start);add('W01',f,a[start:end],'func difference(got, want []byte) string { return "" }\n','weaken agreement comparison',['TestProfileCompilationPlantedFailure','TestCommentFoldMutant','TestPositionIndexMutant'])
add('W02','stage1/cohere/lint/registration_test.go','if bytes.Equal(side.run.output, want) {','if true {','weaken equality comparison',['TestRegistrationMutant'])
# W02 removes want's only remaining use; preserve an actual use in log evidence instead of an unused Go binding.
f='stage1/cohere/lint/registration_test.go';a=base(f);b=a.replace('if bytes.Equal(side.run.output, want) {','if bytes.Equal(want, want) {');(p/'W02.diff').write_text(''.join(difflib.unified_diff(a.splitlines(True),b.splitlines(True),fromfile='a/'+f,tofile='b/'+f)));plans[-1]['after']='if bytes.Equal(want, want) {';plans[-1]['change']='equality becomes reflexive true'
add('S01','stage1/cohere/lint/profile_compilation_main_test.go','const testProfileCompilationShards = 1','const testProfileCompilationShards = 2','change construction constant',['TestProfileCompilationUnion','TestProfileCompilation_000'])
add('S02','stage1/cohere/lint/factory_hooks_shards_test.go','for index := 0; index < 2; index++ {','for index := 0; index < 1; index++ {','off-by-one construction bound',['TestFactoryHooks_Setup'])
add('S03','stage1/cohere/lint/profile_compilation_main_test.go','os.Create(filepath.Join(dir, "program.gob"))','os.Create(filepath.Join(dir, "lost.gob"))','change construction filename',['TestProfileCompilationBuildLower','TestProfileCompilationBuildC','TestProfileCompilationBuildJavaScript','TestProfileCompilationBuildNative','TestProfileCompilation_Setup'])
add('S04','stage1/cohere/lint/profile_compilation_main_test.go','native.Build(string(source), filepath.Join(dir, kind), options)','native.Build(string(source), filepath.Join(dir, kind+"-lost"), options)','change construction filename',['TestProfileCompilationBuildNative','TestProfileCompilation_Setup'])
f='stage1/cohere/lint/profile_test.go';a=base(f);start=a.index('func buildProfile(');end=a.index('\n// Exercise',start);b=a[:start]+'func buildProfile(t *testing.T, directory string) { return }\n'+a[end:]
for imp in ['context','github.com/system-inc/adamic/internal/load','github.com/system-inc/adamic/internal/lower','github.com/system-inc/adamic/internal/native']:b=b.replace('\n\t"'+imp+'"','')
(p/'P02.diff').write_text(''.join(difflib.unified_diff(a.splitlines(True),b.splitlines(True),fromfile='a/'+f,tofile='b/'+f)));plans.append(dict(id='P02',file=f,line=a[:start].count('\n')+1,kind='empty construction probe',rows=['TestProfileArtifacts'],change='buildProfile returns at entry'))
(p/'check-plan.json').write_text(json.dumps(plans,indent=2));(p/'W03-not-run.txt').write_text('NestedOutsideModuleCopy embeds bad-input checks directly against Node module loading and load.Load. No independent agreement helper guards this witness. Making its inline assertions always fail would be tautological; weakening Node or load.Load would mutate the oracle for the copy harness. Not attempted; cannot judge witness strength.\n')
results=[]
# Execution is deliberately a separate command after production switch restoration.
if os.environ.get('U111_CHECKS_RUN')!='1':raise SystemExit()
for c in plans:
 mid=c['id'];d=tmp/mid;d.mkdir(exist_ok=True);f=c['file'];q=d/f;q.parent.mkdir(parents=True,exist_ok=True);q.write_text(base(f));r=subprocess.run(['git','apply','--unsafe-paths','--directory='+str(d),str((p/(mid+'.diff')).resolve())],capture_output=True,text=True);assert r.returncode==0,r.stderr
 overlay=d/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(root/f):str(q)}}));cmd=['timeout','90','go','vet','-overlay',str(overlay),'./stage1/cohere/lint/'];st=time.monotonic()
 with (p/(mid+'-verify.log')).open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
 res=dict(id=mid,vet_command=cmd,vet_status=r.returncode,vet_wall=time.monotonic()-st)
 r=subprocess.run(['git','apply','--check','--cached',str(p/(mid+'.diff'))],capture_output=True,text=True);res['apply_status']=r.returncode
 if mid in ['S03','S04']:
  cache=Path('/workspace/u111-cache')/mid;cache.mkdir(parents=True,exist_ok=True)
  # Keep IR/emission/oracle products warm, but never seed native products or the altered Lower writer.
  for directory in Path('/home/agent/.cache/adamic-build').iterdir():
   if not directory.is_dir() or directory.name.startswith('.'):continue
   names={x.name for x in directory.iterdir()}
   if (mid=='S04' and names in [{'program.gob'},{'main.c'},{'lint.mjs'}]) or names in [{'wire','count'}]:
    target=cache/directory.name
    if not target.exists():target.symlink_to(directory)
  os.environ['ADAMIC_BUILD_CACHE_DIR']=str(cache)
 else:os.environ.pop('ADAMIC_BUILD_CACHE_DIR',None)
 if mid=='P02':os.environ['ADAMIC_LINT_PROFILE_DIR']='/workspace/u111-empty-profile'
 else:os.environ.pop('ADAMIC_LINT_PROFILE_DIR',None)
 os.environ.pop('ADAMIC_LINT_PROFILE_SNAPSHOTS',None)
 cmd=['timeout','120','go','test','-overlay',str(overlay),'-json','-count=1','-timeout','90s','./stage1/cohere/lint/','-run','^('+'|'.join(c['rows'])+')$'];st=time.monotonic()
 with (p/(mid+'.log')).open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
 es=[]
 for l in (p/(mid+'.log')).read_text().splitlines():
  try:es.append(json.loads(l))
  except:pass
 res.update(command=cmd,status=r.returncode,wall=time.monotonic()-st,seconds=next((e.get('Elapsed') for e in reversed(es) if e.get('Action') in ['pass','fail'] and 'Test'not in e),None),failed_tests=[e['Test'] for e in es if e.get('Action')=='fail' and e.get('Test') in c['rows']],cooked=any('test timed out' in e.get('Output','') for e in es),env={k:os.environ.get(k) for k in ['ADAMIC_BUILD_CACHE_DIR','ADAMIC_LINT_PROFILE_DIR']})
 if mid=='P02':res['artifacts_present']={x:Path('/workspace/u111-empty-profile',x).exists() for x in ['scanner','counted','profiled','main.c']}
 results.append(res);(p/'checks.json').write_text(json.dumps(results,indent=2))
