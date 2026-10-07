import json,os,pathlib,subprocess,sys
repo=pathlib.Path(__file__).resolve().parents[6]
cohere=repo/'cohere'
out=pathlib.Path(sys.argv[1]);out.mkdir(parents=True,exist_ok=True)
original=(cohere/'internal/lint/ecmascript/control_flow_graph/cfg.go').read_text()
anchor='func (b *Builder[E]) enter(blk *Block[E]) {\n\tblk.Reachable = blk.hasIncoming\n\tb.cur = blk\n}'
replacement='''func (b *Builder[E]) enter(blk *Block[E]) {
 previous := -1
 if b.cur != nil { previous = int(b.cur.index) }
 before := blk.Reachable
 blk.Reachable = blk.hasIncoming
 b.cur = blk
 adamicObserveEnter(int(blk.index),blk.hasIncoming,before,previous,blk.Reachable,int(b.cur.index),b.cur==blk)
}'''
assert original.count(anchor)==1
(out/'cfg.go').write_text(original.replace(anchor,replacement))
observer='''package control_flow_graph
import("fmt";"os";"sync")
var adamicEnterMutex sync.Mutex
func adamicObserveEnter(id int,incoming,before bool,previous int,reachable bool,current int,same bool) {
 adamicEnterMutex.Lock(); defer adamicEnterMutex.Unlock()
 path:=os.Getenv("ADAMIC_ENTER_CAPTURE");if path=="" {return}
 bit:=func(v bool)int{if v{return 1};return 0}
 f,e:=os.OpenFile(path+".input",os.O_APPEND|os.O_CREATE|os.O_WRONLY,0600);if e!=nil{panic(e)}
 fmt.Fprintf(f,"%d\\t%d\\t%d\\t%d\\n",id,bit(incoming),bit(before),previous);f.Close()
 f,e=os.OpenFile(path+".want",os.O_APPEND|os.O_CREATE|os.O_WRONLY,0600);if e!=nil{panic(e)}
 fmt.Fprintf(f,"%d\\t%d\\t%d\\t%d\\n",bit(incoming),bit(reachable),current,bit(same));f.Close()
}
'''
(out/'observer.go').write_text(observer)
controls='''package control_flow_graph
import "testing"
func TestAdamicEnterControls(t *testing.T){
 for id:=0;id<5;id++{for incoming:=0;incoming<2;incoming++{for before:=0;before<2;before++{for previous:=-1;previous<5;previous++{
 b:=&Builder[int]{};node:=&Block[int]{index:int32(id),hasIncoming:incoming==1,Reachable:before==1}
 if previous>=0{if previous==id{b.cur=node}else{b.cur=&Block[int]{index:int32(previous)}}}
 b.enter(node)
 }}}}
}'''
(out/'controls_test.go').write_text(controls)
virtual=cohere/'internal/lint/ecmascript/control_flow_graph'
replace={str(virtual/'cfg.go'):str(out/'cfg.go'),str(virtual/'adamic_enter_observer.go'):str(out/'observer.go'),str(virtual/'adamic_enter_controls_test.go'):str(out/'controls_test.go')}
(out/'overlay.json').write_text(json.dumps({'Replace':replace}))
consumers=[('array-callback-return','core','TestArrayCallbackReturn'),('consistent-return','core','TestConsistentReturn'),('no-unreachable-loop','core','TestNoUnreachableLoop'),('react-hooks/rules-of-hooks','react','TestRulesOfHooks'),('controls','../ecmascript/control_flow_graph','TestAdamicEnterControls')]
for label,package,test in consumers:
 env=os.environ.copy();env['ADAMIC_ENTER_CAPTURE']=str(out/label.replace('/','_'))
 target='./internal/lint/rules/'+package if label!='controls' else './internal/lint/ecmascript/control_flow_graph'
 with (out/(label.replace('/','_')+'.log')).open('w') as log:
  subprocess.run(['go','test','-overlay='+str(out/'overlay.json'),target,'-run','^'+test,'-count=1','-timeout=10m','-v'],cwd=cohere,env=env,stdout=log,stderr=subprocess.STDOUT,check=True)
 assert pathlib.Path(env['ADAMIC_ENTER_CAPTURE']+'.input').exists(),label+' had no enter calls'
