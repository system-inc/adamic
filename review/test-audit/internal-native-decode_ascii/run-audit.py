import pathlib, json, subprocess, os, time, difflib, re, shutil
root=pathlib.Path.cwd(); evidence=root/'review/test-audit/internal-native-decode_ascii'; plan=json.loads((evidence/'plan.json').read_text()); originals={f:(root/f).read_text() for _,f,_,_ in plan}; originals['internal/native/emit.go']=(root/'internal/native/emit.go').read_text()
def shell(args, log, env=None):
 started=time.monotonic()
 with (evidence/log).open('w') as out: result=subprocess.run(args,stdout=out,stderr=subprocess.STDOUT,env=env)
 return result.returncode,round(time.monotonic()-started,3)
def restore():
 for f,text in originals.items(): (root/f).write_text(text)
def diff(f,text): return ''.join(difflib.unified_diff(originals[f].splitlines(True),text.splitlines(True),fromfile='a/'+f,tofile='b/'+f))
def body(text, signature, replacement):
 start=text.index(signature); opening=text.index('{',start); depth=1; end=opening+1
 while depth:
  if text[end]=='{': depth+=1
  if text[end]=='}': depth-=1
  end+=1
 return text[:opening+1]+'\n'+replacement+'\n'+text[end-1:]
# The coverage inventory is fixed before the first mutation.
covered=[]
for profile in sorted(evidence.glob('coverage-*.out')):
 result=subprocess.run(['go','tool','cover','-func='+str(profile)],text=True,capture_output=True,check=True)
 (evidence/(profile.stem+'.functions')).write_text(result.stdout)
 covered.extend(line for line in result.stdout.splitlines() if line.startswith('github.com/system-inc/adamic/internal/native/') and float(line.split()[-1].strip('%'))>0)
(evidence/'reached-go-functions.txt').write_text('\n'.join(sorted(set(covered)))+'\n')
(evidence/'reached-c-functions.txt').write_text('Decoder: new_string, decode_step, encoded_size, ascii_prefix, decode, adamic_decode_utf8; indexing: adamic_string_char_code_at, adamic_string_char_code, adamic_string_units, adamic_string_unit_view, adamic_string_locate, unit_at, sequence, decode; allocation/release/runtime startup transitive functions are not exhaustively instrumented. Dtoa: all functions declared in runtime/dtoa.c, plus adamic_string_from_number, allocation, concat, release, output, panic.\n')
for n,(id,f,old,new) in enumerate(plan,1):
 text=originals[f]; assert old in text; mutated=text.replace(old,new,1); (evidence/(id+'.diff')).write_text(diff(f,mutated)); line=text[:text.index(old)].count('\n')+1
 (evidence/(id+'.location')).write_text(f'{f}:{line}: {old} -> {new}\n')
 (root/f).write_text(mutated)
 args=['go','vet','./internal/native/'] if f.endswith('.go') else ['clang','-std=c11','-Wall','-Wextra','-Werror','-pedantic','-O2','-ffp-contract=off','-fno-optimize-sibling-calls','-I','internal/native/runtime','-c',f,'-o','/tmp/u046-standalone.o']
 rc,secs=shell(args,id+'-compile.log'); print(id,'compile',rc,secs,flush=True); restore(); assert rc==0
switched=originals.copy()
# Runtime selector is cached per process; it does not change the native product between runs.
helper='\n#include <stdlib.h>\nstatic int audit_select(void) { static int ready; static int selected; if (!ready) { const char *value = getenv("ADAMIC_MUTANT"); selected = value == NULL ? 0 : atoi(value); ready = 1; } return selected; }\n'
for f in ['internal/native/runtime/input.c','internal/native/runtime/dtoa.c']: switched[f]=switched[f].replace('#include "adamic.h"','#include "adamic.h"'+helper,1)
cchanges=[('lead >= 0xc2','lead >= (audit_select() == 1 ? 0xc3 : 0xc2)'),('upper = 0x9f;','upper = audit_select() == 2 ? 0xbf : 0x9f;'),('string->units = length + 1;','string->units = length + (audit_select() == 3 ? 2 : 1);'),('bytes[offset] < 0x80','(audit_select() == 4 ? bytes[offset] <= 0x80 : bytes[offset] < 0x80)'),('int exponent = decimal_point - 1;','int exponent = decimal_point - (audit_select() == 5 ? 0 : 1);'),('exponent < -6','exponent < (audit_select() == 6 ? -5 : -6)'),('digits < 0 ||','digits < (audit_select() == 7 ? -1 : 0) ||'),('digits < 1 ||','digits < (audit_select() == 8 ? 0 : 1) ||')]
for (_,f,_,_), (old,new) in zip(plan[:8],cchanges): switched[f]=switched[f].replace(old,new,1)
for n,(id,f,old,new) in enumerate(plan[8:],9):
 if n==9: old,new='!class.Static','class.Static'
 switched[f]=switched[f].replace(old,f'auditChoice({n}, {old}, {new})',1)
