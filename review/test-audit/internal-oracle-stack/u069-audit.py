import pathlib,json,subprocess,os,time,difflib
p=pathlib.Path('review/test-audit/internal-oracle-stack');(p/'diffs').mkdir(exist_ok=True);(p/'probes').mkdir(exist_ok=True);scratch=pathlib.Path('/tmp/u069');scratch.mkdir(exist_ok=True);selector=scratch/'mutant';selector.write_text('0')
plan=[]
def add(id,file,old,new,kind):
 s=pathlib.Path(file).read_text();assert s.count(old)==1,(id,s.count(old));plan.append(dict(id=id,file=file,line=s[:s.index(old)].count('\n')+1,old=old,new=new,kind=kind))
c='internal/native/runtime/stack.c';g='internal/native/emit_functions.go'
add('M01',c,'uintptr_t arguments = size / 4 > ARGUMENTS_FLOOR ? size / 4 : ARGUMENTS_FLOOR;','uintptr_t arguments = 0;','change constant: argument reservation to zero')
add('M02',c,'size / 8 < MARGIN ? size / 8 : MARGIN','size / 8 > MARGIN ? size / 8 : MARGIN','flip condition')
add('M03',g,'\te.line("ADAMIC_CHECK_STACK();")','', 'drop statement')
add('M04',c,'"RangeError: Maximum call stack size exceeded"','"RangeError: stack limit exceeded"','change constant')
(p/'mutant-plan.json').write_text(json.dumps(plan,indent=2));original={f:pathlib.Path(f).read_text() for f in [c,g]};original['internal/native/emit.go']=pathlib.Path('internal/native/emit.go').read_text();results=[];probe_results=[]
pattern='^(TestLongArgumentsLeaveTheStackItsLimit|TestSmallStacksStillPanic|TestNativeAgreesWithNode)$/^(arguments|arguments_alone|environment|1024_KiB|512_KiB|256_KiB|internal)$/^oracle$/^testdata$/^stack_overflow[.]a$'
flags=['-std=c11','-Wall','-Wextra','-Werror','-Wcast-function-type-strict','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls']
def validate(q,after):
 file=pathlib.Path(q['file']);before=file.read_text();start=time.monotonic();commands=[]
 try:
  file.write_text(after)
  with (p/'logs'/('validate-'+q['id']+'.log')).open('w') as out:
   if q['file'].endswith('.c'):
    for name,extra in [('release',['-O2']),('sanitized',['-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all'])]:
     cmd=['clang']+flags+extra+['-I','internal/native/runtime','-c',q['file'],'-o','/tmp/u069/'+q['id']+'-'+name+'.o'];commands.append(' '.join(cmd));subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT,check=True)
   else:
    cmd=['go','vet','./internal/native/'];commands.append(' '.join(cmd));subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT,check=True)
  q['validation_commands']=commands;q['validation_seconds']=time.monotonic()-start;q['validated']=True
 finally:file.write_text(before)
def run(id,regex,path):
 selector.write_text(str(int(id[1:]) if id.startswith('M') else {'P01':6,'P02':5}[id]));env=dict(os.environ,ADAMIC_GATE_UNCACHED='1',ADAMIC_BUILD_CACHE_DIR='/tmp/u069/cache/'+id);cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',regex];start=time.monotonic()
 with path.open('w') as out:r=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
 es=[]
 for s in path.read_text().splitlines():
  try:es.append(json.loads(s))
  except:pass
 fails=sorted(set(e['Test'].split('/')[0] for e in es if e['Action']=='fail' and e.get('Test')));panic=any(e.get('Output','').startswith('panic:') for e in es);assert not panic and r.returncode!=124,'isolate reruns needed'
 return dict(id=id,exit=r.returncode,command=' '.join(cmd),selector_value=selector.read_text(),environment={'ADAMIC_GATE_UNCACHED':'1','ADAMIC_BUILD_CACHE_DIR':env['ADAMIC_BUILD_CACHE_DIR']},wall_seconds=time.monotonic()-start,kills=fails,events=es,bounded=True)
