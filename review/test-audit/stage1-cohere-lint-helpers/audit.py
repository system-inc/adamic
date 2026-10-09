import pathlib,json,re,subprocess,os,time,difflib,gzip
r=pathlib.Path('/workspace/adamic');p=r/'review/test-audit/stage1-cohere-lint-helpers';d=r/'stage1/cohere/lint/helpers';tmp=pathlib.Path('/tmp/u114');tmp.mkdir(exist_ok=True);started=time.monotonic();runs=json.loads((p/'runs.json').read_text()) if (p/'runs.json').exists() else [];names=['TestHelpersMatchCohere','TestHelperMutants','TestMessageRefusalsMatchGo','TestKnownGapsAreExplicit'];prod=[x for x in names if x!='TestHelperMutants'];base={str(f.relative_to(r)):f.read_text() for f in d.glob('*.ts')};base['stage1/cohere/lint/helpers/helpers_test.go']=(d/'helpers_test.go').read_text()
menu=[dict(id='M1',file='stage1/cohere/lint/helpers/options_json.ts',old='a.number === b.number',new='a.number !== b.number',kind='flip condition',function='OptionsJson.equal'),dict(id='M2',file='stage1/cohere/lint/helpers/option_schema.ts',old='!known.includes(key)',new='known.includes(key)',kind='flip condition',function='OptionSchema.audit'),dict(id='M3',file='stage1/cohere/lint/helpers/strict_options.ts',old="if(value.kind === 'null') { return true; }",new="if(value.kind === 'null') { return false; }",kind='change constant',function='StrictOptions.matches'),dict(id='M4',file='stage1/cohere/lint/helpers/policy_message.ts',old='chosen.size !== count',new='chosen.size !== count + 1',kind='off-by-one bound',function='PolicyMessage.render')]
def diff(f,new):return ''.join(difflib.unified_diff(base[f].splitlines(True),new.splitlines(True),fromfile='a/'+f,tofile='b/'+f))
for m in menu:
 assert base[m['file']].count(m['old'])==1;m['line']=base[m['file']][:base[m['file']].index(m['old'])].count('\n')+1;(p/(m['id']+'.diff')).write_text(diff(m['file'],base[m['file']].replace(m['old'],m['new'],1)))
(p/'menu.json').write_text(json.dumps(menu,indent=2));(p/'rows.txt').write_text('\n'.join(names)+'\n')
functions=[];callbacks=[]
for f,s in base.items():
 if not f.endswith('.ts'):continue
 cls=''
 for i,l in enumerate(s.splitlines(),1):
  cm=re.match(r'export class (\w+)',l)
  if cm:cls=cm[1]
  fm=re.match(r'    (constructor|\w+)\([^)]*\):?',l)
  if fm:functions.append(dict(file=f,line=i,name=cls+'.'+fm[1]))
  fm=re.match(r'function (\w+)\(',l)
  if fm:functions.append(dict(file=f,line=i,name=fm[1]))
  if '=>' in l:callbacks.append(dict(file=f,line=i,source=l.strip()))
