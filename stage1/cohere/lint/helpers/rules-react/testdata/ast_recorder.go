//go:build lintoracle
package react
import ("encoding/json";"fmt";"os";"strings";"sync";"unicode/utf16";"unicode/utf8";"github.com/microsoft/TypeScript/tsc/shim/ast")
type adamicPair struct{Name string;Found bool}
type adamicNode struct {Kind,Text,Operator string;Pos,End int;Children []int}
type adamicArena struct {ID,Root int;Source string;Nodes []adamicNode;Parents []int;ids map[*ast.Node]int}
var adamicAstMutex sync.Mutex
var adamicArenas=map[*ast.Node]*adamicArena{}
func recordAdamicAst(name string,input *ast.Node,value any){
 adamicAstMutex.Lock();defer adamicAstMutex.Unlock()
 f,e:=os.OpenFile(os.Getenv("ADAMIC_AST_CAPTURE"),os.O_CREATE|os.O_APPEND|os.O_WRONLY,0600);if e!=nil{panic(e)};defer f.Close();encoder:=json.NewEncoder(f)
 root:=input;for root!=nil && root.Parent!=nil{root=root.Parent}
 arena:=adamicArenas[root]
 if arena==nil{
  arena=&adamicArena{ID:len(adamicArenas),Root:-1,Nodes:[]adamicNode{},Parents:[]int{},ids:map[*ast.Node]int{}}
  if root!=nil && root.Kind==ast.KindSourceFile{arena.Source=root.AsSourceFile().Text()}
  if !utf8.ValidString(arena.Source){panic("invalid UTF-8 source: "+name)}
  offsets:=map[int]int{};unit:=0;for position,point:=range arena.Source{offsets[position]=unit;if point>65535{unit+=2}else{unit++}};offsets[len(arena.Source)]=unit
  var visit func(*ast.Node)int
  visit=func(node *ast.Node)int{
   if node==nil{return -1};children:=[]int{};node.ForEachChild(func(child *ast.Node)bool{children=append(children,visit(child));return false})
   data:=adamicNode{Kind:strings.TrimPrefix(node.Kind.String(),"Kind"),Pos:offsets[node.Pos()],End:offsets[node.End()],Children:children}
   switch node.Kind{case ast.KindIdentifier,ast.KindPrivateIdentifier,ast.KindStringLiteral,ast.KindNumericLiteral,ast.KindBigIntLiteral,ast.KindNoSubstitutionTemplateLiteral:data.Text=node.Text()}
   if node.Kind==ast.KindHeritageClause{data.Operator=strings.TrimPrefix(node.AsHeritageClause().Token.String(),"Kind")}
   id:=len(arena.Nodes);arena.Nodes=append(arena.Nodes,data);arena.Parents=append(arena.Parents,-1);arena.ids[node]=id
   for _,child:=range children{arena.Parents[child]=id};return id
  }
  arena.Root=visit(root);adamicArenas[root]=arena;if e=encoder.Encode(arena);e!=nil{panic(e)}
 }
 index:=func(node *ast.Node)int{if node==nil{return -1};id,ok:=arena.ids[node];if !ok{panic("uncaptured node")};return id}
 out:="";switch v:=value.(type){case bool:out=fmt.Sprint(v);case string:out=v;case *ast.Node:out=fmt.Sprint(index(v));case []*ast.Node:parts:=[]string{};for _,node:=range v{parts=append(parts,fmt.Sprint(index(node)))};out=strings.Join(parts,",");case adamicPair:out=fmt.Sprintf("%t|%s",v.Found,v.Name);default:panic(fmt.Sprintf("result %T",value))}
 if !utf8.ValidString(out){panic("invalid UTF-8 helper result: "+name)}
 units:="";for _,u:=range utf16.Encode([]rune(out)){units+=fmt.Sprintf("%d,",u)}
 if e=encoder.Encode(struct{Name string;Arena,Input int;Want string}{name,arena.ID,index(input),name+"|"+units});e!=nil{panic(e)}
}
