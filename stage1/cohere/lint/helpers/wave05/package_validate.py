from pathlib import Path
import subprocess,json,os,hashlib,shutil,random
here=Path(__file__).resolve().parent;repo=here.parents[4];cohere=repo/'cohere';scratch=Path('/tmp/w05-helper-package');scratch.mkdir(exist_ok=True)
def run(args,label,cwd=repo,env=None):
 with (scratch/(label+'.log')).open('wb') as out,(scratch/(label+'.stderr.log')).open('wb') as err:
  subprocess.run(args,cwd=cwd,env=env,stdout=out,stderr=err,check=True)
 assert (scratch/(label+'.stderr.log')).stat().st_size==0,label
source=cohere/'internal/lint/rules/tailwind/design_system.go';text=source.read_text();anchor='func findTailwindPackageRoot(start string, fileExists func(path string) bool) string {';assert text.count(anchor)==1
(scratch/'variant.go').write_text(text.replace(anchor,'func ownedOriginalPackageRoot(start string, fileExists func(path string) bool) string {'))
bridge=cohere/'internal/lint/rules/tailwind/adamic_package.go';virtual=cohere/'adamic_package_oracle.go'
replace={str(source):str(scratch/'variant.go'),str(bridge):str(here/'package_bridge.go.txt'),str(virtual):str(here/'package_oracle.go.txt')}
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
env=os.environ|{'ADAMIC_PACKAGE_CAPTURE':str(trace),'COHERE_DOCS_CAPTURE':str(capture)}
pattern='^Test(EnforceCanonicalClasses|Canonical(Message|Proposes|Fixtures|Collapse|Logical)|EnforceConsistent(Class|Variant)Order|EnforceShorthandClasses|NoConflictingClasses|ConflictingClassesPropose|NoUnknownClasses|Unknown(Ignore|ClassFixtures)|KnownRootWithUnknownValue)'
run(['go','test','-overlay='+str(scratch/'test-overlay.json'),'./internal/lint/rules/tailwind','-run',pattern,'-count=1','-timeout=10m','-v'],'consumer-tests',cohere,env)
ledger=json.loads((here.parent/'readiness.json').read_text());symbol='github.com/system-inc/cohere/internal/lint/rules/tailwind.findTailwindPackageRoot';consumers={r['rule'] for r in ledger['remaining'] if symbol in r['remaining_helpers']}
rows=[]
for f in capture.glob('*.jsonl'):
 for line in f.read_text().splitlines():
  row=json.loads(line)
  if row['rule'] in consumers:rows.append(row)
assert consumers=={r['rule'] for r in rows},sorted(consumers-{r['rule'] for r in rows})
starts={json.loads(line) for line in trace.read_text().splitlines()};tracecount=len(trace.read_text().splitlines())
# Actual paths are retained with their observed temp directories; bounded roots
# supplement every consumer with every possible search hit and competing hits.
starts.update(['','.', '/', '//', '..', '../..', '/a/b/c', '/a/b/', '/a/../b', 'a/b', 'a//b/./c', 'a/../../b', 'C:\\a\\b', '/😀/é/甲', '/a/..', 'a/..'])
values=[]
import posixpath
def clean(s):
 r=posixpath.normpath(s)
 return '/'+r.lstrip('/') if r.startswith('/') else r
def parent(s):
 i=s.rfind('/');return clean('' if i<0 else s[:i+1])
for start in sorted(starts):
 current=start;probes=[]
 while True:
  path=clean((current+'/' if current else '')+'node_modules/tailwindcss')+'/index.css';probes.append(path)
  nxt=parent(current)
  if nxt==current:break
  current=nxt
 unique=list(dict.fromkeys(probes))
 for files in [[],unique]+[[p] for p in unique]:values.append({'start':start,'exists':files})
 # unrelated files and candidates below/above the caller do not cause a hit.
 values.append({'start':start,'exists':[p+'/wrong.css' for p in unique]})
cases=scratch/'cases.json';cases.write_text(json.dumps(values,ensure_ascii=True));(scratch/'package_consumers.json').write_text(json.dumps({'consumers':sorted(consumers),'capturedFixtures':len(rows),'actualCalls':tracecount,'uniqueCallArguments':len(starts),'cases':len(values),'inputSha256':hashlib.sha256(cases.read_bytes()).hexdigest()},indent=2)+'\n')
(scratch/'oracle-overlay.json').write_text(json.dumps({'Replace':replace}));run(['go','build','-overlay='+str(scratch/'oracle-overlay.json'),'-o',str(scratch/'oracle'),str(virtual)],'oracle-build',cohere);run([str(scratch/'oracle'),str(cases)],'Go');want=(scratch/'Go.log').read_bytes()
build=here/'build.go';(scratch/'build-overlay.json').write_text(json.dumps({'Replace':{str(build):str(here/'build.go.txt')}}))
def compare(entry,label,mutant=False):
 binary=scratch/(label+'-binary')
 run(['go','run','-overlay='+str(scratch/'build-overlay.json'),str(build),str(entry),str(binary)],label+'-build')
 for side,args in [('Node',['node','--disable-warning=ExperimentalWarning',str(repo/'oracle/node.mjs'),str(entry),str(cases)]),('JavaScript',['node','--disable-warning=ExperimentalWarning',str(repo/'oracle/node.mjs'),str(binary)+'.mjs',str(cases)]),('native',[str(binary),str(cases)])]:
  run(args,label+'-'+side);got=(scratch/(label+'-'+side+'.log')).read_bytes()
  assert (got!=want) if mutant else (got==want),(label,side)
  print(label,side,'mutant caught' if mutant else 'identical',len(got),'bytes',flush=True)
compare(here/'package_driver.a','baseline')
mutant=scratch/'mutant';mutant.mkdir(exist_ok=True)
s=(here.parent/'tailwind_find_package_root.a').read_text();anchor="if(parent === current) { return ''; }";assert s.count(anchor)==1
(mutant/'tailwind_find_package_root.a').write_text(s.replace(anchor,"if(parent === current || current.startsWith('/')) { return ''; }"))
s=(here/'package_driver.a').read_text().replace("'../options_json.ts'",json.dumps(str(here.parent/'options_json.ts'))).replace("'../tailwind_find_package_root.a'","'./tailwind_find_package_root.a'").replace("'../../../../typescript/parser/nodes.ts'",json.dumps(str(repo/'stage1/typescript/parser/nodes.ts')))
(mutant/'driver.a').write_text(s);compare(mutant/'driver.a','mutant',True)
print('consumers',len(consumers),'fixtures',len(rows),'actual calls',tracecount,'cases',len(values),flush=True)
