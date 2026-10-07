import json, os, pathlib, subprocess, sys

repo = pathlib.Path(__file__).resolve().parents[6]
cohere = repo / 'cohere'
out = pathlib.Path(sys.argv[1])
out.mkdir(parents=True, exist_ok=True)
original = (cohere / 'internal/lint/ecmascript/control_flow_graph/cfg.go').read_text()
anchor = 'func (b *Builder[E]) reachedStatement(node *ast.Node) {'
replacement = '''func (b *Builder[E]) reachedStatement(node *ast.Node) {
 originalHook := b.hooks.Statement
 before := adamicStatementInput(b,node,originalHook != nil)
 calls,kind,pos,end,current := 0,-1,-1,-1,-1
 sameBuilder,sameNode := false,false
 if originalHook != nil {
  b.hooks.Statement = func(actual *Builder[E],handed *ast.Node) {
   calls++;kind,pos,end=adamicStatementNode(handed)
   current=-1;if actual.cur != nil {current=int(actual.cur.index)}
   sameBuilder=actual==b;sameNode=handed==node
   originalHook(actual,handed)
  }
 }
 defer func(){
  b.hooks.Statement=originalHook
  adamicObserveStatement(before,calls,kind,pos,end,current,sameBuilder,sameNode)
 }()'''
assert original.count(anchor) == 1
(out / 'cfg.go').write_text(original.replace(anchor, replacement))
observer = '''package control_flow_graph
import("fmt";"os";"sync";"github.com/microsoft/TypeScript/tsc/shim/ast")
var adamicStatementMutex sync.Mutex
func adamicStatementBit(v bool)int{if v{return 1};return 0}
func adamicStatementNode(node *ast.Node)(int,int,int){
 if node==nil{return -1,-1,-1};return int(node.Kind),int(node.Pos()),int(node.End())
}
func adamicStatementInput[E any](b *Builder[E],node *ast.Node,hook bool)string{
 kind,pos,end:=adamicStatementNode(node);current:=-1;incoming,reachable:=false,false
 if b.cur!=nil{current=int(b.cur.index);incoming=b.cur.hasIncoming;reachable=b.cur.Reachable}
 return fmt.Sprintf("%d\\t%d\\t%d\\t%d\\t%d\\t%d\\t%d",kind,pos,end,current,adamicStatementBit(incoming),adamicStatementBit(reachable),adamicStatementBit(hook))
}
func adamicObserveStatement(input string,calls,kind,pos,end,current int,sameBuilder,sameNode bool){
 adamicStatementMutex.Lock();defer adamicStatementMutex.Unlock()
 path:=os.Getenv("ADAMIC_STATEMENT_CAPTURE");if path==""{return}
 f,e:=os.OpenFile(path+".input",os.O_APPEND|os.O_CREATE|os.O_WRONLY,0600);if e!=nil{panic(e)}
 fmt.Fprintln(f,input);f.Close()
 f,e=os.OpenFile(path+".want",os.O_APPEND|os.O_CREATE|os.O_WRONLY,0600);if e!=nil{panic(e)}
 fmt.Fprintf(f,"%d\\t%d\\t%d\\t%d\\t%d\\t%d\\t%d\\n",calls,kind,pos,end,current,adamicStatementBit(sameBuilder),adamicStatementBit(sameNode));f.Close()
}'''
(out / 'observer.go').write_text(observer)
controls = '''package control_flow_graph
import("testing";"github.com/microsoft/TypeScript/tsc/shim/ast")
func TestAdamicStatementControls(t *testing.T){
 factory:=ast.NewNodeFactory(ast.NodeFactoryHooks{})
 for _,node:=range []*ast.Node{nil,factory.NewIdentifier("handed"),factory.NewEmptyStatement()}{
  for current:=-1;current<2;current++{for incoming:=0;incoming<2;incoming++{for reachable:=0;reachable<2;reachable++{for hook:=0;hook<2;hook++{
   b:=&Builder[int]{}
   if current>=0{b.cur=&Block[int]{index:int32(current),hasIncoming:incoming==1,Reachable:reachable==1}}
   if hook==1{b.hooks.Statement=func(actual *Builder[int],handed *ast.Node){}}
   b.reachedStatement(node)
  }}}}
 }
}'''
(out / 'controls_test.go').write_text(controls)
virtual = cohere / 'internal/lint/ecmascript/control_flow_graph'
replace = {str(virtual / 'cfg.go'): str(out / 'cfg.go'),
           str(virtual / 'adamic_statement_observer.go'): str(out / 'observer.go'),
           str(virtual / 'adamic_statement_controls_test.go'): str(out / 'controls_test.go')}
(out / 'overlay.json').write_text(json.dumps({'Replace': replace}))
consumers = [('array-callback-return', 'core', 'TestArrayCallbackReturn'),
             ('consistent-return', 'core', 'TestConsistentReturn'),
             ('no-unreachable-loop', 'core', 'TestNoUnreachableLoop'),
             ('react-hooks/rules-of-hooks', 'react', 'TestRulesOfHooks'),
             ('controls', 'control_flow_graph', 'TestAdamicStatementControls')]
for label, package, test in consumers:
    env = os.environ.copy()
    env['ADAMIC_STATEMENT_CAPTURE'] = str(out / label.replace('/', '_'))
    target = './internal/lint/rules/' + package if label != 'controls' else './internal/lint/ecmascript/control_flow_graph'
    with (out / (label.replace('/', '_') + '.log')).open('w') as log:
        subprocess.run(['go', 'test', '-overlay=' + str(out / 'overlay.json'), target,
                        '-run', '^' + test, '-count=1', '-timeout=10m', '-v'],
                       cwd=cohere, env=env, stdout=log, stderr=subprocess.STDOUT, check=True)
    assert pathlib.Path(env['ADAMIC_STATEMENT_CAPTURE'] + '.input').exists(), label + ' had no calls'
