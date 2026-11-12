"""Add mapper state and actual debug.Stack top-15 frames to the retained probe."""
from pathlib import Path
import json
out=Path(__file__).resolve().parent;build=Path('/workspace/dao-mapper');build.mkdir(exist_ok=True);main=Path('/workspace/orm-main')
meta=json.loads(Path('/workspace/dao-probe/overlay.json').read_text())
s=Path('/workspace/dao-probe/census_small.go').read_text().replace('import (','import (\n "runtime/debug"\n "strings"',1)
a='if !related { fmt.Fprintln(os.Stderr,"PARAMETER",label'
b='''if !related {
            if index == 1 && (label == "overload 1 of assertNode" || label == "overload 4 of assertEachNode") {
                fmt.Fprintf(os.Stderr,"MAPPER PARAMETER label=%q mapper_non_nil=%t mapper=%p unit=%q given=%q takes=%q\\n",label,l.typeMapper != nil,l.typeMapper,latentFindingOwner,l.checker.TypeToString(given),l.checker.TypeToString(takes))
                stack := strings.Split(string(debug.Stack()),"\\n")
                limit := 1+2*15; if len(stack)<limit { limit=len(stack) }
                fmt.Fprintln(os.Stderr,"MAPPER STACK BEGIN")
                fmt.Fprintln(os.Stderr,strings.Join(stack[:limit],"\\n"))
                fmt.Fprintln(os.Stderr,"MAPPER STACK END")
            }
            fmt.Fprintln(os.Stderr,"PARAMETER",label'''
assert s.count(a)==1;s=s.replace(a,b)
p=build/'census_small.go';p.write_text(s);(out/'mapper-census_small.go.txt').write_text(s);meta['Replace'][str(main/'internal/lower/census_small.go')]=str(p)
s=(main/'internal/lower/generic.go').read_text().replace('import (','import (\n "fmt"\n "os"',1)
a='\tl.typeMapper = l.instanceMapper(declaration, concreteTypes)';assert s.count(a)==1
b=a+'''
    if declaration.Name()!=nil && (declaration.Name().Text()=="assertNode" || declaration.Name().Text()=="assertEachNode") {
        fmt.Fprintf(os.Stderr,"MAPPER SETTER original=generic.go:148 function=instantiateFunctionSignature mapper=%p unit=%q call=%q generic_declaration=%q resolved_declaration=%q\\n",l.typeMapper,latentFindingOwner,l.program.Where(call),l.program.Where(declaration),l.program.Where(resolved.Declaration()))
        for _, parameter := range declaration.TypeParameters() {
            source := l.checker.GetTypeAtLocation(parameter.Name())
            target := concreteTypes[source]
            fmt.Fprintf(os.Stderr,"MAPPER BIND parameter=%q target=%q\\n",parameter.Name().Text(),l.checker.TypeToString(target))
        }
    }
'''
s=s.replace(a,b);p=build/'generic.go';p.write_text(s);(out/'mapper-generic.go.txt').write_text(s);meta['Replace'][str(main/'internal/lower/generic.go')]=str(p)
(build/'overlay.json').write_text(json.dumps(meta,indent=2)+'\n');(out/'mapper-overlay.json').write_text(json.dumps(meta,indent=2)+'\n');print('added top 15 stack frames and matching mapper setter observations')