functions.append(dict(file='stage1/cohere/lint/helpers/main.ts',line=7,name='module entry'))
(p/'functions.json').write_text(json.dumps(dict(named=functions,callbacks=callbacks,scope='All four helper classes and main entry; the full agreement corpus reaches their functions. Callback entries show anonymous function expressions; branch coverage is not asserted.'),indent=2))
(p/'CODE-AND-ORACLE.md').write_text('CODE UNDER TEST: stage1/cohere/lint/helpers TypeScript port, compiled by Adamic and executed natively, also executed as source on Node. main.ts dispatches OptionsJson, OptionSchema, StrictOptions and PolicyMessage. functions.json lists source functions and callbacks before mutations.\nORACLE: independent Go encoding/json, Go cohere optionschema, generated Go target decoding and policy.Messages.Render. TestKnownGapsAreExplicit instead pins handwritten NotYet/valid strings without an outside authority. Node runs the implementation; it is not substituted for the independent Go expectation.\nFour production mutants were selected from those functions before observing mutant failures: invert numeric equality; invert known-keyword audit; reject null; shift required phrase count by one. They are standalone diffs with no selector and each gets a native build. P is a separate empty-main probe, not a mutant. W1 is a separate comparison-accepts-everything harness weakening for the built-in mutant witness.\nThe clean whole package hit 90 seconds; remaining rows passed a narrowed clean run. Production matrices run all three production rows, excluding the witness whose production results are inadmissible. Package uniqueness is therefore reported bounded.\n')
f='stage1/cohere/lint/helpers/main.ts';(p/'P.diff').write_text(diff(f,'export {};\n'))
f='stage1/cohere/lint/helpers/helpers_test.go';weak='\t\t\tif bytes.Equal(got, want) {';assert base[f].count(weak)==1;(p/'W1.diff').write_text(diff(f,base[f].replace(weak,'\t\t\tif bytes.Equal(got, got) {',1)))

def run(cmd,log,id='clean',cwd=r,expected=None):
 if any(x['log']==log and x['exit']==0 for x in runs):return 0
 if '-timing-' in log and (p/log).exists():
  events=[]
  for line in (p/log).read_text().splitlines():
   try:events.append(json.loads(line))
   except:pass
  passed=[e for e in events if e.get('Action')=='pass' and 'Test' not in e]
  if passed:
   runs.append(dict(command=[str(c) for c in cmd],cwd=str(cwd),log=log,selector=id,exit=0,wall=None,recovered_binary_elapsed=passed[0]['Elapsed']));(p/'runs.json').write_text(json.dumps(runs,indent=2));return 0
 pathlib.Path('/tmp/u114-mutant').write_text(id);env=os.environ.copy();env['ADAMIC_MUTANT']=id;begin=time.monotonic()
 with (p/log).open('w') as out:q=subprocess.run(cmd,cwd=cwd,env=env,stdout=out,stderr=subprocess.STDOUT)
 v=dict(command=[str(c) for c in cmd],cwd=str(cwd),log=log,selector=id,exit=q.returncode,wall=time.monotonic()-begin);runs.append(v);(p/'runs.json').write_text(json.dumps(runs,indent=2));print(log,q.returncode,round(v['wall'],3),flush=True)
 if expected is not None:assert q.returncode==expected,(log,q.returncode)
 return q.returncode

