import pathlib,json,sys
src,out,repo=map(pathlib.Path,sys.argv[1:]);out.mkdir(parents=True,exist_ok=True);m=json.loads((src/'overlay.json').read_text());r=m['Replace']
def change(suffix,edit):
 k=next(k for k in r if k.endswith('/'+suffix));p=out/suffix.replace('/','_');p.write_text(edit(pathlib.Path(r[k]).read_text()));r[k]=str(p.resolve())
def units(s):
 s=s.replace('import (','import (\n "os"',1)
 s=s.replace('fresh().latentScan(file)','if os.Getenv("LWP_SKIP_SCAN")!="1" {fresh().latentScan(file)}')
 needle='node := candidate.node\n';assert s.count(needle)==2;at=s.rindex(needle);s=s[:at]+s[at:].replace(needle,needle+' if node.Name()==nil || node.Name().Text()!="elaborateJsxComponents" {continue}\n',1)
 needle='err := l.lowerFunction(index, node, -1)';assert s.count(needle)==1;s=s.replace(needle,'l.pinPrepare()\n'+needle+'\n l.pinOthers()')
 return s
def full(s):
 needle='return l.latentStatementRaw(node)';at=s.rindex(needle);s=s[:at]+s[at:].replace(needle,'result,failure = l.latentStatementRaw(node)\n l.pinStatement(node,result,failure)\n return',1);return s
change('internal/lower/latent_units.go',units);change('internal/lower/latent_full.go',full)
def hook(s):
 needle='probe := *l';assert s.count(needle)==1
 return s.replace(needle,needle+'\n if os.Getenv("PIN_ISOLATE_PREFLIGHT")=="1" {probe=*l.latentSnapshot()}')
change('internal/lower/latent_hook.go',hook)

p=out/'pin.go';p.write_text(pathlib.Path(__file__).with_name('pin.go.txt').read_text());r[str(repo/'internal/lower/pin_probe.go')]=str(p.resolve())
s=(repo/'internal/lower/iterable_views.go').read_text();needle='yields := l.result.Functions[functions[0]].Yields';assert s.count(needle)==2;s=s.replace(needle,'if os.Getenv("PIN_LOG")!="" {l.pinFunctions(view,functions)}\n\t'+needle,1);s=s.replace('yields := l.result.Functions[functions[0]].Yields','if os.Getenv("PIN_MUTANT_OLD")=="1" && len(functions)>1 {return nil,l.notYet(view.where,"a view that generators of more than one function feed (a step names the one function whose throws it may leave) (#qk8rztp S2c-1)")}\n\t'+'yields := l.result.Functions[functions[0]].Yields',1);s=s.replace('import (','import (\n "os"',1);p=out/'iterable_views.go';p.write_text(s);r[str(repo/'internal/lower/iterable_views.go')]=str(p.resolve())
(out/'overlay.json').write_text(json.dumps(m,indent=2)+'\n')
