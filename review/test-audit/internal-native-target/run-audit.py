import pathlib,json,subprocess,time,os,difflib,re
root=pathlib.Path('/workspace/adamic');os.chdir(root);p=root/'review/test-audit/internal-native-target';raw=json.loads((p/'rows.json').read_text());ordinary=[r for r in raw if not r.startswith(('TestWASIUnit','TestSplitTSGoAgrees')) and r!='TestMeasureClangUnits'];wasi=[r for r in raw if r.startswith('TestWASIUnit')];regex='^('+'|'.join(ordinary)+')$';sdk=dict(os.environ,ADAMIC_TEST_WASI='1',WASI_SYSROOT='/tmp/u054/sdk/share/wasi-sysroot',PATH='/tmp/u054/sdk/bin:'+os.environ['PATH']);timings=[]
for name,pattern,env in [(r,'^'+r+'$',os.environ) for r in ordinary]+[('TestWASIUnit family','^TestWASIUnit[0-9]+$',sdk)]:
 for n in range(1,4):
  log=p/('timing-'+name.replace(' ','_')+'-'+str(n)+'.log');start=time.monotonic()
  with log.open('w') as f:r=subprocess.run(['timeout','120','go','test','-count=1','-timeout','90s','./internal/native/','-run',pattern],env=env,stdout=f,stderr=subprocess.STDOUT)
  timings.append(dict(test=name,run=n,status=r.returncode,wall=time.monotonic()-start,log=log.name));(p/'timings.json').write_text(json.dumps(timings,indent=2))
  if r.returncode:raise RuntimeError('clean timing '+name)
plans=[('M01','internal/native/native.go','"-O2"','"-O0"','change option'),('M02','internal/native/native.go','"-ffp-contract=off"','"-ffp-contract=fast"','change option'),('M03','internal/native/native.go','"-mno-atomics"','"-matomics"','change option'),('M04','internal/native/target.go','options.Request && options.Target != "wasm32-wasi"','options.Request && options.Target == "wasm32-wasi"','flip condition'),('M05','internal/native/target.go','if options.Sanitize {','if !options.Sanitize {','flip condition'),('M06','internal/native/target.go','if options.cpu != "" {','if options.cpu == "" {','flip condition'),('M07','internal/native/target.go','if sysroot == "" {','if sysroot != "" {','flip condition'),('M08','internal/native/units.go','names[d.name] = "adamic_unit_" + d.name','names[d.name] = "adamic_mutant_" + d.name','change constant'),('M09','internal/native/units.go','if strings.ContainsRune(" \\t\\r\\n", rune(c)) {','if !strings.ContainsRune(" \\t\\r\\n", rune(c)) {','flip condition'),('M10','internal/native/units.go','"-E", unit.name','"-EP", unit.name','change option'),('M11','internal/native/units.go','append([]string{"adamic-units-v2"}, flags...)','[]string{"adamic-units-v2"}','change option'),('M12','internal/native/units.go','state.WriteString(rewrite(tokens) + "\\n")','','drop statement'),('M13','internal/native/runtime/typed_array.c','? 256.0 : 4294967296.0','? 255.0 : 4294967296.0','change constant'),('M14','internal/native/runtime/typed_array.c','index < (double)array->length','index <= (double)array->length','off-by-one bound'),('M15','internal/native/runtime/view_unions_mixed.c','return "unsupported representation";','return "unknown";','change constant'),('M16','internal/native/runtime/view_unions_mixed.c','member->contract == 0 || match == NULL','member->contract != 0 || match == NULL','flip condition'),('M17','internal/native/runtime/weak.c','((adamic_weak *)reveal(*entry))->target = 0;','','drop statement')]
files={f for _,f,_,_,_ in plans}|{'internal/native/emit.go','internal/native/tsgo.go','internal/native/library.go','internal/native/tsgo_features_test.go'};orig={f:(root/f).read_text() for f in files};meta=[]
for id,f,old,new,menu in plans:
 assert old in orig[f],id
 meta.append(dict(id=id,file=f,line=orig[f][:orig[f].index(old)].count('\n')+1,old=old,new=new,menu=menu))
