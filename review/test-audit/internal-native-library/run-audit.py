import pathlib,json,re,subprocess,time,difflib,os,sys
root=pathlib.Path('/workspace/adamic');p=root/'review/test-audit/internal-native-library';src=root/'internal/native';rows=json.loads((p/'rows.json').read_text());production=[r for r in rows if r!='TestMapHashProbeCatchesMutants'];pattern='^('+ '|'.join(production)+')$'
def run(cmd,name,env=None):
 t=time.monotonic()
 with (p/name).open('w') as f:r=subprocess.run(cmd,cwd=root,stdout=f,stderr=subprocess.STDOUT,env=env)
 return {'exit':r.returncode,'wall':time.monotonic()-t,'command':' '.join(cmd)}
def replace_scope(s,old,new,scope):
 start=s.index(scope) if scope else 0;at=s.index(old,start);return s[:at]+new+s[at+len(old):]
if sys.argv[1]=='plan':
 specs=[
 ('library.go','part(version)','','if !auditMutant("M01") { part(version) }','func runtimeKey'),
 ('library.go','part(compiler)','','if !auditMutant("M02") { part(compiler) }','func runtimeKey'),
 ('library.go','for _, flag := range flags {\n\t\tpart(flag)\n\t}','','if !auditMutant("M03") { for _, flag := range flags { part(flag) } }','func runtimeKey'),
 ('library.go','fmt.Fprintf(hash, "%d:", len(value));','','if !auditMutant("M04") { fmt.Fprintf(hash, "%d:", len(value)) };','func runtimeKey'),
 ('library.go','hash.Write(file.contents)','','if !auditMutant("M05") { hash.Write(file.contents) }','func runtimeKey'),
 ('library.go','!strings.HasSuffix(entry.Name(), ".h")','!strings.HasSuffix(entry.Name(), ".hh")','auditValue("M06", !strings.HasSuffix(entry.Name(), ".h"), !strings.HasSuffix(entry.Name(), ".hh"))','func readRuntime'),
 ('library.go','info.Mode().IsRegular()','!info.Mode().IsRegular()','auditValue("M07", info.Mode().IsRegular(), !info.Mode().IsRegular())','func cachedRuntime'),
 ('library.go','os.Rename(temporary, directory)','os.Rename(directory, temporary)','os.Rename(auditValue("M08", temporary, directory), auditValue("M08", directory, temporary))','func cachedRuntime'),
 ('library.go','"--whole-archive"','"--no-whole-archive"','auditValue("M09", "--whole-archive", "--no-whole-archive")','func RuntimeLinkFlags'),
 ('element_borrow.go','program.Locals[loop.Local].Borrowed = true','program.Locals[loop.Local].Borrowed = false','program.Locals[loop.Local].Borrowed = !auditMutant("M10")','func planElementBorrows'),
 ('element_borrow.go','lending[loop.Iterable.(ir.Read).Local] = true','lending[loop.Iterable.(ir.Read).Local] = false','lending[loop.Iterable.(ir.Read).Local] = !auditMutant("M11")','func planElementBorrows'),
 ('element_borrow.go','!changes(program, changing, loop.Body)','changes(program, changing, loop.Body)','auditValue("M12", !changes(program, changing, loop.Body), changes(program, changing, loop.Body))','func planElementBorrows'),
 ('element_borrow.go','return false','return true','return auditMutant("M13")','case ir.Call:\n'),
 ('element_borrow.go','!held.Captured &&','','auditValue("M14", !held.Captured, true) &&','func borrowable'),
 ('borrow.go','!pure(argument)','pure(argument)','auditValue("M15", !pure(argument), pure(argument))','func (e *emitter) lentArgument'),
 ('emit_statements.go','!e.elementBorrows[e.at]','e.elementBorrows[e.at]','auditValue("M16", !e.elementBorrows[e.at], e.elementBorrows[e.at])','func (e *emitter) forOf'),
 ('emit_statements.go','else if e.elementBorrows[e.at]','else if !e.elementBorrows[e.at]','else if auditValue("M17", e.elementBorrows[e.at], !e.elementBorrows[e.at])','func (e *emitter) forOf'),
 ('runtime/string_build_impl.h','units > ADAMIC_STRING_MAX_UNITS','units >= ADAMIC_STRING_MAX_UNITS','(getenv("ADAMIC_MUTANT") != NULL && strcmp(getenv("ADAMIC_MUTANT"), "M18") == 0 ? units >= ADAMIC_STRING_MAX_UNITS : units > ADAMIC_STRING_MAX_UNITS)','void adamic_string_check_length'),
 ('runtime/string_build_impl.h','units > ADAMIC_STRING_MAX_UNITS','units > ADAMIC_STRING_MAX_UNITS + 1','(getenv("ADAMIC_MUTANT") != NULL && strcmp(getenv("ADAMIC_MUTANT"), "M19") == 0 ? units > ADAMIC_STRING_MAX_UNITS + 1 : units > ADAMIC_STRING_MAX_UNITS)','void adamic_string_check_length'),
 ('runtime/map_set.c','hash ^= hash >> 12','hash ^= hash >> 11','hash ^= hash >> (getenv("ADAMIC_MUTANT") != NULL && strcmp(getenv("ADAMIC_MUTANT"), "M20") == 0 ? 11 : 12)','uint64_t adamic_map_number_hash')]
 originals={f:(src/f).read_text() for f,*_ in specs};plan=[];switched=dict(originals)
 for i,(f,old,new,switch,scope) in enumerate(specs,1):
  mid=f'M{i:02}';s=originals[f];start=s.index(scope);at=s.index(old,start);mut=replace_scope(s,old,new,scope);diff=''.join(difflib.unified_diff(s.splitlines(True),mut.splitlines(True),fromfile='a/internal/native/'+f,tofile='b/internal/native/'+f));diff=re.sub(r'^\+\s+\n','+\n',diff,flags=re.M)
  (p/(mid+'.diff')).write_text(diff);plan.append({'id':mid,'file':f,'line':s[:at].count('\n')+1,'old':old,'new':new});switched[f]=replace_scope(switched[f],old,switch,scope)
 # M19's original condition is now inside M18's ternary. Replacing its FIRST occurrence would alter M18's mutated branch. Build both predicates together explicitly.
 f='runtime/string_build_impl.h';s=originals[f];old='units > ADAMIC_STRING_MAX_UNITS';new='(getenv("ADAMIC_MUTANT") != NULL && strcmp(getenv("ADAMIC_MUTANT"), "M18") == 0 ? units >= ADAMIC_STRING_MAX_UNITS : getenv("ADAMIC_MUTANT") != NULL && strcmp(getenv("ADAMIC_MUTANT"), "M19") == 0 ? units > ADAMIC_STRING_MAX_UNITS + 1 : units > ADAMIC_STRING_MAX_UNITS)';switched[f]=replace_scope(s,old,new,'void adamic_string_check_length')
 (p/'plan.json').write_text(json.dumps(plan,indent=2));(p/'switched-source.json').write_text(json.dumps(switched));(p/'matrix-rows.json').write_text(json.dumps(production,indent=2));exit()
