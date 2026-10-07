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
source=(COHERE/'internal/lint/rules/tailwind/collapse/theme.go').read_text()
anchor='func (theme *Theme) Add(key, value string, options ThemeOptions) error {'
assert source.count(anchor)==1
source=source.replace(anchor,anchor+'\n adamicCapture(key,value,options)')
start=source.index(anchor);end=source.index('\n// delete removes',start)
body=source[start:end]
for old,new in [('theme.clearAll()','adamicClearAll(theme)'),('theme.ClearNamespace(strings.TrimSuffix(key, "-*"), ThemeOptionNone)','adamicClearNamespace(theme, strings.TrimSuffix(key, "-*"), ThemeOptionNone)'),('theme.delete(key)','adamicDelete(theme,key)')]:
 assert body.count(old)==1,old;body=body.replace(old,new)
source=source[:start]+body+source[end:];(S/'theme.go').write_text(source)
virtual=COHERE/'adamic_owned_add_oracle.go'
replace={str(COHERE/'internal/lint/rules/tailwind/collapse/theme.go'):str(S/'theme.go'),str(COHERE/'internal/lint/rules/tailwind/collapse/adamic_owned_add_exports.go'):str(HERE/'exports.go.txt'),str(virtual):str(HERE/'oracle.go.txt')}
for filename, constant in [('no_unknown_classes_test.go','unknownFixtureSearchRoot'),('enforce_consistent_class_order_test.go','classOrderFixtureSearchRoot')]:
 original=COHERE/'internal/lint/rules/tailwind'/filename
 text=original.read_text();old='const '+constant+' = \"/Users/kirkouimet/Projects/ahra/app/_theme/styles\"'
 assert text.count(old)==1
 copy=S/filename;copy.write_text(text.replace(old,'const '+constant+' = \"/tmp/adamic-helper-wave1-03-tailwind\"'))
 replace[str(original)]=str(copy)
(S/'go-overlay.json').write_text(json.dumps({'Replace':replace}))
run(['go','build','-overlay='+str(S/'go-overlay.json'),'-o',S/'go-oracle',virtual],'oracle-build',COHERE)
consumers=[('better-tailwindcss/enforce-canonical-classes','EnforceCanonicalClasses'),('better-tailwindcss/enforce-consistent-class-order','EnforceConsistentClassOrder|ClassOrderOptions|ClassOrderMessageNamesTheOrder|ClassOrderFixMatchesThePlugin|ClassOrderReportsWithoutAFixItCannotPlace|ClassOrderSortsTemplateRunsLikeThePlugin|ClassOrderLeavesTemplateRunsThePluginLeaves'),('better-tailwindcss/enforce-consistent-variant-order','EnforceConsistentVariantOrder'),('better-tailwindcss/enforce-shorthand-classes','EnforceShorthandClasses'),('better-tailwindcss/no-conflicting-classes','NoConflictingClasses|ConflictingClassesProposeSuggestionsNotFixes'),('better-tailwindcss/no-unknown-classes','NoUnknownClasses|UnknownIgnoreExemptsProjectClasses|UnknownClassFixturesActuallyRan')]
coverage=[];combined=[];wanted=b''
for i,(name,pattern) in enumerate(consumers):
 capture=S/('capture-'+str(i)+'.jsonl');env=os.environ|{'ADAMIC_THEME_CAPTURE':str(capture)}
 output=run(['go','test','-overlay='+str(S/'go-overlay.json'),'./internal/lint/rules/tailwind','-run','^Test('+pattern+')','-count=1','-v','-timeout=10m'],'consumer-'+str(i),COHERE,env)
 assert b'PASS' in output and b'no tests to run' not in output
 entries=sorted([json.loads(line) for line in capture.read_text().splitlines()],key=lambda row:(row['Key'],row['Value'],row['Options']));assert entries,name
 (S/('entries-'+str(i)+'.json')).write_text(json.dumps(entries))
 raw=strict([S/'go-oracle',S/('entries-'+str(i)+'.json'),name,'--cases'],'cases-'+str(i))
 rows=json.loads(raw);combined.extend(rows)
 expected=strict([S/'go-oracle',S/('entries-'+str(i)+'.json'),name],'expected-'+str(i));wanted+=expected
 coverage.append({'rule':name,'distinct_actual_Add_inputs':len(entries),'state_cases':len(rows)})
controls=[]
for key in ['--*','--font-*','--font-weight-*','--seed','--color-😀','', '--x-*z','--x-*']:
 for value in ['initial','red','', '😀','a`b\n']:
  for options in [0,1,2,3,4,5,6,8,15,16,31]:controls.append({'Key':key,'Value':value,'Options':options})
(S/'controls.json').write_text(json.dumps(controls));combined.extend(json.loads(strict([S/'go-oracle',S/'controls.json','controls','--cases'],'control-cases')))
wanted+=strict([S/'go-oracle',S/'controls.json','controls'],'control-expected')
assert ('\n'.join(row['Output'] for row in combined)+'\n').encode()==wanted
(S/'cases.json').write_text(json.dumps(combined));(S/'go.out').write_bytes(wanted)
virtual_build=ROOT/'adamic_owned_add_build.go';(S/'build-overlay.json').write_text(json.dumps({'Replace':{str(virtual_build):str(HERE/'build.go.txt')}}))
def build(entry,dir,label):run(['go','run','-overlay='+str(S/'build-overlay.json'),virtual_build,entry,dir],label)
def compare(entry,dir,label,mutant=False,corpus=None,expected=None):
 if corpus is None:corpus=S/'cases.json'
 if expected is None:expected=wanted
 cmds=[['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',entry,corpus],['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',dir/'driver.mjs',corpus],[dir/'sanitized',corpus]]
 for path,cmd in zip(['node','emitted','native'],cmds):
  got=strict(cmd,label+'-'+path)
  assert (got!=expected if mutant else got==expected),(label,path)
  if mutant:
   lines=[i for i,(x,y) in enumerate(zip(got.splitlines(),expected.splitlines()),1) if x!=y];assert lines
   print('MUTANT',label,path,'caught by output comparison at line',lines[0])
build(HERE/'main.a',S/'baseline','build-baseline');compare(HERE/'main.a',S/'baseline','baseline')
for name,old,new in [('default-precedence','(existing.options & 4) === 0','(existing.options & 4) !== 0'),('overwrite-order','if(!theme.values.has(key))','if(true)'),('directive-message',"'0: Invalid theme value `'","'0: Wrong theme value `'")]:
 d=S/name;d.mkdir(exist_ok=True)
 helper=(HERE/'theme_add.a').read_text();assert helper.count(old)==1;helper=helper.replace(old,new)
 (d/'theme_add.a').write_text(helper);(d/'main.a').write_text((HERE/'main.a').read_text().replace('../options_json.ts',str(HERE.parent/'options_json.ts')))
 witnesses=[row for row in combined if row['Name']=='controls' and row['Key']==('--font-*' if name=='directive-message' else '--seed') and row['Value']=='red' and row['Options']==(4 if name=='default-precedence' else 0)]
 assert witnesses
 witness=d/'witness.json';witness.write_text(json.dumps(witnesses))
 expected=('\n'.join(row['Output'] for row in witnesses)+'\n').encode()
 build(d/'main.a',d/'built','build-'+name);compare(d/'main.a',d/'built',name,True,witness,expected)
summary={'cases':len(combined),'bytes':len(wanted),'sha256':hashlib.sha256(wanted).hexdigest(),'consumers':coverage,'control_cases':len(controls)*5}
(S/'coverage.json').write_text(json.dumps(summary,indent=2)+'\n');print('PASS',json.dumps(summary))
