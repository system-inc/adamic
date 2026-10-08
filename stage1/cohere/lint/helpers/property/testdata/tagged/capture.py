"""Capture NameTagged's actual calls, including recursive calls, from every consumer."""
import collections, json, os, subprocess, tempfile
from pathlib import Path
HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[6]
COHERE=ROOT/'cohere'
SOURCE=COHERE/'internal/lint/ecmascript/property/name.go'
with tempfile.TemporaryDirectory(prefix='name-tagged-') as tmp:
    tmp=Path(tmp)
    source=SOURCE.read_text()
    assert source.count('func NameTagged(')==1
    source=source.replace('func NameTagged(', 'func adamicOriginalNameTagged(')
    source=source.replace('import (','import (\n "encoding/json"\n "fmt"\n "os"\n "sync"\n "strings"',1)
    source += r'''
var adamicTaggedMutex sync.Mutex
func NameTagged(node *ast.Node, accept Kinds) (string,bool) {
    value, known := adamicOriginalNameTagged(node,accept)
    path:=os.Getenv("ADAMIC_TAGGED_CAPTURE")
    if path=="" {return value,known}
    type fact struct {Kind string `json:"kind"`; Text string `json:"text"`; Expression int `json:"expression"`}
    nodes:=[]fact{}
    var project func(*ast.Node) int
    project=func(n *ast.Node) int {
        if n==nil {return -1}
        index:=len(nodes)
        f:=fact{Kind:strings.TrimPrefix(n.Kind.String(),"Kind"),Expression:-1}
        switch n.Kind {
        case ast.KindIdentifier,ast.KindPrivateIdentifier,ast.KindStringLiteral,ast.KindNoSubstitutionTemplateLiteral,ast.KindNumericLiteral: f.Text=n.Text()
        }
        nodes=append(nodes,f)
        switch n.Kind {
        case ast.KindComputedPropertyName:nodes[index].Expression=project(n.AsComputedPropertyName().Expression)
        case ast.KindParenthesizedExpression:nodes[index].Expression=project(n.AsParenthesizedExpression().Expression)
        }
        return index
    }
    key:=project(node)
    adamicTaggedMutex.Lock();defer adamicTaggedMutex.Unlock()
    f,err:=os.OpenFile(path,os.O_CREATE|os.O_WRONLY|os.O_APPEND,0600);if err!=nil{panic(err)};defer f.Close()
    if err:=json.NewEncoder(f).Encode(map[string]any{"nodes":nodes,"key":key,"accept":int(accept),"result":fmt.Sprintf("%t|%s",known,value)});err!=nil{panic(err)}
    return value,known
}
'''
    instrumented=tmp/'name.go';instrumented.write_text(source)
    controls=tmp/'controls_test.go'
    controls.write_text(r'''package property
import (
 "testing"
 "github.com/microsoft/TypeScript/tsc/shim/ast"
 "github.com/microsoft/TypeScript/tsc/shim/core"
 "github.com/microsoft/TypeScript/tsc/shim/parser"
)
func TestAdamicTaggedControls(t *testing.T) {
 for _,source:=range []string{"class C { [((1e1))]() {} ['10']() {} [variable]() {} #name() {} null() {} }", "class C { [((`x`))]() {} [null]() {} [1n]() {} [(true)]() {} }"} {
  file:=parser.ParseSourceFile(ast.SourceFileParseOptions{FileName:"/keys.ts",PathKey:"/keys.ts"},source,core.ScriptKindTS)
  var walk func(*ast.Node)
  walk=func(n *ast.Node){if name:=n.Name();name!=nil{for mask:=0;mask<64;mask++{NameTagged(name,Kinds(mask))}};n.ForEachChild(func(c *ast.Node)bool{walk(c);return false})}
  walk(file.AsNode())
 }
 for mask:=0;mask<64;mask++{NameTagged(nil,Kinds(mask))}
}
''')
    overlay=tmp/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(SOURCE):str(instrumented),str(COHERE/'internal/lint/ecmascript/property/adamic_tagged_controls_test.go'):str(controls)}}))
    rows=[];coverage={}
    for family,package,selector in [('core','rules/core','^TestNoDupeClassMembers'),('typescript','rules/typescript','^TestNoDupeClassMembers'),('react','rules/react','^TestJsxPropsNoSpreadMulti'),('property','ecmascript/property','.')]:
        capture=tmp/(family+'.jsonl')
        proc=subprocess.run(['go','test','-overlay='+str(overlay),'./internal/lint/'+package,'-run',selector,'-count=1','-json'],cwd=COHERE,env=os.environ|{'ADAMIC_TAGGED_CAPTURE':str(capture)},stdout=subprocess.PIPE,stderr=subprocess.STDOUT,text=True)
        (HERE/('capture-'+family+'.jsonl')).write_text(proc.stdout)
        assert proc.returncode==0,proc.stdout[-6000:]
        events=[json.loads(s) for s in proc.stdout.splitlines() if s.startswith('{')]
        assert not any(e['Action']=='skip' for e in events), 'upstream skip'
        observed=[json.loads(s) for s in capture.read_text().splitlines()]
        assert observed, family
        coverage[family]={'calls':len(observed),'passing_test_events':sum(e['Action']=='pass' and bool(e.get('Test')) for e in events)}
        rows.extend(observed)
    (HERE/'calls.json').write_text(json.dumps(rows,separators=(',',':'))+'\n')
    (HERE/'coverage.json').write_text(json.dumps(coverage,indent=2)+'\n')
    print(json.dumps(coverage))
