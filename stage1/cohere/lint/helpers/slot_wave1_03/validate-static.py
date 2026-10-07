"""Actual Go Add argument capture, independent state controls, three-backend parity."""
import argparse,json,os,subprocess,hashlib
from pathlib import Path
HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[4]
COHERE=ROOT/'cohere'
p=argparse.ArgumentParser();p.add_argument('--scratch',required=True);a=p.parse_args()
S=Path(a.scratch).resolve();S.mkdir(parents=True,exist_ok=True)
def run(cmd,log,cwd=ROOT,env=None):
 with (S/(log+'.out')).open('wb') as out,(S/(log+'.err')).open('wb') as err:
  completed=subprocess.run([str(x) for x in cmd],cwd=cwd,env=env,stdout=out,stderr=err)
 assert completed.returncode==0,(log,completed.returncode,(S/(log+'.err')).read_text()[-5000:])
 return (S/(log+'.out')).read_bytes()
def strict(cmd,log,cwd=ROOT):
 out=run(cmd,log,cwd);assert not (S/(log+'.err')).read_bytes(),log;return out

source=(COHERE/'internal/lint/rules/tailwind/collapse/static_utility.go').read_text()
anchor='func FrameworkStaticReading(name string) (Reading, bool) {'
assert source.count(anchor)==1
source=source.replace(anchor,anchor+'\n adamicCaptureStatic(name)')
old='PropertySort(nodesFromStaticDeclarations(declarations))';assert source.count(old)==1
source=source.replace(old,'adamicStaticReadBody(declarations)')
(S/'static_utility.go').write_text(source)
virtual=COHERE/'adamic_owned_static_oracle.go'
replace={str(COHERE/'internal/lint/rules/tailwind/collapse/static_utility.go'):str(S/'static_utility.go'),str(COHERE/'internal/lint/rules/tailwind/collapse/adamic_owned_static_exports.go'):str(HERE/'static-exports.go.txt'),str(virtual):str(HERE/'static-oracle.go.txt')}
for filename,constant in [('no_unknown_classes_test.go','unknownFixtureSearchRoot'),('enforce_consistent_class_order_test.go','classOrderFixtureSearchRoot')]:
 original=COHERE/'internal/lint/rules/tailwind'/filename;text=original.read_text()
 old='const '+constant+' = "/Users/kirkouimet/Projects/ahra/app/_theme/styles"';assert text.count(old)==1
 copy=S/filename;copy.write_text(text.replace(old,'const '+constant+' = "/tmp/adamic-helper-wave1-03-tailwind"'));replace[str(original)]=str(copy)
(S/'go-overlay.json').write_text(json.dumps({'Replace':replace}))
run(['go','build','-overlay='+str(S/'go-overlay.json'),'-o',S/'go-oracle',virtual],'oracle-build',COHERE)
consumers=[('better-tailwindcss/enforce-canonical-classes','EnforceCanonicalClasses'),('better-tailwindcss/enforce-consistent-class-order','EnforceConsistentClassOrder|ClassOrderOptions|ClassOrderMessageNamesTheOrder|ClassOrderFixMatchesThePlugin|ClassOrderReportsWithoutAFixItCannotPlace|ClassOrderSortsTemplateRunsLikeThePlugin|ClassOrderLeavesTemplateRunsThePluginLeaves'),('better-tailwindcss/enforce-consistent-variant-order','EnforceConsistentVariantOrder'),('better-tailwindcss/enforce-shorthand-classes','EnforceShorthandClasses'),('better-tailwindcss/no-conflicting-classes','NoConflictingClasses|ConflictingClassesProposeSuggestionsNotFixes'),('better-tailwindcss/no-unknown-classes','NoUnknownClasses|UnknownIgnoreExemptsProjectClasses|UnknownClassFixturesActuallyRan')]

