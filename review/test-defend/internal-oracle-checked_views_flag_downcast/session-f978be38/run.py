import pathlib,subprocess,os,json,time,difflib
root=pathlib.Path('/workspace/adamic');p=root/'review/test-defend/internal-oracle-checked_views_flag_downcast/session-f978be38'
base=os.environ.copy();base['ADAMIC_GATE_UNCACHED']='1'
def run(label,selector,cover=False,mut='clean'):
 env=base.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/checked-views-defense/cache/'+mut
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s']
 if cover:cmd+=['-coverpkg=./internal/lower,./internal/native,./internal/javascript','-coverprofile='+str(p/(label+'.cover'))]
 cmd+=['./internal/oracle/','-run',selector];start=time.monotonic()
 with (p/(label+'.log')).open('w') as out:r=subprocess.run(cmd,cwd=root,env=env,stdout=out,stderr=subprocess.STDOUT)
 events=[]
 for s in (p/(label+'.log')).read_text().splitlines():
  try:events.append(json.loads(s))
  except:pass
 rec=dict(label=label,mutant=mut,command='ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+ ' '.join(cmd),exit=r.returncode,wall_seconds=time.monotonic()-start,failures=[x.get('Test') for x in events if x.get('Action')=='fail' and x.get('Test')],passes=[x.get('Test') for x in events if x.get('Action')=='pass' and x.get('Test')],cooked=any('test timed out' in x.get('Output','') for x in events))
 runs.append(rec);(p/'matrix.json').write_text(json.dumps(runs,indent=2));print(label,r.returncode,round(rec['wall_seconds'],2),rec['failures'],flush=True);return rec
selectors=['^(Test.*View.*|TestInterfaceCast.*|TestCheckedCast.*|TestUncheckableCastAdmission|TestNarrowedUnion.*|TestNarrowedFieldUsesSharedReadiness)$','^(TestFractionalPowersReachRuntime|TestParseIntMapIndexRadixAgreesWithNode|TestReviewPrograms.*)$', '^TestNativeAgreesWithNode$/.*(cast_|view_|predicate|bitwise|truthy|tuple|logical|while|do_while).*']
mutants=[dict(id='D1',target='Tuple',file='internal/javascript/view_unions_untagged.go',old="  if(contract.FixedTuple && (!Array.isArray(value) || value.length !== contract.Tuple.length)) return false;\n",new='',change='drop fixed tuple array identity and length guard'),dict(id='D2',target='Producer',file='internal/javascript/view_unions_untagged.go',old='if !certified {',new='if certified {',change='flip callable producer certification bypass'),dict(id='D3',target='Moved',file='internal/lower/view_objects.go',old='func (l *lowering) structuralViewCast(node *ast.Node, value ir.Expression, source, target *checker.Type) (ir.Expression, error) {',new='func (l *lowering) structuralViewCast(node *ast.Node, value ir.Expression, source, target *checker.Type) (ir.Expression, error) {\n\treturn nil, nil',change='return early before untagged structural cast admission'),dict(id='D4',target='Moved',file='internal/javascript/javascript.go',old='ir.BitAnd: "&"',new='ir.BitAnd: "|"',change='change bitmask AND operator to OR'),dict(id='D5',target='Moved',file='internal/native/runtime/view_unions_untagged.c',old='child->of != adamic_rep_number && child->of != adamic_rep_boolean && child->of != adamic_rep_string',new='child->of != adamic_rep_boolean && child->of != adamic_rep_boolean && child->of != adamic_rep_string',change='change open kind selector number storage option to boolean')]
runs=[]
for i,s in enumerate(selectors):
 r=run('clean-'+str(i),s,cover=i<2)
 if r['exit']:raise RuntimeError('red bounded baseline')
# bounded rest coverage for untrue rows, excluding both requested rows.
run('rest-coverage','^(Test.*View.*|TestInterfaceCast.*|TestCheckedCast.*|TestUncheckableCastAdmission|TestNarrowedUnion.*|TestNarrowedFieldUsesSharedReadiness)$' .replace('Test.*View.*','Test(?!CheckedViewV2TupleIdentityMutant$|CheckedViewV2CallableProducerMutant$).*View.*'),cover=True) if False else None
for m in mutants:
 f=root/m['file'];orig=f.read_text();assert orig.count(m['old'])==1,(m,orig.count(m['old']))
 m['line']=orig[:orig.index(m['old'])].count('\n')+1;new=orig.replace(m['old'],m['new']);(p/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(orig.splitlines(True),new.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
 (p/'menu.json').write_text(json.dumps(mutants,indent=2));f.write_text(new)
 try:
  pkg='./'+str(pathlib.Path(m['file']).parent)+'/'
  if m['file'].endswith('.c'):pkg='./internal/native/'
  with (p/(m['id']+'-vet.log')).open('w') as out:v=subprocess.run(['go','vet',pkg],cwd=root,env=base,stdout=out,stderr=subprocess.STDOUT)
  if v.returncode:print('invalid',m['id'],flush=True);continue
  for i,s in enumerate(selectors):run(m['id']+'-'+str(i),s,mut=m['id'])
 finally:f.write_text(orig)
print('done',flush=True)
