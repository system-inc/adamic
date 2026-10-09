import pathlib,json,re,subprocess,time,os,difflib
R=pathlib.Path('/workspace/adamic');E=R/'review/test-audit/internal-lower-predicates_overload';P=R/'internal/lower';base={f.name:f.read_text() for f in P.glob('*.go') if not f.name.endswith('_test.go')};rows='TestPredicateOverloadRuntime TestIndirectPredicateOverloadIsPending TestPredicateBodyProof TestConditionAssertionAdmission TestPredicateCallbackContracts TestEveryNeedsCallbackEffects TestPredicateOverloadCallback TestPredicateUseRegions TestPredicateUsesBelongToEachCall TestUnprovenPredicateReturnsAreRefused TestPredicateBodiesAreProven TestPrimitiveAdmittingSlotsUseBoxes'.split();listed=(E/'test-list.log').read_text().splitlines();assert all(r in listed for r in rows)
(E/'scope.json').write_text(json.dumps(rows));(E/'base.json').write_text(json.dumps(base));(E/'functions.txt').write_text('\n'.join(re.findall(r'^func .*','\n'.join(base.values()),re.M))+'\nConservative complete production function inventory; mutation sites below are in scoped proof, lowering and representation chains.\n');plans=[]
def add(id,f,old,new,kind,nth=0):
 s=base[f];positions=[m.start() for m in re.finditer(re.escape(old),s)];assert len(positions)>nth,(id,old);pos=positions[nth];plans.append(dict(id=id,file='internal/lower/'+f,line=s[:pos].count('\n')+1,pos=pos,old=old,new=new,kind=kind))
def a(n,f,o,v,k='flip condition',nth=0):add('M%02d'%n,f,o,v,k,nth)
a(1,'predicates.go','if changed != nil {','if changed == nil {')
a(2,'predicates.go','literal && truth != (expression.Kind == ast.KindTrueKeyword)','literal && truth == (expression.Kind == ast.KindTrueKeyword)')
a(3,'predicates.go','case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:\n\t\t// A property read may dispatch a getter through a structural view.\n\t\treturn true','case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:\n\t\t// A property read may dispatch a getter through a structural view.\n\t\treturn false','change constant')
a(4,'predicates.go','flags = ast.FlowFlagsTrueCondition','flags = ast.FlowFlagsFalseCondition','change option')
a(5,'predicates.go','if original == nil {','if original != nil {')
a(6,'predicates.go','if signature == nil {\n\t\treturn nil','if signature != nil {\n\t\treturn nil')
a(7,'predicates.go','return found && safe','return found && !safe')
a(8,'predicates_proof.go','proof.TaggedView = true','proof.TaggedView = false','change constant')
a(9,'predicates_proof.go','changed = true','changed = false','change constant')
a(10,'predicates_proof.go','(t&predicateTrue)<<1','(t&predicateTrue)>>1','change option')
a(11,'predicates_proof.go','if and {','if !and {')
a(12,'predicates_proof.go','yes = v.cells[path.cell] == constant.Text()','yes = v.cells[path.cell] != constant.Text()')
a(13,'predicates_proof.go','case ast.KindThrowStatement:\n\t\treturn true','case ast.KindThrowStatement:\n\t\treturn false','change constant')
a(14,'predicates_proof.go','cells: []string{"truthy", "falsy"}','cells: []string{"truthy"}','change option')
a(15,'predicates_proof.go','return proof.proveBody(overload.Type()) == nil','return proof.proveBody(overload.Type()) != nil')
a(16,'predicates_proof.go','ir.Binary{Operator: ir.NotEqual, Left: returned, Right: membership}','ir.Binary{Operator: ir.Equal, Left: returned, Right: membership}','change option')
a(17,'predicates_proof.go','counts.Unobservable++','','drop statement')
a(18,'predicates_proof.go','if flow.Flags&ast.FlowFlagsAssignment != 0 && l.predicateSameReference(flow.Node, reference) && predicateWriteOnly(flow.Node) {\n\t\treturn 0','if flow.Flags&ast.FlowFlagsAssignment != 0 && l.predicateSameReference(flow.Node, reference) && predicateWriteOnly(flow.Node) {\n\t\treturn predicateEither','change constant')
a(19,'expression.go','if flags&(checker.TypeFlagsUnknown|checker.TypeFlagsNonPrimitive) != 0 {\n\t\treturn ir.Union, true','if flags&(checker.TypeFlagsUnknown|checker.TypeFlagsNonPrimitive) != 0 {\n\t\treturn ir.Object, true','change constant')
a(20,'expression.go','// the runtime brand with the same boxes used for scalar/reference unions.\n\t\treturn ir.Union, true','// the runtime brand with the same boxes used for scalar/reference unions.\n\t\treturn ir.Object, true','change constant')
(E/'plan.json').write_text(json.dumps(plans,indent=2));(E/'diffs').mkdir(exist_ok=True)
for p in plans:
 f=p['file'].split('/')[-1];s=base[f];changed=s[:p['pos']]+p['new']+s[p['pos']+len(p['old']):];(E/'diffs'/(p['id']+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),changed.splitlines(True),fromfile='a/'+p['file'],tofile='b/'+p['file'])))
for row in rows:
 for n in range(3):
  with (E/f'timing-{row}-{n}.log').open('w') as log:subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^'+row+'$'],cwd=R,stdout=log,stderr=subprocess.STDOUT,env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/u042/cache/timing'))
(E/'timing.done').write_text('done')
