import json,os,pathlib,subprocess,sys
repo=pathlib.Path(__file__).resolve().parents[6];cohere=repo/'cohere'
out=pathlib.Path(sys.argv[1]);out.mkdir(parents=True,exist_ok=True)
original=(cohere/'internal/lint/ecmascript/control_flow_graph/cfg.go').read_text()
anchor='func (b *Builder[E]) throwFrame() int {'
replacement='''func (b *Builder[E]) throwFrame() (result int) {
 defer func(){
  flags:=""
  for _,frame:=range b.tryStack { if flags!="" {flags+=","}; flags+=strconv.Itoa(int(frame.position))+":"; if frame.hasFinally {flags+="1"}else{flags+="0"} }
  if flags=="" {flags="."}
  adamicObserveThrowFrame(flags,result)
 }()'''
assert original.count(anchor)==1
(out/'cfg.go').write_text(original.replace(anchor,replacement).replace('import (', 'import (\n \"strconv\"',1))
observer='''package control_flow_graph
import("fmt";"os";"sync")
var adamicThrowMutex sync.Mutex
func adamicObserveThrowFrame(flags string,result int){
 adamicThrowMutex.Lock();defer adamicThrowMutex.Unlock()
 path:=os.Getenv("ADAMIC_THROW_CAPTURE");if path==""{return}
 f,e:=os.OpenFile(path+".input",os.O_APPEND|os.O_CREATE|os.O_WRONLY,0600);if e!=nil{panic(e)}
 fmt.Fprintln(f,flags);f.Close()
 f,e=os.OpenFile(path+".want",os.O_APPEND|os.O_CREATE|os.O_WRONLY,0600);if e!=nil{panic(e)}
 fmt.Fprintln(f,result);f.Close()
}'''
(out/'observer.go').write_text(observer)
controls='''package control_flow_graph
import "testing"
func TestAdamicThrowControls(t *testing.T){
 for length:=0;length<=4;length++{count:=1;for i:=0;i<length;i++{count*=6};for code:=0;code<count;code++{
 b:=&Builder[int]{};bits:=code
 for i:=0;i<length;i++{entry:=bits%6;bits/=6;b.tryStack=append(b.tryStack,&tryFrame[int]{position:tryPosition(entry/2),hasFinally:entry%2==1})}
 b.throwFrame()
 }}
 for position:=0;position<256;position++{for finally:=0;finally<2;finally++{
 b:=&Builder[int]{tryStack:[]*tryFrame[int]{{position:tryPosition(position),hasFinally:finally==1}}};b.throwFrame()
 }}
}'''
(out/'controls_test.go').write_text(controls)
virtual=cohere/'internal/lint/ecmascript/control_flow_graph'
replace={str(virtual/'cfg.go'):str(out/'cfg.go'),str(virtual/'adamic_throw_observer.go'):str(out/'observer.go'),str(virtual/'adamic_throw_controls_test.go'):str(out/'controls_test.go')}
(out/'overlay.json').write_text(json.dumps({'Replace':replace}))
consumers=[('array-callback-return','core','TestArrayCallbackReturn'),('consistent-return','core','TestConsistentReturn'),('no-unreachable-loop','core','TestNoUnreachableLoop'),('react-hooks/rules-of-hooks','react','TestRulesOfHooks'),('controls','control_flow_graph','TestAdamicThrowControls')]
for label,package,test in consumers:
 env=os.environ.copy();env['ADAMIC_THROW_CAPTURE']=str(out/label.replace('/','_'))
 target='./internal/lint/rules/'+package if label!='controls' else './internal/lint/ecmascript/control_flow_graph'
 with (out/(label.replace('/','_')+'.log')).open('w') as log:
  subprocess.run(['go','test','-overlay='+str(out/'overlay.json'),target,'-run','^'+test,'-count=1','-timeout=10m','-v'],cwd=cohere,env=env,stdout=log,stderr=subprocess.STDOUT,check=True)
 assert pathlib.Path(env['ADAMIC_THROW_CAPTURE']+'.input').exists(),label+' had no throwFrame calls'