(p/'plan.json').write_text(json.dumps(meta,indent=2));(p/'diffs').mkdir(exist_ok=True)
(p/'code-and-oracle.txt').write_text('Before mutations: native options, flag construction, splitting and compilation; typed-array, mixed-union and weak-reference runtimes. Oracles: Node for typed arrays and WASI; self-written flags, split output and diagnostics; clang/UBSan runtime observations. TSGoFeatures contains a built-in missing-flags check, judged as a witness via W01. Plan frozen before outcomes.\nReached Go functions measured in coverage-functions.txt; conservative full C inventories in runtime-functions.txt.\n')
(p/'reached-functions.txt').write_text('\n'.join(x for x in (p/'coverage-functions.txt').read_text().splitlines() if not x.endswith('0.0%'))+'\n')
(p/'runtime-functions.txt').write_text('\n'.join(f+':'+str(n)+' '+m[1] for f in ['internal/native/runtime/typed_array.c','internal/native/runtime/view_unions_mixed.c','internal/native/runtime/weak.c'] for n,line in enumerate(orig[f].splitlines(),1) if (m:=re.match(r'^(?:static )?\w[^;]*?\b(\w+)\([^;]*\) \{',line)))+'\n')
flags=['-std=c11','-Wall','-Wextra','-Werror','-Wcast-function-type-strict','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all'];validation=[]
for m in meta:
 f=m['file'];new=orig[f].replace(m['old'],m['new'],1);(p/'diffs'/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(orig[f].splitlines(True),new.splitlines(True),fromfile='a/'+f,tofile='b/'+f)));(root/f).write_text(new);start=time.monotonic();cmd=['go','vet','./internal/native/'] if f.endswith('.go') else ['clang',*flags,'-I','internal/native/runtime','-c',f,'-o','/tmp/u054/'+m['id']+'.o']
 with (p/('validate-'+m['id']+'.log')).open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
 (root/f).write_text(orig[f]);validation.append(dict(id=m['id'],status=r.returncode,seconds=time.monotonic()-start,command=' '.join(cmd)));(p/'validation.json').write_text(json.dumps(validation,indent=2));assert r.returncode==0,m['id']
sw=dict(orig)
for m in meta:
 f=m['file'];old=m['old'];new=m['new'];id=m['id'];c=f.endswith('.c');sel='audit_selected' if c else 'auditSelected';cond=sel+'("'+id+'")'
 if id in ['M01','M02','M03']:rep='auditString("'+id+'", '+old+', '+new+')'
 elif id=='M04':rep='auditBool("'+id+'", '+old+', '+new+')'
 elif id in ['M05','M06','M07','M09']:rep='if auditBool("'+id+'", '+old[3:-2]+', '+new[3:-2]+') {'
 elif id=='M10':rep='auditString("M10", "-E", "-EP"), unit.name'
 elif id=='M11':rep='auditFlags(flags)'
 elif id=='M13':rep='? (audit_selected("M13") ? 255.0 : 256.0) : 4294967296.0'
 elif id=='M14':rep='(audit_selected("M14") ? index <= (double)array->length : index < (double)array->length)'
 elif id=='M15':rep='return audit_selected("M15") ? "unknown" : "unsupported representation";'
 elif id=='M16':rep='(audit_selected("M16") ? member->contract != 0 : member->contract == 0) || match == NULL'
 else:rep='if '+('('+cond+')' if c else cond)+' { '+new+' } else { '+old+' }'
 sw[f]=sw[f].replace(old,rep,1)
probes=[('PFlags','internal/native/native.go','func Flags(options Options) []string {','return nil'),('PValidate','internal/native/target.go','func ValidateOptions(options Options) error {','return nil'),('PSplit','internal/native/units.go','func splitC(source string) (string, []compilationUnit, error) {','return "", nil, nil'),('PCompile','internal/native/units.go','func compileUnit(unit compilationUnit, files []runtimeFile, flags []string, compiler, version, cache, directory string, uncached bool) (string, error) {','return "", nil'),('PTSFlags','internal/native/tsgo.go','func tsgoFlags(source string, options Options) []string {','return nil'),('PFeatureFlags','internal/native/library.go','func featureFlags(source string) []string {','return nil'),('PBuild','internal/native/native.go','func Build(source string, output string, options Options) error {','return nil'),('PC','internal/native/emit.go','func C(program *ir.Program) string {','return ""')]
for id,f,sig,ret in probes:assert sig in sw[f];sw[f]=sw[f].replace(sig,sig+'\n if auditSelected("'+id+'") { '+ret+' }',1)
typed={'new':'NULL','from_numbers':'NULL','get':'(adamic_maybe_number){false, 0}','check_write':None,'set':None,'length':'0','fill':'NULL','set_from':None,'subarray':'NULL','iterate':'NULL','iterator_next':'false'}
for name,value in typed.items():
 id='PT_'+name;f='internal/native/runtime/typed_array.c';match=re.search(r'(?m)^.*\badamic_typed_array_'+name+r'\([^\n]*\) \{',sw[f]);assert match,name;sig=match[0];ret='return'+(' '+value if value is not None else '')+';';sw[f]=sw[f].replace(sig,sig+' if (audit_selected("'+id+'")) { '+ret+' }',1);probes.append((id,f,sig,ret))