probe_specs=[('P01',101,'internal/native/runtime/input.c','adamic_string *adamic_decode_utf8(const unsigned char *bytes, size_t length)','return new_string(0);'),('P02',102,'internal/native/devirtualize.go','func (e *emitter) exactReceiverClass(value ir.Expression) int','return 0'),('P03',103,'internal/native/class_inheritance.go','func (e *emitter) callCode(call ir.Call, arguments []string) string','return ""'),('P04',104,'internal/native/emit.go','func C(program *ir.Program) string','return ""'),('P05',105,'internal/native/runtime/dtoa.c','adamic_string *adamic_number_to_exponential(double value, double fraction_digits, bool has_digits)','return conversion_result("", 0);'),('P06',106,'internal/native/runtime/dtoa.c','adamic_string *adamic_number_to_precision(double value, double precision, bool has_precision)','return conversion_result("", 0);')]
for id,n,f,sig,ret in probe_specs:
 text=switched[f]; start=text.index(sig); opening=text.index('{',start)+1; guard=f'\n\tif (audit_select() == {n}) {{ {ret} }}\n' if f.endswith('.c') else f'\n\tif auditSelect() == {n} {{ {ret} }}\n'; switched[f]=text[:opening]+guard+text[opening:]
 if f.endswith('.go'): standalone=body(originals[f],sig,'\t'+ret)
 else:
  text=originals[f]; opening=text.index('{',text.index(sig))+1; standalone=text[:opening]+'\n\t'+ret+text[opening:]
 (evidence/(id+'.diff')).write_text(diff(f,standalone))
for f,text in switched.items(): (root/f).write_text(text)
helperfile=root/'internal/native/audit_selector.go'; helperfile.write_text('package native\nimport("os"; "strconv")\nfunc auditSelect() int { n,_:=strconv.Atoi(os.Getenv("ADAMIC_MUTANT")); return n }\nfunc auditChoice(id int, normal, changed bool) bool { if auditSelect()==id{return changed};return normal }\n')
shell(['gofmt','-w',str(helperfile),'internal/native/devirtualize.go','internal/native/class_inheritance.go','internal/native/emit.go'],'switch-format.log')
(evidence/'switch.diff').write_text(subprocess.run(['git','diff','--','internal/native'],capture_output=True,text=True).stdout+'\n'+helperfile.read_text())
env=os.environ|{'ADAMIC_TEST_WASI':'1','WASI_SYSROOT':'/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot'}
allpattern='^(TestDecodeASCII(WASI)?Unit[0-9]+|TestDevirtualizedCalls|TestExactReceiverRejectsAssignments|TestToExponentialAndToPrecision.*)$'
def runOne(id,n,pattern):
 directory=pathlib.Path('/tmp/u046/cache')/id
 if id!='inactive' and not directory.exists(): shutil.copytree('/tmp/u046/cache/inactive',directory,copy_function=os.link)
 runenv=env|{'ADAMIC_MUTANT':str(n),'ADAMIC_BUILD_CACHE_DIR':str(directory)}; args=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run',pattern]
 rc,secs=shell(args,id+'.log',runenv); events=[]
 for line in (evidence/(id+'.log')).read_text().splitlines():
  try: events.append(json.loads(line))
  except: pass
 failed=sorted(set(x['Test'].split('/')[0] for x in events if x.get('Action')=='fail' and 'Test' in x)); passed=sorted(set(x['Test'].split('/')[0] for x in events if x.get('Action')=='pass' and 'Test' in x)); print(id,rc,secs,'failed',len(failed),flush=True)
 return {'id':id,'seconds':secs,'exit':rc,'command':'ADAMIC_MUTANT='+str(n)+' ADAMIC_BUILD_CACHE_DIR='+str(directory)+' '+' '.join(args),'failed':failed,'passed':passed,'pattern':pattern}
def run(id,n,pattern):
 if pattern=='^TestDecodeASCII(WASI)?Unit[0-9]+$':
  parts=[runOne(id+'-'+target,n,'^TestDecodeASCII'+prefix+'Unit[0-9]+$') for target,prefix in [('native',''),('wasi','WASI')]]
  (evidence/(id+'.log')).write_text(''.join((evidence/(part['id']+'.log')).read_text() for part in parts))
  return {'id':id,'seconds':sum(part['seconds'] for part in parts),'exit':max(part['exit'] for part in parts),'command':' ; '.join(part['command'] for part in parts),'failed':sorted(set(test for part in parts for test in part['failed'])),'passed':sorted(set(test for part in parts for test in part['passed'])),'pattern':pattern,'parts':parts}
 return runOne(id,n,pattern)
results=[]
try:
 baseline=runOne('inactive',0,'^(TestDevirtualizedCalls|TestExactReceiverRejectsAssignments|TestToExponentialAndToPrecision.*)$'); assert baseline['exit']==0, 'inactive nondecode baseline failed'
 baseline=run('inactive-decode',0,'^TestDecodeASCII(WASI)?Unit[0-9]+$'); assert baseline['exit']==0, 'inactive decode baseline failed' 
 for n,(id,f,old,new) in enumerate(plan,1):
  pattern='^TestDecodeASCII(WASI)?Unit[0-9]+$' if n<=4 else ('^TestToExponentialAndToPrecision' if n<=8 else '^(TestDevirtualizedCalls|TestExactReceiverRejectsAssignments)$')
  result=run(id,n,pattern)
  if 'panic:' in (evidence/(id+'.log')).read_text() and n>=9:
   observations=[]
   for test in ['TestDevirtualizedCalls','TestExactReceiverRejectsAssignments']:
    observation=run(id+'-alone-'+test,n,'^'+test+'$'); observations.append(observation)
   result['reruns']=observations
   result['failed']=sorted(set(test for observation in observations for test in observation['failed']))
   result['passed']=sorted(set(test for observation in observations for test in observation['passed']))
  results.append(result); (evidence/'matrix.json').write_text(json.dumps(results,indent=2))
 for id,n,f,sig,ret in probe_specs:
  pattern='^TestDecodeASCII(WASI)?Unit[0-9]+$' if n==101 else ('^TestToExponentialAndToPrecision' if n>=105 else '^(TestDevirtualizedCalls|TestExactReceiverRejectsAssignments)$')
  results.append(run(id,n,pattern)); (evidence/'matrix.json').write_text(json.dumps(results,indent=2))
finally:
 restore(); helperfile.unlink()
