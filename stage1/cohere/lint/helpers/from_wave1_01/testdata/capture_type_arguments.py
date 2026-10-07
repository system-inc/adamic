import json, os, pathlib, subprocess, sys

repo = pathlib.Path(__file__).resolve().parents[6]
cohere = repo / 'cohere'
out = pathlib.Path(sys.argv[1])
out.mkdir(parents=True, exist_ok=True)
original = (cohere / 'internal/lint/ecmascript/control_flow_graph/expressions.go').read_text()
anchor = '''func (b *Builder[E]) typeArguments(node *ast.Node) {
	for _, typeArg := range node.TypeArguments() {
		b.expr(typeArg)
	}
}'''
replacement = '''func (b *Builder[E]) typeArguments(node *ast.Node) {
 ids:=map[*ast.Node]int{nil:-1}
 input:=[]string{}
 for index,arg:=range node.TypeArguments(){
  if _,ok:=ids[arg];!ok{ids[arg]=index}
  input=append(input,adamicTypeArgumentInput(arg,ids))
 }
 trace:=[]string{}
 defer func(){adamicObserveTypeArguments(input,trace)}()
 for _, typeArg := range node.TypeArguments() {
  adamicExpressionHandoff(b,typeArg,b,ids,&trace)
 }
}'''
assert original.count(anchor) == 1
(out / 'expressions.go').write_text(original.replace(anchor, replacement))
observer = '''package control_flow_graph
import("fmt";"os";"strings";"sync";"github.com/microsoft/TypeScript/tsc/shim/ast")
var adamicTypeArgumentsMutex sync.Mutex
func adamicTypeArgumentInput(node *ast.Node,ids map[*ast.Node]int)string{
 kind,pos,end:=-1,-1,-1
 if node!=nil{kind=int(node.Kind);pos=int(node.Pos());end=int(node.End())}
 return fmt.Sprintf("%d:%d:%d:%d",ids[node],kind,pos,end)
}
func adamicExpressionHandoff[E any](actual *Builder[E],node *ast.Node,original *Builder[E],ids map[*ast.Node]int,trace *[]string){
 same,present:=0,0;if actual==original{same=1};if _,ok:=ids[node];ok{present=1}
 *trace=append(*trace,fmt.Sprintf("%s:%d:%d",adamicTypeArgumentInput(node,ids),same,present))
 actual.expr(node)
}
func adamicObserveTypeArguments(input,trace []string){
 adamicTypeArgumentsMutex.Lock();defer adamicTypeArgumentsMutex.Unlock()
 path:=os.Getenv("ADAMIC_TYPE_ARGUMENTS_CAPTURE");if path==""{return}
 row:=strings.Join(input,",");if row==""{row="."}
 f,e:=os.OpenFile(path+".input",os.O_APPEND|os.O_CREATE|os.O_WRONLY,0600);if e!=nil{panic(e)}
 fmt.Fprintln(f,row);f.Close()
 f,e=os.OpenFile(path+".want",os.O_APPEND|os.O_CREATE|os.O_WRONLY,0600);if e!=nil{panic(e)}
 fmt.Fprintln(f,strings.Join(trace,","));f.Close()
}'''
(out / 'observer.go').write_text(observer)
controls = '''package control_flow_graph
import("testing";"github.com/microsoft/TypeScript/tsc/shim/ast")
func TestAdamicTypeArgumentsControls(t *testing.T){
 factory:=ast.NewNodeFactory(ast.NodeFactoryHooks{})
 name:=factory.NewIdentifier("Generic")
 choices:=[]*ast.Node{nil,factory.NewTypeReferenceNode(factory.NewIdentifier("First"),nil),factory.NewTypeReferenceNode(factory.NewIdentifier("Second"),nil)}
 (&Builder[int]{cur:&Block[int]{}}).typeArguments(factory.NewTypeReferenceNode(name,nil))
 for length:=0;length<=4;length++{
  count:=1;for i:=0;i<length;i++{count*=3}
  for code:=0;code<count;code++{
   args:=[]*ast.Node{};value:=code
   for i:=0;i<length;i++{args=append(args,choices[value%3]);value/=3}
   node:=factory.NewTypeReferenceNode(name,factory.NewNodeList(args))
   (&Builder[int]{cur:&Block[int]{}}).typeArguments(node)
  }
 }
}'''
(out / 'controls_test.go').write_text(controls)
virtual = cohere / 'internal/lint/ecmascript/control_flow_graph'
replace = {str(virtual / 'expressions.go'): str(out / 'expressions.go'),
           str(virtual / 'adamic_type_arguments_observer.go'): str(out / 'observer.go'),
           str(virtual / 'adamic_type_arguments_controls_test.go'): str(out / 'controls_test.go')}