if sys.argv[1]=='timings':
 out={}
 for row in rows:out[row]=[run(['timeout','120','go','test','-count=1','-timeout','90s','./internal/native/','-run','^'+row+'$'],row+f'-{i}.log') for i in range(3)]
 (p/'timings.json').write_text(json.dumps(out,indent=2));exit()
if sys.argv[1]=='prepare':
 checks={}
 for m in json.loads((p/'plan.json').read_text()):
  mid=m['id'];q=p/(mid+'.diff');subprocess.run(['git','apply','--check',str(q)],cwd=root,check=True);subprocess.run(['git','apply',str(q)],cwd=root,check=True)
  if m['file'].endswith('.go'):cmd=['timeout','90','go','vet','./internal/native/']
  else:cmd=['timeout','90','clang','-std=c11','-Wall','-Wextra','-Werror','-Wcast-function-type-strict','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all','-Iinternal/native/runtime','-c','internal/native/runtime/'+('string.c' if m['file'].endswith('.h') else 'map_set.c'),'-o','/tmp/u048-'+mid+'.o']
  checks[mid]=run(cmd,mid+'-compile.log');assert checks[mid]['exit']==0,(mid,checks[mid]);subprocess.run(['git','restore','--source=HEAD','--','internal/native/'+m['file']],cwd=root,check=True)
 (p/'compile-status.json').write_text(json.dumps(checks,indent=2))
 for f,s in json.loads((p/'switched-source.json').read_text()).items():(src/f).write_text(s)
 (src/'audit_selector.go').write_text('package native\nimport "os"\nfunc auditMutant(id string) bool {return os.Getenv("ADAMIC_MUTANT")==id}\nfunc auditValue[T any](id string, normal, mutant T) T {if auditMutant(id) {return mutant}; return normal}\n')
 probes=[('library.go','func runtimeKey(files []runtimeFile, flags []string, compiler string, version string) string {','return ""','P01'),('library.go','func cachedRuntime(files []runtimeFile, flags []string, compiler string, version string, cache string) (string, error) {','return "", nil','P02'),('native.go','func Build(source string, output string, options Options) error {','return nil','P03'),('element_borrow.go','func planElementBorrows(program *ir.Program) (map[*ir.Statement]bool, map[int]bool) {','return nil, nil','P04'),('emit.go','func C(program *ir.Program) string {','return ""','P05'),('borrow.go','func (e *emitter) lentArgument(call ir.Call, index int) (string, bool) {','return "", false','P06')]
 for f,sig,ret,mid in probes:
  q=src/f;s=q.read_text();q.write_text(s.replace(sig,sig+'\n if auditMutant("'+mid+'") { '+ret+' }',1))
 for f,sig,ret,mid in [('runtime/string_build_impl.h','void adamic_string_check_length(double units) {','return;','P07'),('runtime/map_set.c','uint64_t adamic_map_number_hash(double number) {','return 0;','P08')]:
  q=src/f;s=q.read_text();s=s.replace('#include <string.h>', '#include <string.h>\n#include <stdlib.h>',1) if f.endswith('map_set.c') else s;q.write_text(s.replace(sig,sig+'\n if (getenv("ADAMIC_MUTANT") != NULL && strcmp(getenv("ADAMIC_MUTANT"), "'+mid+'") == 0) { '+ret+' }',1))
 subprocess.run(['gofmt','-w']+[str(src/f) for f in ['library.go','element_borrow.go','borrow.go','emit_statements.go','native.go','emit.go','audit_selector.go']],check=True);exit()
