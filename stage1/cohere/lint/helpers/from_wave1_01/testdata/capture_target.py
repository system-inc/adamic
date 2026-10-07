import json,os,pathlib,subprocess,sys
repo=pathlib.Path(__file__).resolve().parents[6];cohere=repo/'cohere'
out=pathlib.Path(sys.argv[1]);out.mkdir(parents=True,exist_ok=True)
original=(cohere/'internal/lint/ecmascript/control_flow_graph/cfg.go').read_text()
anchor='func throwTarget[E any](f *tryFrame[E]) *Block[E] {'
replacement='''func throwTarget[E any](f *tryFrame[E]) (target *Block[E]) {
 defer func(){adamicObserveTarget(int(f.position),f.catchEntry,f.finallyEntry,target)}()'''
assert original.count(anchor)==1
(out/'cfg.go').write_text(original.replace(anchor,replacement))
observer='''package control_flow_graph
import("fmt";"os";"sync")
var adamicTargetMutex sync.Mutex
func adamicObserveTarget[E any](position int,caught,final,target *Block[E]){
 adamicTargetMutex.Lock();defer adamicTargetMutex.Unlock()
 path:=os.Getenv("ADAMIC_TARGET_CAPTURE");if path==""{return}
 id:=func(b *Block[E])int{if b==nil{return -1};return int(b.index)}
 bit:=func(value bool)int{if value{return 1};return 0}
 f,e:=os.OpenFile(path+".input",os.O_APPEND|os.O_CREATE|os.O_WRONLY,0600);if e!=nil{panic(e)}
 fmt.Fprintf(f,"%d\\t%d\\t%d\\n",position,id(caught),id(final));f.Close()
 f,e=os.OpenFile(path+".want",os.O_APPEND|os.O_CREATE|os.O_WRONLY,0600);if e!=nil{panic(e)}
 fmt.Fprintf(f,"%d\\t%d\\t%d\\n",id(target),bit(target!=nil&&target==caught),bit(target!=nil&&target==final));f.Close()
}'''
(out/'observer.go').write_text(observer)
controls='''package control_flow_graph
import "testing"
func TestAdamicTargetControls(t *testing.T){
 blocks:=[]*Block[int]{nil,{index:0},{index:1},{index:2}}
 for position:=0;position<256;position++{for caught:=0;caught<len(blocks);caught++{for final:=0;final<len(blocks);final++{
 throwTarget(&tryFrame[int]{position:tryPosition(position),catchEntry:blocks[caught],finallyEntry:blocks[final]})
 }}}
}'''
(out/'controls_test.go').write_text(controls)
virtual=cohere/'internal/lint/ecmascript/control_flow_graph'
replace={str(virtual/'cfg.go'):str(out/'cfg.go'),str(virtual/'adamic_target_observer.go'):str(out/'observer.go'),str(virtual/'adamic_target_controls_test.go'):str(out/'controls_test.go')}
(out/'overlay.json').write_text(json.dumps({'Replace':replace}))
consumers=[('array-callback-return','core','TestArrayCallbackReturn'),('consistent-return','core','TestConsistentReturn'),('no-unreachable-loop','core','TestNoUnreachableLoop'),('react-hooks/rules-of-hooks','react','TestRulesOfHooks'),('controls','control_flow_graph','TestAdamicTargetControls')]
for label,package,test in consumers:
 env=os.environ.copy();env['ADAMIC_TARGET_CAPTURE']=str(out/label.replace('/','_'))
 target='./internal/lint/rules/'+package if label!='controls' else './internal/lint/ecmascript/control_flow_graph'
 with (out/(label.replace('/','_')+'.log')).open('w') as log:
  subprocess.run(['go','test','-overlay='+str(out/'overlay.json'),target,'-run','^'+test,'-count=1','-timeout=10m','-v'],cwd=cohere,env=env,stdout=log,stderr=subprocess.STDOUT,check=True)
 assert pathlib.Path(env['ADAMIC_TARGET_CAPTURE']+'.input').exists(),label+' had no throwTarget calls'