core_controls = '''package core
import("testing";"github.com/system-inc/cohere/internal/lint/testing")
func TestArrayCallbackReturnAdamicTypeArguments(t *testing.T){
 for _,body:=range []string{"read<typeof first, typeof second>();", "new Factory<typeof first, typeof second>();", "tag<typeof first, typeof second>`value`;", "read?.<typeof first, typeof second>();"}{
  rule_testing.Run(t,ArrayCallbackReturn,"arguments.ts","items.map(() => {"+body+"return 1;});")
 }
}
func TestConsistentReturnAdamicTypeArguments(t *testing.T){
 for _,body:=range []string{"read<typeof first, typeof second>();", "new Factory<typeof first, typeof second>();", "tag<typeof first, typeof second>`value`;", "read?.<typeof first, typeof second>();"}{
  rule_testing.Run(t,ConsistentReturn,"arguments.ts","function work() {"+body+"return 1;}")
 }
}
func TestNoUnreachableLoopAdamicTypeArguments(t *testing.T){
 for _,body:=range []string{"read<typeof first, typeof second>();", "new Factory<typeof first, typeof second>();", "tag<typeof first, typeof second>`value`;", "read?.<typeof first, typeof second>();"}{
  rule_testing.Run(t,NoUnreachableLoop,"arguments.ts","function work() { for (let i=0;i<2;i++) {"+body+"} }")
 }
}'''
react_controls = '''package react
import("testing";"github.com/system-inc/cohere/internal/lint/testing")
func TestRulesOfHooksAdamicTypeArguments(t *testing.T){
 for _,body:=range []string{"read<typeof first, typeof second>();", "new Factory<typeof first, typeof second>();", "tag<typeof first, typeof second>`value`;", "read?.<typeof first, typeof second>();"}{
  rule_testing.Run(t,RulesOfHooks,"arguments.ts","function Component() { useHook();"+body+"return 1;}")
 }
}'''
(out / 'core_controls_test.go').write_text(core_controls)
(out / 'react_controls_test.go').write_text(react_controls)
for package in ['core', 'react']:
    replace[str(cohere / 'internal/lint/rules' / package / 'adamic_type_arguments_test.go')] = str(out / (package + '_controls_test.go'))
(out / 'overlay.json').write_text(json.dumps({'Replace': replace}))
consumers = [('array-callback-return', 'core', 'TestArrayCallbackReturn'),
             ('consistent-return', 'core', 'TestConsistentReturn'),
             ('no-unreachable-loop', 'core', 'TestNoUnreachableLoop'),
             ('react-hooks/rules-of-hooks', 'react', 'TestRulesOfHooks'),
             ('controls', 'control_flow_graph', 'TestAdamicTypeArgumentsControls')]
for label, package, test in consumers:
    env = os.environ.copy()
    env['ADAMIC_TYPE_ARGUMENTS_CAPTURE'] = str(out / label.replace('/', '_'))
    target = './internal/lint/rules/' + package if label != 'controls' else './internal/lint/ecmascript/control_flow_graph'
    with (out / (label.replace('/', '_') + '.log')).open('w') as log:
        subprocess.run(['go', 'test', '-overlay=' + str(out / 'overlay.json'), target,
                        '-run', '^' + test, '-count=1', '-timeout=10m', '-v'],
                       cwd=cohere, env=env, stdout=log, stderr=subprocess.STDOUT, check=True)
    captured = pathlib.Path(env['ADAMIC_TYPE_ARGUMENTS_CAPTURE'] + '.input')
    assert captured.exists(), label + ' had no calls'
    assert any(row != '.' for row in captured.read_text().splitlines()), label + ' had no non-empty argument calls'
