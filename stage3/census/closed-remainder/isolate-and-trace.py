import json,pathlib,sys
src,out,repo=map(pathlib.Path,sys.argv[1:4]);out.mkdir(parents=True,exist_ok=True)
m=json.loads((src/'overlay.json').read_text());r=m['Replace']
k=next(k for k in r if k.endswith('/internal/lower/latent_hook.go'))
s=pathlib.Path(r[k]).read_text();s=s.replace('if !lwpSelected() { continue }','if !lwpSelected() { continue }; if !lwp2Selected(program.Where(node)) { continue }')
a='ordinary := probe.refuseNode(node)';assert s.count(a)==1
s=s.replace(a,a+'\n if os.Getenv("LWP_TRACE")!="" && len(probe.viewSources)>len(l.viewSources) {probe.lwp2ProbeTrace(l)}')
assert s.count('probe := *l')==1
s=s.replace('probe := *l','probe := *l\n if os.Getenv("LWP_ISOLATE_PROBE")=="1" {probe = *l.latentSnapshot()}')
p=out/'hook.go';p.write_text(s);r[k]=str(p.resolve())
p=out/'trace.go';p.write_text(pathlib.Path(__file__).with_name('trace.go.txt').read_text()+'''
var lwp2Targets map[string]bool
func lwp2Selected(where string) bool {
 if os.Getenv("LWP_TARGETS")=="" {return true}
 if lwp2Targets==nil {lwp2Targets=map[string]bool{};b,e:=os.ReadFile(os.Getenv("LWP_TARGETS"));if e!=nil{panic(e)};for _,s:=range strings.Split(string(b),"\\n"){lwp2Targets[strings.TrimSpace(s)]=true}}
 at:=strings.Index(where,"/src/");if at>=0{where=where[at+1:]};return lwp2Targets[where]
}
''');r[str(repo/'internal/lower/lwp2_trace.go')]=str(p.resolve())
s=(repo/'internal/lower/view_widened.go').read_text();needle='\t\tif demanded {\n\t\t\tthrough := "an unnamed type"';assert s.count(needle)==1
s=s.replace(needle,'\t\tif demanded {\n if os.Getenv("LWP_TRACE")!="" { l.lwp2Trace(read,reaches,inside,lies,unentered) }\n\t\t\tthrough := "an unnamed type"')
s=s.replace('import (','import (\n "os"',1);p=out/'view_widened.go';p.write_text(s);r[str(repo/'internal/lower/view_widened.go')]=str(p.resolve())
(out/'overlay.json').write_text(json.dumps(m,indent=2)+'\n')
