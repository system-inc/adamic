import pathlib,subprocess,json,difflib,os,time
p=pathlib.Path('review/test-defend/internal-lower-arguments_length'); menu=[]
def add(id,file,old,new,rows,lead,kind):
 s=pathlib.Path(file).read_text();assert old in s,(id,old);menu.append(dict(id=id,file=file,line=s[:s.index(old)].count('\n')+1,old=old,new=new,rows=rows,lead=lead,kind=kind))
A='internal/lower/arguments_length.go';L='internal/lower/library_array_predicate.go';I='internal/lower/invariance.go';C='internal/lower/census_small.go'
r1='TestArgumentsLengthRefusals';r2='TestArgumentsLengthRefusalFixtures';ap='TestArrayPredicatePreservesDeclaredElementContract';ai='TestArrayPredicateCannotInventAnElementContract';at='TestArrayPredicateDoesNotMisclassifyNativeTuples';au='TestUnknownArrayPredicateRefusesUnrepresentedObservations';ar='TestArrayPredicateCoexistsWithUnknownReflection'
add('D01',A,'\t\tfor length.Parent != nil && length.Parent.Kind == ast.KindParenthesizedExpression {\n\t\t\tlength = length.Parent\n\t\t}\n','',[r1],'Exclusive block 42-43: only refusal row tests parenthesized write','drop whole loop')
add('D02',A,'What: "writing arguments.length"','What: "writing the argument count"',[r1,r2],'Writing refusals share line 47, unlike read neighbors','change constant')
add('D03',A,'What: "arguments inside an arrow function"','What: "arguments inside a closure"',[r1,r2],'Arrow refusals share line 30, unlike neighbors capturing count','change constant')
add('D04',A,'What: "arguments other than a read of arguments.length"','What: "arguments object usage"',[r1,r2],'Object usages share line 49; fixture includes both extensions','change constant')
add('D05',L,'l.checker.IsArrayType(member) == array','l.checker.IsArrayType(member) != array',[ap,ai],'Declared union members versus unknown reflection: exclusive block 61-62','flip condition')
add('D06','internal/lower/predicates.go','p.l.checker.IsTypeAssignableTo(narrowed, p.target)','p.l.checker.IsTypeAssignableTo(p.target, narrowed)',[ai],'Invented element contract reverses target/source relation on shared proof lines','swap arguments')
add('D07',L,'predicateReadonlyArray(l.checker, member, true)','predicateReadonlyArray(l.checker, member, false)',[ap,ai,ar],'Mutable unknown[] refusal versus readonly unknown[] acceptance','change option')
add('D08',L,'if proven.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsTypeParameter|checker.TypeFlagsIntersection) != 0 || checker.IsTupleType(proven) {\n\t\treturn false','if proven.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsTypeParameter|checker.TypeFlagsIntersection) != 0 || checker.IsTupleType(proven) {\n\t\treturn true',[at],'Tuple branch exclusive versus reflection; admit unrepresented domain','change constant')
add('D09',L,'"Array.isArray on a tuple or an erased object/any/unknown view"','"Array.isArray on an unsupported representation"',[at],'Exclusive refusal block27, exact tuple diagnostic requirement','change constant')
add('D10','internal/lower/object.go','return 0, l.notYet(node, "an array of unknown with erased element storage (retain its declared element type before reading elements)")','return ir.Union, nil',[au],'Unknown indexed-read refusal reached by untrue row, audit never mutated elementType','return early')
add('D11','internal/lower/expression.go','"an overloaded function as a value"','"a callable overload value"',['TestOverloadedShorthandFunctionValueStaysNotYet'],'Untrue row pins this boundary while ordinary overload users call directly','change constant')
add('D12',I,'return &widening{source: source, target: target}','return nil',['TestCensusMarkerResultIsAssignable'],'Exclusive result-incompatible branch121, zero-call marker has compatible void result','return early')
add('D13',I,'return &widening{source: takes, target: given, parameter: true}','return &widening{source: takes, target: given, parameter: false}',['TestOptionalFunctionValueRelation'],'Exclusive widening failure versus accepted optional neighbors; classification flag','change constant')
add('D14',C,'" cannot satisfy implementation parameter "','" cannot satisfy overload parameter "',['TestCensusOverloadBinderGuards'],'Binder guard pins relation role whereas subsumer pins overload prefix','change constant')
add('D15','internal/lower/unknown.go','Name: key.Text()}, nil','Name: ""}, nil',[ar],'Exclusive inProperty return block67: reflection checks presence of code','change constant')
(p/'menu.json').write_text(json.dumps(menu,indent=2)+'\n')
(p/'code-and-oracles.md').write_text('''Code under test: Adamic Lower, specifically argumentsRefusal, arrayPredicateMembers, arrayPredicateDomain, arrayIsArray, predicateFlowProof.returned, elementType, functionValue, widened, censusOverload and inProperty. Load prepares typed input; no preparation, oracle, harness or test is mutated.
Oracles: arguments refusals/fixtures, invented contracts, tuple refusals, unknown observations, marker result, optional relation, shorthand overload and binder guards use self-pinned diagnostic types/text. PreservesDeclaredElementContract and CoexistsWithUnknownReflection now execute source and lowered JavaScript on Node and compare stdout, stderr and exit (external-run); the old audit called their former acceptance-only assertions self. Coverage uses only internal/lower as coverpkg.
Bounded scope: full package timed out at90s without an assertion failure. The green102-test matrix includes whole test files mentioning arguments, Array.isArray, never-rest markers, overloads or optional parameters. It includes newly added tests in these files. Outside-row catches are unknown; uniqueness supports only a bounded defense, pending replay. No test may be deleted from these findings.
''')
if '--plan' in __import__('sys').argv:print('planned',len(menu));raise SystemExit
originals={m['file']:pathlib.Path(m['file']).read_text() for m in menu};runs=[]
try:
 for m in menu:
  id=m['id'];file=m['file'];old=originals[file];new=old.replace(m['old'],m['new'],1);diff=''.join(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='a/'+file,tofile='b/'+file));(p/(id+'.diff')).write_text(diff)
  subprocess.run(['git','apply','--check',str(p/(id+'.diff'))],check=True)
  pathlib.Path(file).write_text(new)
  with (p/(id+'-vet.log')).open('w') as log:vet=subprocess.run(['timeout','90','go','vet','./internal/lower/'],stdout=log,stderr=subprocess.STDOUT)
  assert vet.returncode==0,id
  env={**os.environ,'ADAMIC_BUILD_CACHE_DIR':'/tmp/defend-arguments/cache/'+id}
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run',(p/'matrix.regex').read_text()];start=time.monotonic()
  with (p/(id+'.log')).open('w') as log:r=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
  pathlib.Path(file).write_text(old)
  es=[]
  for l in (p/(id+'.log')).read_text().splitlines():
   try:es.append(json.loads(l))
   except:pass
  failed=sorted({e['Test'].split('/')[0] for e in es if e.get('Test') and e['Action']=='fail'});passed=sorted({e['Test'] for e in es if e.get('Test') and '/' not in e['Test'] and e['Action']=='pass'});errors=[e for e in es if e.get('OutputType')=='error'];run=dict(id=id,command=cmd,environment={'ADAMIC_BUILD_CACHE_DIR':env['ADAMIC_BUILD_CACHE_DIR']},exit=r.returncode,seconds=time.monotonic()-start,rows_failed=failed,rows_passed=passed,errors=errors)
  runs.append(run);(p/'runs.json').write_text(json.dumps(runs,indent=2)+'\n');print(id,failed,'passed',len(passed),flush=True)
finally:
 for f,s in originals.items():pathlib.Path(f).write_text(s)
