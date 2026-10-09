from pathlib import Path
import json
r=Path('/workspace/adamic');o=r/'review/compiler/records-next/ruling-a';replacements={}
def overlay(name,text):
 p=o/(name.replace('/','_')+'.txt');p.write_text(text);replacements[str(r/name)]=str(p)
# Same measurement-only loader approach as stage3/census/latent/make_overlay.py.
s=(r/'internal/load/load.go').read_text();assert s.count('return nil, &CheckError{Diagnostics: diagnostics}')==1
s=s.replace('return nil, &CheckError{Diagnostics: diagnostics}','_ = diagnostics // measurement only: retain checker-rejected AST')
s=s.replace('func Load(paths []string)','func LatentLoad(paths []string)').replace('func LoadOverlay(paths []string, overlay map[string]string)','func latentLoadOverlay(paths []string, overlay map[string]string)')
s+='\nfunc Load(paths []string) (*Program,error) { return nil, fmt.Errorf("record census: measurement only") }\nfunc LoadOverlay(paths []string,overlay map[string]string) (*Program,error) { return Load(paths) }\n'
overlay('internal/load/load.go',s)
s=(r/'internal/lower/lower.go').read_text();a=s.index('\tfiles := program.Files()');b=s.index('\n}\n\ntype lowering',a);s=s[:a]+'\treturn nil, fmt.Errorf("record census: no IR output")'+s[b:];s=s.replace('\n\t"path/filepath"','');overlay('internal/lower/lower.go',s)
overlay('internal/lower/record_ruling_census.go','''package lower
import("context";"github.com/system-inc/adamic/internal/load";"github.com/microsoft/TypeScript/tsc/shim/ast";"github.com/system-inc/adamic/internal/ir";"github.com/microsoft/TypeScript/tsc/shim/checker")
func RecordRulingCensus(p *load.Program) map[string]any {
 c,release:=p.Checker(context.Background(),p.Files()[0]);defer release();l:=&lowering{program:p,checker:c,result:&ir.Program{},this:-1,functionIndex:-1};rows:=[]map[string]string{};targets:=0
 var visit ast.Visitor
 visit=func(n *ast.Node)bool {
  if l.recordTarget(n) {targets++}
  if err:=l.recordDivergenceRefusal(n);err!=nil {r:=err.(*Refused);shape:="prototype_read";if n.Kind==ast.KindBinaryExpression {shape="prototype_in";if n.AsBinaryExpression().OperatorToken.Kind==ast.KindEqualsToken {shape="prototype_set"}} else if l.recordScalarNarrowed(n) && l.recordInterveningMutation(n) {shape="narrowed_number";e:=l.recordElement(l.checker.GetTypeAtLocation(recordReceiver(n)));if e!=nil && e.Flags()&checker.TypeFlagsUnion!=0 {shape="named_invalidated"}} else if n.Parent!=nil && n.Parent.Kind==ast.KindCallExpression {for i,a:=range n.Parent.AsCallExpression().Arguments.Nodes {if a==n && len(n.Parent.AsCallExpression().Arguments.Nodes)>1 {if i==0 {shape="compare_properties_left"} else {shape="compare_properties_right"}}}} else if n.Parent!=nil && n.Parent.Kind==ast.KindBinaryExpression && n.Parent.AsBinaryExpression().OperatorToken.Kind==ast.KindQuestionQuestionToken {shape="environment_boundary"};rows=append(rows,map[string]string{"shape":shape,"where":r.Where,"reason":r.What,"fix":r.Fix})}
  return n.ForEachChild(visit)
 }
 for _,f:=range p.Files(){f.AsNode().ForEachChild(visit)};return map[string]any{"hits":rows,"record_accesses":targets}
}
''')
overlay('stage3/census/tool/main.go','''package main
import("encoding/json";"os";"path/filepath";"github.com/system-inc/adamic/internal/load";"github.com/system-inc/adamic/internal/lower")
func main(){paths:=[]string{};err:=filepath.WalkDir(os.Args[1],func(p string,e os.DirEntry,err error)error{if err!=nil{return err};if !e.IsDir() && filepath.Ext(p)==".ts" {paths=append(paths,p)};return nil});if err!=nil{panic(err)};p,err:=load.LatentLoad(paths);if err!=nil{panic(err)};result:=lower.RecordRulingCensus(p);result["measurement"]="checker-rejected AST only; no IR or backend output";result["roots"]=len(paths);if err:=json.NewEncoder(os.Stdout).Encode(result);err!=nil{panic(err)}}
''')
(o/'census-overlay.json').write_text(json.dumps({'Replace':replacements},indent=2))
