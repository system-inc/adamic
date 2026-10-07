import json, os, pathlib, subprocess, sys

repo = pathlib.Path(__file__).resolve().parents[6]
cohere = repo / 'cohere'
out = pathlib.Path(sys.argv[1])
out.mkdir(parents=True, exist_ok=True)
original = (cohere / 'internal/lint/ecmascript/control_flow_graph/cfg.go').read_text()
anchor = 'func (b *Builder[E]) enterDisconnected(blk *Block[E]) {'
replacement = '''func (b *Builder[E]) enterDisconnected(blk *Block[E]) {
 previous := b.cur
 incoming, before, priorReachable := blk.hasIncoming, blk.Reachable, previous.Reachable
 defer func() {
  adamicObserveDisconnected(int(blk.index), incoming, before, int(previous.index), priorReachable,
   blk.hasIncoming, blk.Reachable, int(b.cur.index), b.cur == blk, previous.Reachable)
 }()'''
assert original.count(anchor) == 1
(out / 'cfg.go').write_text(original.replace(anchor, replacement))
observer = '''package control_flow_graph
import("fmt";"os";"sync")
var adamicDisconnectedMutex sync.Mutex
func adamicObserveDisconnected(id int,incoming,before bool,previous int,priorReachable,afterIncoming,reachable bool,current int,same,afterPrevious bool){
 adamicDisconnectedMutex.Lock();defer adamicDisconnectedMutex.Unlock()
 path:=os.Getenv("ADAMIC_DISCONNECTED_CAPTURE");if path==""{return}
 bit:=func(v bool)int{if v{return 1};return 0}
 f,e:=os.OpenFile(path+".input",os.O_APPEND|os.O_CREATE|os.O_WRONLY,0600);if e!=nil{panic(e)}
 fmt.Fprintf(f,"%d\\t%d\\t%d\\t%d\\t%d\\n",id,bit(incoming),bit(before),previous,bit(priorReachable));f.Close()
 f,e=os.OpenFile(path+".want",os.O_APPEND|os.O_CREATE|os.O_WRONLY,0600);if e!=nil{panic(e)}
 fmt.Fprintf(f,"%d\\t%d\\t%d\\t%d\\t%d\\n",bit(afterIncoming),bit(reachable),current,bit(same),bit(afterPrevious));f.Close()
}'''
(out / 'observer.go').write_text(observer)
controls = '''package control_flow_graph
import "testing"
func TestAdamicDisconnectedControls(t *testing.T){
 for id:=0;id<5;id++{for incoming:=0;incoming<2;incoming++{for before:=0;before<2;before++{
  node:=&Block[int]{index:int32(id),hasIncoming:incoming==1,Reachable:before==1}
  (&Builder[int]{cur:node}).enterDisconnected(node)
  for previous:=0;previous<5;previous++{if previous==id{continue};for reachable:=0;reachable<2;reachable++{
   node:=&Block[int]{index:int32(id),hasIncoming:incoming==1,Reachable:before==1}
   b:=&Builder[int]{cur:&Block[int]{index:int32(previous),Reachable:reachable==1}}
   b.enterDisconnected(node)
  }}
 }}}
}'''
(out / 'controls_test.go').write_text(controls)
virtual = cohere / 'internal/lint/ecmascript/control_flow_graph'
replace = {str(virtual / 'cfg.go'): str(out / 'cfg.go'),
           str(virtual / 'adamic_disconnected_observer.go'): str(out / 'observer.go'),
           str(virtual / 'adamic_disconnected_controls_test.go'): str(out / 'controls_test.go')}
core_controls = '''package core
import("testing";"github.com/system-inc/cohere/internal/lint/testing")
func TestArrayCallbackReturnAdamicDisconnected(t *testing.T){
 for _,body:=range []string{"work();", "if (flag) return 1;", "if (flag) break;", "if (flag) continue;", "throw problem;"}{
  rule_testing.Run(t,ArrayCallbackReturn,"disconnected.ts","items.map(() => { for (let i=0;i<2;i++) {"+body+"} return 2; });")
 }
}
func TestConsistentReturnAdamicDisconnected(t *testing.T){
 for _,body:=range []string{"work();", "if (flag) return 1;", "if (flag) break;", "if (flag) continue;", "throw problem;"}{
  rule_testing.Run(t,ConsistentReturn,"disconnected.ts","function work() { for (let i=0;i<2;i++) {"+body+"} return 2; }")
 }
}
func TestNoUnreachableLoopAdamicDisconnected(t *testing.T){
 for _,body:=range []string{"work();", "if (flag) return 1;", "if (flag) break;", "if (flag) continue;", "throw problem;"}{
  rule_testing.Run(t,NoUnreachableLoop,"disconnected.ts","function work() { for (let i=0;i<2;i++) {"+body+"} return 2; }")
 }
}'''
react_controls = '''package react
import("testing";"github.com/system-inc/cohere/internal/lint/testing")
func TestRulesOfHooksAdamicDisconnected(t *testing.T){
 for _,body:=range []string{"work();", "if (flag) return 1;", "if (flag) break;", "if (flag) continue;", "throw problem;"}{
  rule_testing.Run(t,RulesOfHooks,"disconnected.ts","function Component() { useHook(); for (let i=0;i<2;i++) {"+body+"} return 2; }")
 }
}'''
(out / 'core_controls_test.go').write_text(core_controls)
(out / 'react_controls_test.go').write_text(react_controls)
for package in ['core', 'react']:
    replace[str(cohere / 'internal/lint/rules' / package / 'adamic_disconnected_test.go')] = str(out / (package + '_controls_test.go'))
(out / 'overlay.json').write_text(json.dumps({'Replace': replace}))
consumers = [('array-callback-return', 'core', 'TestArrayCallbackReturn'),
             ('consistent-return', 'core', 'TestConsistentReturn'),
             ('no-unreachable-loop', 'core', 'TestNoUnreachableLoop'),
             ('react-hooks/rules-of-hooks', 'react', 'TestRulesOfHooks'),
             ('controls', 'control_flow_graph', 'TestAdamicDisconnectedControls')]
for label, package, test in consumers:
    env = os.environ.copy()
    env['ADAMIC_DISCONNECTED_CAPTURE'] = str(out / label.replace('/', '_'))
    target = './internal/lint/rules/' + package if label != 'controls' else './internal/lint/ecmascript/control_flow_graph'
    with (out / (label.replace('/', '_') + '.log')).open('w') as log:
        subprocess.run(['go', 'test', '-overlay=' + str(out / 'overlay.json'), target,
                        '-run', '^' + test, '-count=1', '-timeout=10m', '-v'],
                       cwd=cohere, env=env, stdout=log, stderr=subprocess.STDOUT, check=True)
    assert pathlib.Path(env['ADAMIC_DISCONNECTED_CAPTURE'] + '.input').exists(), label + ' had no calls'
