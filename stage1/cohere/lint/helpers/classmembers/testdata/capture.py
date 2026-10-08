"""Observe every classmembers-helper invocation in both unchanged consuming suites."""
import collections, json, os, subprocess, tempfile
from pathlib import Path
HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[5]
COHERE = ROOT / 'cohere'
SOURCE = COHERE / 'internal/lint/ecmascript/classmembers/duplicates.go'
with tempfile.TemporaryDirectory(prefix='classmembers-capture-') as temp:
    temp = Path(temp)
    source = SOURCE.read_text()
    for name in ('MemberName', 'IsOverloadSignature', 'IsAccessorKind', 'KeyOf', 'ForEachDuplicate'):
        anchor = 'func ' + name + '('
        assert source.count(anchor) == 1
        source = source.replace(anchor, 'func adamicOriginal' + name + '(')
    source = source.replace('import (', 'import (\n "encoding/json"\n "os"\n "sync"\n "strings"\n "fmt"', 1)
    source += r'''
var adamicCaptureMutex sync.Mutex
func adamicRecord(symbol, kind string, name int, body bool, result any) {
    path := os.Getenv("ADAMIC_CLASSMEMBERS_CAPTURE")
    if path == "" { return }
    adamicCaptureMutex.Lock()
    defer adamicCaptureMutex.Unlock()
    f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
    if err != nil { panic(err) }
    defer f.Close()
    if err := json.NewEncoder(f).Encode(map[string]any{"symbol":symbol,"kind":strings.TrimPrefix(kind,"Kind"),"name":name,"body":body,"result":result}); err != nil { panic(err) }
}
func MemberName(member *ast.Node) *ast.Node {
    result := adamicOriginalMemberName(member)
    name := -1
    if member.Name() != nil { name = 0 }
    value := -1
    if result != nil {
        if result != member.Name() { panic("member identity differs") }
        value = 0
    }
    adamicRecord("MemberName", member.Kind.String(), name, false, value)
    return result
}
func IsOverloadSignature(member *ast.Node) bool {
    result := adamicOriginalIsOverloadSignature(member)
    adamicRecord("IsOverloadSignature", member.Kind.String(), -1, member.Body() != nil, result)
    return result
}
func IsAccessorKind(kind ast.Kind) bool {
    result := adamicOriginalIsAccessorKind(kind)
    adamicRecord("IsAccessorKind", kind.String(), -1, false, result)
    return result
}
'''
    source += r'''
type adamicPropertyFact struct {
 Kind string `json:"kind"`; Text string `json:"text"`; Expression int `json:"expression"`
}
type adamicMemberFact struct {
 Kind string `json:"kind"`; Name int `json:"name"`; Body bool `json:"body"`; Static bool `json:"static"`
}
func adamicProject(members []*ast.Node, name *ast.Node) ([]adamicPropertyFact, []adamicMemberFact, map[*ast.Node]int) {
 nodes:=[]adamicPropertyFact{};ids:=map[*ast.Node]int{}
 var project func(*ast.Node) int
 project=func(n *ast.Node)int{
  if n==nil{return -1};if index,ok:=ids[n];ok{return index}
  index:=len(nodes);ids[n]=index
  fact:=adamicPropertyFact{Kind:strings.TrimPrefix(n.Kind.String(),"Kind"),Expression:-1}
  switch n.Kind{case ast.KindIdentifier,ast.KindPrivateIdentifier,ast.KindStringLiteral,ast.KindNoSubstitutionTemplateLiteral,ast.KindNumericLiteral:fact.Text=n.Text()}
  nodes=append(nodes,fact)
  var child *ast.Node
  switch n.Kind{case ast.KindComputedPropertyName:child=n.AsComputedPropertyName().Expression;case ast.KindParenthesizedExpression:child=n.AsParenthesizedExpression().Expression}
  if child!=nil{edge:=project(child);nodes[index].Expression=edge}
  return index
 }
 facts:=[]adamicMemberFact{}
 for _,m:=range members{facts=append(facts,adamicMemberFact{Kind:strings.TrimPrefix(m.Kind.String(),"Kind"),Name:project(m.Name()),Body:m.Body()!=nil,Static:ast.HasStaticModifier(m)})}
 project(name)
 return nodes,facts,ids
}
func adamicRecordFull(row map[string]any){
 path:=os.Getenv("ADAMIC_CLASSMEMBERS_CAPTURE");if path==""{return}
 adamicCaptureMutex.Lock();defer adamicCaptureMutex.Unlock()
 f,err:=os.OpenFile(path,os.O_CREATE|os.O_APPEND|os.O_WRONLY,0600);if err!=nil{panic(err)};defer f.Close()
 if err:=json.NewEncoder(f).Encode(row);err!=nil{panic(err)}
}
func KeyOf(member *ast.Node,name *ast.Node)(Key,bool){
 key,known:=adamicOriginalKeyOf(member,name)
 nodes,members,ids:=adamicProject([]*ast.Node{member},name)
 identity:=-1;if name!=nil{identity=ids[name]}
 adamicRecordFull(map[string]any{"symbol":"KeyOf","kind":"","nodes":nodes,"members":members,"name":identity,"result":fmt.Sprintf("%t|%s|%t|%t",known,key.Name,key.IsStatic,key.IsPrivate)})
 return key,known
}
func ForEachDuplicate(members *ast.NodeList,report func(*ast.Node,Key)){
 raw:=[]*ast.Node{};if members!=nil{raw=members.Nodes}
 nodes,facts,ids:=adamicProject(raw,nil);reports:=[]string{}
 adamicOriginalForEachDuplicate(members,func(name *ast.Node,key Key){
  identity,ok:=ids[name];if !ok{panic("report names an unrelated node")}
  reports=append(reports,fmt.Sprintf("%d|%s|%t|%t",identity,key.Name,key.IsStatic,key.IsPrivate))
  report(name,key)
 })
 adamicRecordFull(map[string]any{"symbol":"ForEachDuplicate","kind":"","nodes":nodes,"members":facts,"result":strings.Join(reports,"\n")})
}
'''
    instrumented = temp / 'duplicates.go'
    instrumented.write_text(source)
    overlay = temp / 'overlay.json'
    overlay.write_text(json.dumps({'Replace':{str(SOURCE):str(instrumented)}}))
    rows, metadata = [], {}
    for family in ('core','typescript'):
        capture = temp / (family+'.jsonl')
        env = os.environ | {'ADAMIC_CLASSMEMBERS_CAPTURE':str(capture)}
        run = subprocess.run(['go','test','-overlay='+str(overlay),'./internal/lint/rules/'+family,'-run','^TestNoDupeClassMembers','-count=1','-json'],cwd=COHERE,env=env,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,text=True)
        (HERE / ('capture-'+family+'.jsonl')).write_text(run.stdout)
        assert run.returncode == 0, run.stdout[-6000:]
        events = [json.loads(line) for line in run.stdout.splitlines() if line.startswith('{')]
        assert not any(e['Action']=='skip' for e in events), 'upstream suite skipped'
        observations = [json.loads(line) for line in capture.read_text().splitlines()]
        metadata[family]={'passing_test_events':sum(e['Action']=='pass' and bool(e.get('Test')) for e in events),'calls':dict(collections.Counter(r['symbol'] for r in observations))}
        rows.extend(observations)
    assert set(r['symbol'] for r in rows) == {'MemberName','IsOverloadSignature','IsAccessorKind','KeyOf','ForEachDuplicate'}
    (HERE/'consumer-calls.json').write_text(json.dumps(rows,separators=(',',':'))+'\n')
    (HERE/'coverage.json').write_text(json.dumps(metadata,indent=2)+'\n')
    print(json.dumps(metadata))
