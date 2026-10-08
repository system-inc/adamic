"""Reuse the latent loader boundary without changing production source."""
import json,pathlib,sys
repo=pathlib.Path.cwd(); scratch=pathlib.Path(sys.argv[1]);scratch.mkdir(parents=True,exist_ok=True)
root=repo/'stage3/shape-conformance/latent';replace={}
def overlay(name,text):
 output=scratch/name.replace('/','_');output.write_text(text);replace[str(repo/name)]=str(output)
loader=(repo/'internal/load/load.go').read_text()
loader=loader.replace('type Program struct {','type Program struct {\n latentSites []*ast.Diagnostic\n latentDiagnostics []string')
assert loader.count('return nil, &CheckError{Diagnostics: diagnostics}')==1
loader=loader.replace('return nil, &CheckError{Diagnostics: diagnostics}','loaded.latentDiagnostics = diagnostics')
loader=loader.replace('formatted := make([]string, 0, len(all))','p.latentSites = all\n formatted := make([]string, 0, len(all))')
loader=loader.replace('func Load(paths []string)','func LatentLoad(paths []string)').replace('func LoadOverlay(paths []string, overlay map[string]string)','func latentLoadOverlay(paths []string, overlay map[string]string)')
loader+='\nfunc Load(paths []string) (*Program,error) { return nil, fmt.Errorf("latent shape measurement: no production load") }\nfunc LoadOverlay(paths []string, overlay map[string]string) (*Program,error) { return Load(paths) }\n'
overlay('internal/load/load.go',loader)
overlay('internal/load/latent_hook.go',(repo/'stage3/census/latent/load.go.txt').read_text())
source=(repo/'internal/lower/lower.go').read_text();start=source.index('\tfiles := program.Files()');end=source.index('\n}\n\ntype lowering',start)
source=source[:start]+'\treturn nil, fmt.Errorf("latent shape measurement: no output IR")'+source[end:];source=source.replace('\n\t"path/filepath"','')
overlay('internal/lower/lower.go',source)
overlay('internal/lower/latent_shape_hook.go',(pathlib.Path(sys.argv[2]) if len(sys.argv)>2 else root/'lower.go.txt').read_text())
overlay('stage3/shape-conformance/latent/tool/main.go',(root/'main.go.txt').read_text())
(scratch/'overlay.json').write_text(json.dumps({'Replace':replace},indent=2)+'\n')
print(scratch/'overlay.json')