try:
 for q in plan:
  before=original[q['file']];after=before.replace(q['old'],q['new']);(p/'diffs'/(q['id']+'.diff')).write_text(''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='a/'+q['file'],tofile='b/'+q['file'])));validate(q,after)
 probes=[dict(id='P01',file='internal/native/emit.go',old='func C(program *ir.Program) string {',new='func C(program *ir.Program) string {\n\tif true { return "" }',kind='empty C entry'),dict(id='P02',file=c,old='__attribute__((constructor)) static void find_stack_limit(void) {\n\tuintptr_t base',new='__attribute__((constructor)) static void find_stack_limit(void) {\n\treturn;\n\tuintptr_t base',kind='empty runtime initialization')]
 for q in probes:
  before=original[q['file']];after=before.replace(q['old'],q['new']);assert after!=before;q['line']=before[:before.index(q['old'])].count('\n')+1;(p/'probes'/(q['id']+'.diff')).write_text(''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='a/'+q['file'],tofile='b/'+q['file'])));validate(q,after)
 (p/'probe-plan.json').write_text(json.dumps(probes,indent=2))
 switched=original[c].replace('uintptr_t arguments = size / 4 > ARGUMENTS_FLOOR ? size / 4 : ARGUMENTS_FLOOR;','uintptr_t arguments = audit_mutant() == 1 ? 0 : (size / 4 > ARGUMENTS_FLOOR ? size / 4 : ARGUMENTS_FLOOR);').replace('uintptr_t margin = size / 8 < MARGIN ? size / 8 : MARGIN;','uintptr_t margin = audit_mutant() == 2 ? (size / 8 > MARGIN ? size / 8 : MARGIN) : (size / 8 < MARGIN ? size / 8 : MARGIN);')
 switched=switched.replace('#include <stdint.h>','#include <stdint.h>\n#include <stdio.h>\nstatic int audit_mutant(void) { FILE *file = fopen("/tmp/u069/mutant", "r"); int selected = 0; if (file != NULL) { (void)fscanf(file, "%d", &selected); (void)fclose(file); } return selected; }')
 switched=switched.replace('__attribute__((constructor)) static void find_stack_limit(void) {\n\tuintptr_t base','__attribute__((constructor)) static void find_stack_limit(void) {\n\tif (audit_mutant() == 5) return;\n\tuintptr_t base')
 switched=switched.replace('static const char message[] = "RangeError: Maximum call stack size exceeded";','if (audit_mutant() == 4) { static const char changed[] = "RangeError: stack limit exceeded"; adamic_panic(changed, sizeof changed - 1); }\n\tstatic const char message[] = "RangeError: Maximum call stack size exceeded";')
 pathlib.Path(c).write_text(switched);pathlib.Path(g).write_text(original[g].replace('\te.line("ADAMIC_CHECK_STACK();")','\tif !auditSelected("3") { e.line("ADAMIC_CHECK_STACK();") }'))
 pathlib.Path('internal/native/emit.go').write_text(original['internal/native/emit.go'].replace('func C(program *ir.Program) string {','func C(program *ir.Program) string {\n\tif auditSelected("6") {return ""}'))
 pathlib.Path('internal/native/audit_switch.go').write_text('package native\nimport("os";"strings")\nfunc auditSelected(id string)bool{b,_:=os.ReadFile("/tmp/u069/mutant");return strings.TrimSpace(string(b))==id}\n')
 start=time.monotonic()
 with (p/'logs'/'compile-switch.log').open('w') as out:subprocess.run(['go','test','-c','-o','/tmp/u069.test','./internal/oracle/'],stdout=out,stderr=subprocess.STDOUT,check=True)
 (p/'switch-build.json').write_text(json.dumps(dict(seconds=time.monotonic()-start)))
 result=run('M00',pattern,p/'logs'/'switch-baseline.log');assert result['exit']==0,'red switched baseline';(p/'switch-baseline.json').write_text(json.dumps(result,indent=2))
 for q in plan:
  result=run(q['id'],pattern,p/'logs'/(q['id']+'.log'));results.append(result);(p/'matrix.json').write_text(json.dumps(results,indent=2));print(q['id'],result['exit'],result['kills'],flush=True)
 for q in probes:
  for name in ['TestLongArgumentsLeaveTheStackItsLimit','TestSmallStacksStillPanic']:
   result=run(q['id'],'^'+name+'$',p/'logs'/(q['id']+'-'+name+'.log'));result['test']=name;probe_results.append(result);(p/'probe-results.json').write_text(json.dumps(probe_results,indent=2));print(q['id'],name,result['exit'],flush=True)
finally:
 for f,s in original.items():pathlib.Path(f).write_text(s)
 pathlib.Path('internal/native/audit_switch.go').unlink(missing_ok=True);selector.write_text('0');(p/'mutant-plan.json').write_text(json.dumps(plan,indent=2))