f='internal/native/runtime/view_unions_mixed.c';sig=next(x for x in sw[f].splitlines() if x.startswith('size_t adamic_view_mixed_union_select('));sw[f]=sw[f].replace(sig,sig+' if (audit_selected("PMixed")) { return 0; }',1);probes.append(('PMixed',f,sig,'return 0;'))
sw['internal/native/tsgo_features_test.go']=sw['internal/native/tsgo_features_test.go'].replace('compile := func(flags []string) error {','compile := func(flags []string) error {\n if auditSelected("W01") { return nil }',1)
for f in ['internal/native/runtime/typed_array.c','internal/native/runtime/view_unions_mixed.c','internal/native/runtime/weak.c']:
 insert='#include <string.h>\nstatic bool audit_selected(const char *id) { const char *selected = getenv("ADAMIC_MUTANT"); return selected != NULL && strcmp(selected, id) == 0; }\n';sw[f]=sw[f].replace('#include <stdlib.h>','#include <stdlib.h>\n'+insert,1)
helper=root/'internal/native/audit_switch.go';helper.write_text('package native\nimport "os"\nfunc auditSelected(id string) bool {return os.Getenv("ADAMIC_MUTANT")==id}\nfunc auditBool(id string,a,b bool) bool {if auditSelected(id){return b};return a}\nfunc auditString(id,a,b string) string {if auditSelected(id){return b};return a}\nfunc auditFlags(flags []string) []string {if auditSelected("M11"){return []string{"adamic-units-v2"}};return append([]string{"adamic-units-v2"},flags...)}\n')
for f,s in sw.items():(root/f).write_text(s)
subprocess.run(['gofmt','-w',*[f for f in sw if f.endswith('.go')],str(helper)],check=True)
with (p/'switch-vet.log').open('w') as f:subprocess.run(['go','vet','./internal/native/'],stdout=f,stderr=subprocess.STDOUT,check=True)
(p/'switch.diff').write_text(subprocess.check_output(['git','diff'],text=True));(p/'audit_switch.go.txt').write_text(helper.read_text());(p/'probe-plan.json').write_text(json.dumps([dict(id=id,file=f,signature=sig,empty=ret) for id,f,sig,ret in probes],indent=2))
runs=[]
for id in [m['id'] for m in meta]+['W01']+[x[0] for x in probes]:
 patterns=[(regex,dict(os.environ))]
 if id in ['M17','PC']:patterns=[('^TestWASIUnit[0-9]+$',dict(sdk))] if id=='PC' else patterns+[('^TestWASIUnit[0-9]+$',dict(sdk))]
 for pattern,env in patterns:
  env.update(ADAMIC_MUTANT=id,ADAMIC_BUILD_CACHE_DIR='/tmp/u054/cache/'+id);suffix='-wasi' if pattern.startswith('^TestWASIUnit') else '';log=p/(id+suffix+'.log');cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run',pattern];start=time.monotonic()
  with log.open('w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
  runs.append(dict(id=id,status=r.returncode,wall=time.monotonic()-start,command=' '.join(cmd),log=log.name,wasi=bool(suffix)));(p/'runs.json').write_text(json.dumps(runs,indent=2))
  if 'panic: test timed out' in log.read_text():continue
  if 'panic: runtime error:' in log.read_text():
   for row in ordinary:
    one=p/(id+'-'+row+'.log');start=time.monotonic();cmd[-1]='^'+row+'$'
    with one.open('w') as f:rr=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
    runs.append(dict(id=id,test=row,status=rr.returncode,wall=time.monotonic()-start,command=' '.join(cmd),log=one.name,wasi=False));(p/'runs.json').write_text(json.dumps(runs,indent=2))
for f,s in orig.items():(root/f).write_text(s)
helper.unlink();print('completed',len(runs),'matrix runs')
