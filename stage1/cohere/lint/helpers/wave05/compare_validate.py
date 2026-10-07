from pathlib import Path
import subprocess,json,os,hashlib,shutil,random
here=Path(__file__).resolve().parent;repo=here.parents[4];cohere=repo/'cohere';scratch=Path('/tmp/w05-helper-compare');scratch.mkdir(exist_ok=True)
def run(args,label,cwd=repo,env=None):
 with (scratch/(label+'.log')).open('wb') as out,(scratch/(label+'.stderr.log')).open('wb') as err:
  subprocess.run(args,cwd=cwd,env=env,stdout=out,stderr=err,check=True)
 assert (scratch/(label+'.stderr.log')).stat().st_size==0,label
source=cohere/'internal/lint/rules/tailwind/collapse/variant.go';text=source.read_text();anchor='func CompareBreakpoints(left string, right string, ascending bool) int {';assert text.count(anchor)==1
(scratch/'variant.go').write_text(text.replace(anchor,'func ownedOriginalCompareBreakpoints(left string, right string, ascending bool) int {'))
bridge=cohere/'internal/lint/rules/tailwind/collapse/adamic_compare.go';virtual=cohere/'adamic_compare_oracle.go'
replace={str(source):str(scratch/'variant.go'),str(bridge):str(here/'compare_bridge.go.txt'),str(virtual):str(here/'compare_oracle.go.txt')}
harness=cohere/'internal/lint/testing/rule_testing.go';text=harness.read_text();anchor='return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}';assert text.count(anchor)==1
(scratch/'capture.go').write_text(text.replace(anchor,'result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t,result)\n return result'))
testreplace={**replace,str(harness):str(scratch/'capture.go')}
for name in ['no_unknown_classes_test.go','enforce_consistent_class_order_test.go']:
 file=cohere/'internal/lint/rules/tailwind'/name
 text=file.read_text();assert '/Users/kirkouimet/Projects/ahra/app/_theme/styles' in text
 side=scratch/name;side.write_text(text.replace('/Users/kirkouimet/Projects/ahra/app/_theme/styles','/tmp/w05-helper-tailwind'))
 testreplace[str(file)]=str(side)
(scratch/'test-overlay.json').write_text(json.dumps({'Replace':testreplace}))
trace=scratch/'trace.jsonl';trace.write_text('');capture=scratch/'capture';capture.mkdir(exist_ok=True)
for f in capture.glob('*.jsonl'):f.unlink()
env=os.environ|{'ADAMIC_COMPARE_CAPTURE':str(trace),'COHERE_DOCS_CAPTURE':str(capture)}
pattern='^Test(EnforceCanonicalClasses|Canonical(Message|Proposes|Fixtures|Collapse|Logical)|EnforceConsistent(Class|Variant)Order|EnforceShorthandClasses|NoConflictingClasses|ConflictingClassesPropose|NoUnknownClasses|Unknown(Ignore|ClassFixtures)|KnownRootWithUnknownValue)'
run(['go','test','-overlay='+str(scratch/'test-overlay.json'),'./internal/lint/rules/tailwind','-run',pattern,'-count=1','-timeout=10m','-v'],'consumer-tests',cohere,env)
ledger=json.loads((here.parent/'readiness.json').read_text());symbol='github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.CompareBreakpoints';consumers={r['rule'] for r in ledger['remaining'] if symbol in r['remaining_helpers']}
rows=[]
for f in capture.glob('*.jsonl'):
 for line in f.read_text().splitlines():
  row=json.loads(line)
  if row['rule'] in consumers:rows.append(row)
assert consumers=={r['rule'] for r in rows},sorted(consumers-{r['rule'] for r in rows})
values=[json.loads(line) for line in trace.read_text().splitlines()];tracecount=len(values)
inputs=json.loads((here/'leading_cases.json').read_text())
inputs+=['calc(1rem)','calc(2rem)','min(10px)','max(9px)','40rem','20px','0rem','0px','😀','\ue000','甲','é','😀40rem','\ue00040rem','calc(😀)','calc(\ue000)']
for text in inputs:
 for other in ['42','-42','40rem',text]:
  for ascending in [False,True]:values.append({'left':text,'right':other,'ascending':ascending})
rng=random.Random(15)
for i in range(1500):values.append({'left':rng.choice(inputs),'right':rng.choice(inputs),'ascending':bool(i%2)})
cases=scratch/'cases.json';cases.write_text(json.dumps(values,ensure_ascii=True))
(scratch/'oracle-overlay.json').write_text(json.dumps({'Replace':replace}));run(['go','build','-overlay='+str(scratch/'oracle-overlay.json'),'-o',str(scratch/'oracle'),str(virtual)],'oracle-build',cohere)
run([str(scratch/'oracle'),str(cases),'facts'],'facts');cases.write_bytes((scratch/'facts.log').read_bytes())
(scratch/'compare_consumers.json').write_text(json.dumps({'consumers':sorted(consumers),'capturedFixtures':len(rows),'actualCalls':tracecount,'cases':len(values),'inputSha256':hashlib.sha256(cases.read_bytes()).hexdigest(),'dependency':'breakpointBucket facts from actual Go; caller supplies callback'},indent=2)+'\n')
run([str(scratch/'oracle'),str(cases)],'Go');want=(scratch/'Go.log').read_bytes()
build=here/'build.go';(scratch/'build-overlay.json').write_text(json.dumps({'Replace':{str(build):str(here/'build.go.txt')}}))
def compare(entry,label,mutant=False):
 binary=scratch/(label+'-binary')
 run(['go','run','-overlay='+str(scratch/'build-overlay.json'),str(build),str(entry),str(binary)],label+'-build')
 for side,args in [('Node',['node','--disable-warning=ExperimentalWarning',str(repo/'oracle/node.mjs'),str(entry),str(cases)]),('JavaScript',['node','--disable-warning=ExperimentalWarning',str(repo/'oracle/node.mjs'),str(binary)+'.mjs',str(cases)]),('native',[str(binary),str(cases)])]:
  run(args,label+'-'+side);got=(scratch/(label+'-'+side+'.log')).read_bytes()
  assert (got!=want) if mutant else (got==want),(label,side)
  print(label,side,'mutant caught' if mutant else 'identical',len(got),'bytes',flush=True)
compare(here/'compare_driver.a','baseline')
mutant=scratch/'mutant';mutant.mkdir(exist_ok=True)
s=(here.parent/'collapse_compare_breakpoints.a').read_text().replace("'./collapse_leading_integer.a'",json.dumps(str(here.parent/'collapse_leading_integer.a')));anchor='const first = ascending ? a : b;';assert s.count(anchor)==1
(mutant/'collapse_compare_breakpoints.a').write_text(s.replace(anchor,'const first = ascending ? b : a;'))
s=(here/'compare_driver.a').read_text().replace("'../options_json.ts'",json.dumps(str(here.parent/'options_json.ts'))).replace("'../collapse_compare_breakpoints.a'","'./collapse_compare_breakpoints.a'")
(mutant/'driver.a').write_text(s);compare(mutant/'driver.a','mutant',True)
print('consumers',len(consumers),'fixtures',len(rows),'actual calls',tracecount,'cases',len(values),flush=True)
