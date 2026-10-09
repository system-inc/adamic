import pathlib,json,subprocess,os,time,difflib,re
root=pathlib.Path('/workspace/adamic');out=root/'review/test-audit/internal-native-arguments_length';plans=json.loads((out/'plan.json').read_text());names=json.loads(pathlib.Path('/tmp/u045/names.json').read_text());witnesses=['TestClosureConventionDropCount','TestClosureConventionRuntimeDropCount','TestClosureConventionWrongOrder','TestOptionalMethodThunksMatchNode'];originals={m['file']:(root/m['file']).read_text() for m in plans};originals['internal/native/closure_convention_test.go']=(root/'internal/native/closure_convention_test.go').read_text();results=[]
def call(cmd,log,env=None):
 start=time.monotonic()
 with (out/log).open('w') as f:r=subprocess.run(cmd,cwd=root,stdout=f,stderr=subprocess.STDOUT,env=env)
 return dict(code=r.returncode,wall=time.monotonic()-start)
def diff(id,file,old,new):
 s=originals[file];assert s.count(old)==1;(out/(id+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),s.replace(old,new).splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
# Validate each independent production mutation before selector infrastructure.
validation=[]
for m in plans:
 (root/m['file']).write_text(originals[m['file']].replace(m['old'],m['new']))
 cmd=['timeout','90','go','vet','./internal/native/'] if m['file'].endswith('.go') else ['clang','-std=c11','-Wall','-Wextra','-Werror','-pedantic','-Wno-unused-parameter','-O2','-I','internal/native/runtime','-c','internal/native/runtime/case.c','-o','/tmp/u045/'+m['id']+'.o']
 v=call(cmd,m['id']+'-compile.log');validation.append(dict(id=m['id'],**v));(root/m['file']).write_text(originals[m['file']]);(out/'validation.json').write_text(json.dumps(validation,indent=2))
 if v['code']:print('VALIDATION FAILED',m['id'],flush=True);raise SystemExit(1)
# Entry probes. No oracle or preparation-helper edits.
probes=[('P_C','internal/native/emit.go','func C(program *ir.Program) string {','return ""',['TestNoReaderCallingConvention','TestParserHasNoUnusedOptionalMethodThunks','TestClosureConventionRuntimeFeaturesIgnoreLiterals']),('P_BORROW','internal/native/element_borrow.go','func planElementBorrows(program *ir.Program) (map[*ir.Statement]bool, map[int]bool) {','return nil, nil',['TestBorrowChainDeclarations','TestInheritanceMemoryPlans']),('P_CHAIN','internal/native/borrow.go','func chainUnchanged(program *ir.Program, body []ir.Statement, names map[string]bool) bool {','return false',['TestBorrowChainTargets']),('P_CONSUMES','internal/native/borrow.go','func consumes(expression ir.Expression) bool {','return false',['TestPassThroughsAreNotConsumers']),('P_REGIONS','internal/native/region.go','func planRegions(program *ir.Program) *regionPlan {','return &regionPlan{}',['TestInheritanceMemoryPlans']),('P_REUSE','internal/native/reuse.go','func planReuse(program *ir.Program, lending map[int]bool) *reusePlan {','return &reusePlan{}',['TestInheritanceMemoryPlans']),('P_LIBRARY','internal/native/library.go','func RuntimeLibraryForSource(directory string, source string, options Options) (string, error) {','return "", nil',['TestClosureConventionRuntimeFeaturesIgnoreLiterals']),('P_BUILD','internal/native/native.go','func Build(source string, output string, options Options) error {','return nil',['TestArithmeticIsNeverFused']),('P_CASE_LOWER','internal/native/runtime/case.c','adamic_string *adamic_string_to_lower(const adamic_string *string) {','return NULL;',['TestCaseMappingMatchesNode']),('P_CASE_UPPER','internal/native/runtime/case.c','adamic_string *adamic_string_to_upper(const adamic_string *string) {','return NULL;',['TestCaseMappingMatchesNode'])]
(out/'probes.json').write_text(json.dumps(probes,indent=2))
for id,file,old,ret,rows in probes:diff(id,file,old,old+'\n if ('+('true' if file.endswith('.go') else '1')+') { '+ret+' }')
# Weakening the checks which planted-failure witnesses guard.
wold='return fmt.Errorf("native: clang failed: %w\\n%s", err, combined)'
wnew='return nil'
s=originals['internal/native/native.go'];oldblock='if combined, err := command.CombinedOutput(); err != nil {\n\t\t'+wold+'\n\t}'
diff('W_BUILD','internal/native/native.go',oldblock,'if _, err := command.CombinedOutput(); err != nil { return nil }')
diff('W_NODE','internal/native/closure_convention_test.go','if err == nil && string(output) == want {','if true {')
# Switched Go and C source, with helpers confined to scratch infrastructure.
helper='package native\nimport "os"\nfunc auditEnabled(id string) bool { return os.Getenv("ADAMIC_MUTANT")==id }\nfunc auditBool(id string, original, changed bool) bool { if auditEnabled(id) {return changed};return original }\nfunc auditString(id string, original, changed string) string { if auditEnabled(id) {return changed};return original }\n'
(root/'internal/native/audit_mutant.go').write_text(helper)
for file,s in originals.items():
 for m in [m for m in plans if m['file']==file and m['mode']!='static']:
  old,new,id=m['old'],m['new'],m['id'];mode=m['mode']
  if mode in ['bool','string']:replacement=('auditBool' if mode=='bool' else 'auditString')+'("'+id+'", '+old+', '+new+')'
  elif mode=='condition':replacement='if auditBool("'+id+'", '+old[3:-1].strip()+', '+new[3:-1].strip()+') {'
  elif mode=='statement':replacement='if auditEnabled("'+id+'") { '+new+' } else { '+old+' }'
  elif mode=='c-expression':replacement='(adamic_audit_selected("'+id+'") ? ('+new+') : ('+old+'))'
  elif id in ['M05','M06']:replacement=old.replace('return false','return auditBool("'+id+'", false, true)')
  elif id=='M09':replacement=old.replace('if targets.Unknown','if auditBool("M09", targets.Unknown, false)')
  else:raise RuntimeError(id)
  s=s.replace(old,replacement)
 for id,pfile,old,ret,rows in probes:
  if pfile==file:s=s.replace(old,old+'\n if '+('auditEnabled("'+id+'")' if file.endswith('.go') else '(adamic_audit_selected("'+id+'"))')+' { '+ret+' }')
 if file=='internal/native/native.go':s=s.replace(wold,'if auditEnabled("W_BUILD") { return nil }; '+wold)
 if file=='internal/native/closure_convention_test.go':s=s.replace('if err == nil && string(output) == want {','if auditBool("W_NODE", err == nil && string(output) == want, true) {')
 if file.endswith('case.c'):s=s.replace('#include <stdint.h>','#include <stdint.h>\n#include <stdlib.h>\nstatic bool adamic_audit_selected(const char *id) { const char *value=getenv("ADAMIC_MUTANT"); return value!=NULL && strcmp(value,id)==0; }')
 # strcmp prototype precedes selector.
 if file.endswith('case.c'):s=s.replace('#include <stdlib.h>\nstatic bool','#include <stdlib.h>\n#include <string.h>\nstatic bool')
 (root/file).write_text(s)
call(['gofmt','-w',*([f for f in originals if f.endswith('.go')]+['internal/native/audit_mutant.go'])],'gofmt.log')
(out/'switch.diff').write_text(subprocess.check_output(['git','diff','--',*originals.keys()],cwd=root,text=True)+'\n'+helper)
b=call(['timeout','90','go','test','-c','-o','/tmp/u045/native.test','./internal/native/'],'build.log');(out/'build.json').write_text(json.dumps(b));print('BUILD',b,flush=True)
if b['code']:raise SystemExit(1)
def execute(id,rows):
 env=os.environ.copy();env['ADAMIC_MUTANT']=id;env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u045/cache/'+id
 regex='^('+'|'.join(rows)+')$';res=call(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run',regex],id+'.log',env);res.update(id=id,rows=rows)
 raw=(out/(id+'.log')).read_text();res['panic']='panic: test timed out' in raw or 'panic: runtime error' in raw;res['cooked']=res['code']==124 or 'test timed out' in raw
 if res['panic'] or res['cooked']:
  for n in rows:call(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run','^'+n+'$'],id+'-'+n+'.log',env)
 results.append(res);(out/'runs.json').write_text(json.dumps(results,indent=2));print(id,res,flush=True)
for m in plans:
 if m['mode']=='static':(root/m['file']).write_text(originals[m['file']].replace(m['old'],m['new']))
 execute(m['id'],names)
 if m['mode']=='static':(root/m['file']).write_text(originals[m['file']])
execute('W_BUILD',witnesses[:3]);execute('W_NODE',[witnesses[3]])
for id,file,old,ret,rows in probes:execute(id,rows)
for file,s in originals.items():(root/file).write_text(s)
(root/'internal/native/audit_mutant.go').unlink()
print('RESTORED',flush=True)