if sys.argv[1]=='matrix':
 out=json.loads((p/'selector-build-failures/matrix-status.json').read_text())
 for mid in ['M01', 'M02', 'M03', 'M04', 'M05', 'M06', 'M07', 'M08', 'M09', 'M18', 'M19', 'M20', 'P07', 'P08']:
  env=dict(os.environ,ADAMIC_MUTANT=mid,ADAMIC_BUILD_CACHE_DIR='/workspace/u048-products/'+mid,XDG_CACHE_HOME='/workspace/u048-runtime-cache/'+mid,GOCACHE='/home/agent/.cache/go-build',TMPDIR='/workspace/u048-tmp')
  probe_rows={'P01':['TestRuntimeKeyIncludesEveryInput'],'P02':['TestRuntimeCacheRebuildsChangedSources','TestRuntimeCacheConcurrentBuilders','TestRuntimeCacheConcurrentProcesses'],'P03':['TestRuntimeCacheKeepsCountFlags'],'P04':['TestLoopBorrowPlan','TestLoopCallCoverage'],'P05':['TestNbodyBorrowedLoopC','TestLoopArrayHoldC','TestLoopCallCoverage'],'P06':['TestGlobalArgumentLending','TestLoopCallCoverage'],'P07':['TestStringLengthLimit'],'P08':['TestMapHashProbeBound']}
  groups={
   'cache':['TestRuntimeKeyIncludesEveryInput','TestRuntimeCacheKeepsCountFlags','TestRuntimeCacheRebuildsChangedSources','TestRuntimeCacheConcurrentBuilders','TestRuntimeCacheConcurrentProcesses'],
   'borrow':['TestLoopBorrowPlan','TestNbodyBorrowedLoopC','TestGlobalArgumentLending','TestLoopArrayHoldC','TestLoopCallCoverage'],
   'length':['TestStringLengthLimit'],'hash':['TestMapHashProbeBound']}
  if mid.startswith('M'):
   number=int(mid[1:]);selected_rows=groups['cache' if number<=9 else 'borrow' if number<=17 else 'length' if number<=19 else 'hash']
  else:selected_rows=probe_rows[mid]
  selected='^('+ '|'.join(selected_rows)+')$'
  out[mid]=run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run',selected],mid+'.log',env)
  out[mid]['rows']=selected_rows
  (p/'matrix-status.json').write_text(json.dumps(out,indent=2))