combined=[];wanted=b'';coverage=[]
for i,(name,pattern) in enumerate(consumers):
 capture=S/('capture-'+str(i)+'.jsonl');env=os.environ|{'ADAMIC_STATIC_CAPTURE':str(capture)}
 out=run(['go','test','-overlay='+str(S/'go-overlay.json'),'./internal/lint/rules/tailwind','-run','^Test('+pattern+')','-count=1','-v','-timeout=10m'],'consumer-'+str(i),COHERE,env)
 assert b'PASS' in out and b'no tests to run' not in out
 names=sorted({json.loads(line) for line in capture.read_text().splitlines()});assert names,name
 (S/('names-'+str(i)+'.json')).write_text(json.dumps(names))
 combined.extend(json.loads(strict([S/'go-oracle',S/('names-'+str(i)+'.json'),'--cases'],'cases-'+str(i))))
 wanted+=strict([S/'go-oracle',S/('names-'+str(i)+'.json')],'expected-'+str(i))
 coverage.append({'rule':name,'distinct_actual_FrameworkStaticReading_inputs':len(names)})
registered=json.loads(strict([S/'go-oracle','--names'],'registered'))
controls=registered+['','adamic-empty-registration','adamic-body-control','unknown','FLEX','😀','flex\x00','flex ',' flex']
(S/'controls.json').write_text(json.dumps(controls));combined.extend(json.loads(strict([S/'go-oracle',S/'controls.json','--cases'],'control-cases')))
wanted+=strict([S/'go-oracle',S/'controls.json'],'control-expected')
assert ('\n'.join(row['Output'] for row in combined)+'\n').encode()==wanted
(S/'cases.json').write_text(json.dumps(combined));(S/'go.out').write_bytes(wanted)
virtual_build=ROOT/'adamic_owned_static_build.go';(S/'build-overlay.json').write_text(json.dumps({'Replace':{str(virtual_build):str(HERE/'build.go.txt')}}))
def build(entry,d,label):run(['go','run','-overlay='+str(S/'build-overlay.json'),virtual_build,entry,d],label)
def compare(entry,d,label,mutant=False):
 for backend,cmd in [('node',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',entry,S/'cases.json']),('emitted',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',d/'driver.mjs',S/'cases.json']),('native',[d/'sanitized',S/'cases.json'])]:
  got=strict(cmd,label+'-'+backend);assert (got!=wanted if mutant else got==wanted),(label,backend)
  if mutant:
   first=next(i for i,(x,y) in enumerate(zip(got.splitlines(),wanted.splitlines()),1) if x!=y);print('MUTANT',label,backend,'caught only by comparison at line',first)
build(HERE/'static-main.a',S/'baseline','build-baseline');compare(HERE/'static-main.a',S/'baseline','baseline')
for name,old,new in [('missing-found','new StaticReading(false, [], 0, true)','new StaticReading(true, [], 0, true)'),('reading-count','sorted.order, sorted.count, sorted.orderNil','sorted.order, sorted.count + 1, sorted.orderNil'),('reading-order','new StaticReading(true, sorted.order, sorted.count, sorted.orderNil)','new StaticReading(true, [], sorted.count, sorted.orderNil)')]:
 d=S/name;d.mkdir(exist_ok=True)
 code=(HERE/'framework_static_reading.a').read_text();assert code.count(old)==1
 (d/'framework_static_reading.a').write_text(code.replace(old,new));(d/'static-main.a').write_text((HERE/'static-main.a').read_text().replace('../options_json.ts',str(HERE.parent/'options_json.ts')))
 build(d/'static-main.a',d/'built','build-'+name);compare(d/'static-main.a',d/'built',name,True)
summary={'cases':len(combined),'bytes':len(wanted),'sha256':hashlib.sha256(wanted).hexdigest(),'consumers':coverage,'registered_utilities':len(registered),'control_cases':len(controls)}
(S/'coverage.json').write_text(json.dumps(summary,indent=2)+'\n');print('PASS',json.dumps(summary))