def test(rows,log,id='clean',expected=None):return run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/helpers/','-run','^('+'|'.join(rows)+')$'],log,id,expected=expected)
try:
 # Serial isolated timings with unchanged source, no contention from another test process.
 for name in names:
  for i in range(1,4):test([name],name+'-timing-'+str(i)+'.log',expected=0)
 # Build the compiler and independent oracle once; collect native replay input and expectation.
 run(['timeout','90','go','build','-o','/tmp/u114/adamic','./cmd/adamic'],'compiler-build.log',expected=0)
 root=r/'cohere';overlay={'Replace':{str(root/'adamic_helper_oracle.go'):str(d/'testdata/oracle.go'),str(root/'policy/adamic_helpers.go'):str(d/'testdata/catalog.go'),str(root/'adamic_helper_descriptors.go'):str(d/'testdata/descriptors.go')}};(tmp/'overlay.json').write_text(json.dumps(overlay));(p/'oracle-overlay.json').write_text(json.dumps(overlay,indent=2));run(['timeout','90','go','build','-overlay='+str(tmp/'overlay.json'),'-o',str(tmp/'go-oracle'),str(root/'adamic_helper_oracle.go'),str(root/'adamic_helper_descriptors.go')],'oracle-build.log',cwd=root,expected=0)
 (tmp/'cases.json').write_bytes(gzip.decompress((d/'testdata/cases.json.gz').read_bytes()));run([str(tmp/'go-oracle'),str(tmp/'cases.json')],'oracle-answers.log',expected=0)
 for m in menu:
  f=m['file'];(r/f).write_text(base[f].replace(m['old'],m['new'],1));run(['timeout','90',str(tmp/'adamic'),'build','stage1/cohere/lint/helpers/main.ts','-o',str(tmp/m['id']),'--sanitize'],m['id']+'-native-build.log',expected=0)
  # Run the native product independently even when the Go test fails first on source Node.
  run([str(tmp/m['id']),str(tmp/'cases.json'),str(d/'testdata/catalog.json')],m['id']+'-native-output.log')
  (r/f).write_text(base[f])
 # Empty-main native proof and comparison-weakening Go vet proof are separate.
 f='stage1/cohere/lint/helpers/main.ts';(r/f).write_text('export {};\n');run(['timeout','90',str(tmp/'adamic'),'build',f,'-o',str(tmp/'P'),'--sanitize'],'P-native-build.log',expected=0);(r/f).write_text(base[f])
 f='stage1/cohere/lint/helpers/helpers_test.go';(r/f).write_text(base[f].replace(weak,'\t\t\tif bytes.Equal(got, got) {',1));run(['timeout','90','go','vet','./stage1/cohere/lint/helpers/'],'W1-vet.log',expected=0);(r/f).write_text(base[f])
 # Existing files carry the selector so the witness's fixed copied-file list stays untouched.
 f='stage1/cohere/lint/helpers/options_json.ts';s=base[f].replace("import { panic } from 'adamic';","import { panic, readTextFile } from 'adamic';",1);s=s.replace('export interface OptionValue',"function auditChoice(): string { const value = readTextFile('/tmp/u114-mutant'); return value.kind === 'Ok' ? value.text.trim() : ''; }\nexport const auditSelection = auditChoice();\nexport interface OptionValue",1);s=s.replace(menu[0]['old'],"(auditSelection === 'M1' ? a.number !== b.number : a.number === b.number)",1);(r/f).write_text(s)
 for m in menu[1:]:
  f=m['file'];s=base[f].replace("import { OptionsJson }", "import { OptionsJson, auditSelection }",1)
  replacement={'M2':"(auditSelection === 'M2' ? known.includes(key) : !known.includes(key))",'M3':"if(value.kind === 'null') { return auditSelection !== 'M3'; }",'M4':"(auditSelection === 'M4' ? chosen.size !== count + 1 : chosen.size !== count)"}[m['id']];(r/f).write_text(s.replace(m['old'],replacement,1))
 f='stage1/cohere/lint/helpers/helpers_test.go';(r/f).write_text(base[f].replace(weak,'\t\t\tif bytes.Equal(got, want) || os.Getenv("ADAMIC_MUTANT") == "W1" {',1))
 for f in base:(p/(pathlib.Path(f).name+'.switch.txt')).write_text((r/f).read_text())
 run(['timeout','90','go','test','-c','-o',str(tmp/'test'),'./stage1/cohere/lint/helpers/'],'switch-build.log',expected=0)
 test(names,'switch-baseline.log',expected=0)
 # Whole bounded matrix over three production rows for each selector. Witness is weakened separately.
 for m in menu:test(prod,m['id']+'-matrix.log',m['id'])
 test(['TestHelperMutants'],'W1.log','W1')
 # Return the port entry's empty answer, separate from the production mutant switch.
 for f,s in base.items():(r/f).write_text(s)
 (d/'main.ts').write_text('export {};\n')
 test(prod,'P.log','P')
finally:
 for f,s in base.items():(r/f).write_text(s)
 pathlib.Path('/tmp/u114-mutant').write_text('clean')
run(['go','vet','./stage1/cohere/lint/helpers/'],'final-vet.log',expected=0)
(p/'elapsed.txt').write_text(str(time.monotonic()-started)+'\n')
