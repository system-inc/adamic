import json, os, pathlib, subprocess, sys

repo = pathlib.Path(__file__).resolve().parents[6]
cohere = repo / 'cohere'
out = pathlib.Path(sys.argv[1])
out.mkdir(parents=True, exist_ok=True)
original = (cohere / 'internal/lint/ecmascript/control_flow_graph/helpers.go').read_text()
anchor = 'func isThrowableIdentifier(node *ast.Node) bool {'
replacement = '''func isThrowableIdentifier(node *ast.Node) (result bool) {
 defer func(){ adamicObserveThrowable(node,result) }()'''
assert original.count(anchor) == 1
(out / 'helpers.go').write_text(original.replace(anchor, replacement))
observer = '''package control_flow_graph
import("fmt";"os";"sync";"github.com/microsoft/TypeScript/tsc/shim/ast")
var adamicThrowableMutex sync.Mutex
func adamicObserveThrowable(node *ast.Node,result bool){
 adamicThrowableMutex.Lock();defer adamicThrowableMutex.Unlock()
 path:=os.Getenv("ADAMIC_THROWABLE_CAPTURE");if path==""{return}
 parentKind,grandparentKind:=-1,-1
 tagName,rest,property,name:=false,false,false,false
 if parent:=node.Parent;parent!=nil{
  parentKind=int(parent.Kind);name=parent.Name()==node
  if parent.Parent!=nil{grandparentKind=int(parent.Parent.Kind)}
  switch parent.Kind{
  case ast.KindJsxOpeningElement,ast.KindJsxClosingElement,ast.KindJsxSelfClosingElement:tagName=parent.TagName()==node
  case ast.KindBindingElement:rest=parent.AsBindingElement().DotDotDotToken!=nil;property=parent.AsBindingElement().PropertyName==node
  }
 }
 bit:=func(v bool)int{if v{return 1};return 0}
 f,e:=os.OpenFile(path+".input",os.O_APPEND|os.O_CREATE|os.O_WRONLY,0600);if e!=nil{panic(e)}
 fmt.Fprintf(f,"%d\\t%d\\t%d\\t%d\\t%d\\t%d\\n",parentKind,bit(tagName),bit(rest),bit(property),grandparentKind,bit(name));f.Close()
 f,e=os.OpenFile(path+".want",os.O_APPEND|os.O_CREATE|os.O_WRONLY,0600);if e!=nil{panic(e)}
 fmt.Fprintln(f,bit(result));f.Close()
}'''
(out / 'observer.go').write_text(observer)
controls = r'''package control_flow_graph
import("testing";"github.com/microsoft/TypeScript/tsc/shim/ast";"github.com/microsoft/TypeScript/tsc/shim/core";"github.com/microsoft/TypeScript/tsc/shim/parser")
func TestAdamicThrowableControls(t *testing.T){
 codes:=[]string{
  `import main, { original as alias, plain } from "module"; import * as space from "module"; export { original as alias, plain }; export * as external from "module";`,
  `outer: for (let i=0;i<limit;i++) { if (skip) continue outer; if (stop) break outer; work(i); }`,
  `const { plain, source: renamed, defaulted = fallback, ...rest } = object; const [ item, ...remaining ] = array;`,
  `try { work(); } catch (problem) { throw problem; }`,
  `function named(parameter = fallback) { return parameter; } const expression = function internal() { return external; }; const arrow = argument => argument;`,
  `class Named { property = value; method() { return value; } get getter() { return value; } set setter(argument) { value = argument; } } const expression = class Inner {}; const object = { property: value, method() {} };`,
  `enum Enumeration { Member } namespace Module { export const value = external; } interface Interface { property: Type } type Alias = Type;`,
  `const element = <Widget value={external} ns:attribute={external}><Other /></Widget>; const nested = <Space.Widget />;`,
  `const { [computed]: result = fallback } = object; const [ defaulted = fallback ] = array;`,
 }
 for _,code:=range codes{
  source:=parser.ParseSourceFile(ast.SourceFileParseOptions{FileName:"/controls.tsx",Path:"/controls.tsx"},code,core.ScriptKindTSX)
  var walk func(*ast.Node)
  walk=func(node *ast.Node){if node.Kind==ast.KindIdentifier{isThrowableIdentifier(node)};node.ForEachChild(func(child *ast.Node)bool{walk(child);return false})}
  walk(source.AsNode())
 }
 factory:=ast.NewNodeFactory(ast.NodeFactoryHooks{})
 isThrowableIdentifier(factory.NewIdentifier("detached"))
}'''
(out / 'controls_test.go').write_text(controls)
virtual = cohere / 'internal/lint/ecmascript/control_flow_graph'
replace = {str(virtual / 'helpers.go'): str(out / 'helpers.go'),
           str(virtual / 'adamic_throwable_observer.go'): str(out / 'observer.go'),
           str(virtual / 'adamic_throwable_controls_test.go'): str(out / 'controls_test.go')}
(out / 'overlay.json').write_text(json.dumps({'Replace': replace}))
consumers = [('array-callback-return', 'core', 'TestArrayCallbackReturn'),
             ('consistent-return', 'core', 'TestConsistentReturn'),
             ('no-unreachable-loop', 'core', 'TestNoUnreachableLoop'),
             ('react-hooks/rules-of-hooks', 'react', 'TestRulesOfHooks'),
             ('controls', 'control_flow_graph', 'TestAdamicThrowableControls')]
for label, package, test in consumers:
    env = os.environ.copy()
    env['ADAMIC_THROWABLE_CAPTURE'] = str(out / label.replace('/', '_'))
    target = './internal/lint/rules/' + package if label != 'controls' else './internal/lint/ecmascript/control_flow_graph'
    with (out / (label.replace('/', '_') + '.log')).open('w') as log:
        subprocess.run(['go', 'test', '-overlay=' + str(out / 'overlay.json'), target,
                        '-run', '^' + test, '-count=1', '-timeout=10m', '-v'],
                       cwd=cohere, env=env, stdout=log, stderr=subprocess.STDOUT, check=True)
    assert pathlib.Path(env['ADAMIC_THROWABLE_CAPTURE'] + '.input').exists(), label + ' had no calls'
