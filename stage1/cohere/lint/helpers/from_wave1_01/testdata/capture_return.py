import json,os,pathlib,subprocess,sys
repo=pathlib.Path(__file__).resolve().parents[6];cohere=repo/'cohere'
out=pathlib.Path(sys.argv[1]);out.mkdir(parents=True,exist_ok=True)
original=(cohere/'internal/lint/ecmascript/control_flow_graph/cfg.go').read_text()
anchor='func (b *Builder[E]) returnFrame() int {'
replacement='''func (b *Builder[E]) returnFrame() (result int) {
 defer func(){
  flags:=""
  for _,frame:=range b.tryStack { if frame.hasFinally {flags+="1"}else{flags+="0"} }
  if flags=="" {flags="."}
  adamicObserveReturnFrame(flags,result)
 }()'''
assert original.count(anchor)==1
(out/'cfg.go').write_text(original.replace(anchor,replacement))
observer='''package control_flow_graph
import("fmt";"os";"sync")
var adamicReturnMutex sync.Mutex
func adamicObserveReturnFrame(flags string,result int){
 adamicReturnMutex.Lock();defer adamicReturnMutex.Unlock()
 path:=os.Getenv("ADAMIC_RETURN_CAPTURE");if path==""{return}
 f,e:=os.OpenFile(path+".input",os.O_APPEND|os.O_CREATE|os.O_WRONLY,0600);if e!=nil{panic(e)}
 fmt.Fprintln(f,flags);f.Close()
 f,e=os.OpenFile(path+".want",os.O_APPEND|os.O_CREATE|os.O_WRONLY,0600);if e!=nil{panic(e)}
 fmt.Fprintln(f,result);f.Close()
}'''
(out/'observer.go').write_text(observer)
controls='''package control_flow_graph
import "testing"
func TestAdamicReturnControls(t *testing.T){
 for length:=0;length<=8;length++{for bits:=0;bits<1<<length;bits++{for position:=0;position<2;position++{
 b:=&Builder[int]{}
 for i:=0;i<length;i++{b.tryStack=append(b.tryStack,&tryFrame[int]{hasFinally:bits&(1<<i)!=0,position:tryPosition(position)})}
 b.returnFrame()
 }}}
}'''
(out/'controls_test.go').write_text(controls)
virtual=cohere/'internal/lint/ecmascript/control_flow_graph'
replace={str(virtual/'cfg.go'):str(out/'cfg.go'),str(virtual/'adamic_return_observer.go'):str(out/'observer.go'),str(virtual/'adamic_return_controls_test.go'):str(out/'controls_test.go')}
(out/'overlay.json').write_text(json.dumps({'Replace':replace}))
consumers=[('array-callback-return','core','TestArrayCallbackReturn'),('consistent-return','core','TestConsistentReturn'),('no-unreachable-loop','core','TestNoUnreachableLoop'),('react-hooks/rules-of-hooks','react','TestRulesOfHooks'),('controls','control_flow_graph','TestAdamicReturnControls')]
for label,package,test in consumers:
 env=os.environ.copy();env['ADAMIC_RETURN_CAPTURE']=str(out/label.replace('/','_'))
 target='./internal/lint/rules/'+package if label!='controls' else './internal/lint/ecmascript/control_flow_graph'
 with (out/(label.replace('/','_')+'.log')).open('w') as log:
  subprocess.run(['go','test','-overlay='+str(out/'overlay.json'),target,'-run','^'+test,'-count=1','-timeout=10m','-v'],cwd=cohere,env=env,stdout=log,stderr=subprocess.STDOUT,check=True)
 assert pathlib.Path(env['ADAMIC_RETURN_CAPTURE']+'.input').exists(),label+' had no returnFrame calls'
