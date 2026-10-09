import pathlib,json,time,subprocess,os
R=pathlib.Path('/workspace/adamic');E=R/'review/test-audit/internal-lower-predicates_overload';P=R/'internal/lower';base=json.load(open(E/'base.json'));plans=json.load(open(E/'plan.json'))
fixture='''package lower
import("testing";"context";"os";"path/filepath";"github.com/system-inc/adamic/internal/load";"github.com/system-inc/adamic/internal/ir";"github.com/microsoft/TypeScript/tsc/shim/ast")
func TestAuditMutationWitness(t *testing.T){path:=filepath.Join(t.TempDir(),"probe.a");source:=`function some(xs: readonly number[] | undefined): xs is readonly number[]; function some(xs: readonly number[] | undefined): boolean {return xs !== undefined;} function use(xs: readonly number[] | undefined): void {if(some(xs)){} xs=[1]; console.log(xs.length.toString());}`;if err:=os.WriteFile(path,[]byte(source),0600);err!=nil{t.Fatal(err)};program,err:=load.Load([]string{path});if err!=nil{t.Fatal(err)};file:=program.Files()[0];checked,release:=program.Checker(context.Background(),file);defer release();l:=&lowering{program:program,checker:checked,result:&ir.Program{}};var call *ast.CallExpression;var assigned *ast.Node;var visit ast.Visitor;visit=func(n *ast.Node)bool{if n.Kind==ast.KindCallExpression && n.Expression().Kind==ast.KindIdentifier && n.Expression().Text()=="some"{call=n.AsCallExpression()};if n.Kind==ast.KindIdentifier && n.Text()=="xs" && ast.IsAssignmentTarget(n){assigned=n};n.ForEachChild(visit);return false};file.AsNode().ForEachChild(visit);if call==nil||assigned==nil{t.Fatal("missing witness nodes")};flow:=&ast.FlowNode{Flags:ast.FlowFlagsAssignment,Node:assigned};t.Logf("replacement helper directions=%d; entry directions=%d",l.predicateFlowDirections(flow,call.AsNode(),assigned,map[*ast.FlowNode]bool{}),l.predicateUseDirections(call))}
'''
(E/'witness.go.fixture').write_text(fixture)
while not (E/'probes.done').exists():time.sleep(1)
record=next(x for x in json.load(open(E/'matrix-time.json')) if x['id']=='M18')
if record['exit']==0:
 target=P/'audit_witness_test.go';target.write_text(fixture)
 try:
  for id in ['clean','M18']:
   p=next(p for p in plans if p['id']=='M18');f=p['file'].split('/')[-1];s=base[f]
   if id=='M18':(P/f).write_text(s[:p['pos']]+p['new']+s[p['pos']+len(p['old']):])
   with (E/('witness-fixed-'+id+'.log')).open('w') as log:subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^TestAuditMutationWitness$'],cwd=R,stdout=log,stderr=subprocess.STDOUT,env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/u042/cache/witness-'+id))
   (P/f).write_text(s)
 finally:target.unlink(missing_ok=True)
(E/'witness.done').write_text('done')
