from pathlib import Path
import json
w=Path('/workspace/scratch/abrq61p');root=Path('/workspace/adamic');rows=[json.loads(x) for x in (w/'base.jsonl').read_text().splitlines()];rel=lambda s:'src/'+s.split('/src/',1)[1];units={rel(u['where']):u for r in rows for u in r.get('units',[])};refs={u['where']:u for u in json.loads((w/'classification-references.json').read_text())['units']}
selected=set()
for where,u in units.items():
 if u['status']!='deferred_generic_instantiation' or u['instantiated'] or not refs[where]['has_body']:continue
 for call in refs[where]['calls']:
  top=call['top_unit'];cu=units.get(top)
  if cu and cu['status'] in ('attempted','panic') and cu['kind']!='KindModuleDeclaration' and cu['kind']!='KindClassDeclaration':selected.add('/workspace/scratch/abrq61p-adapted/'+top)
(w/'caller-probe-units.txt').write_text('\n'.join(sorted(selected))+'\n')
m=json.loads((w/'head-overlay/overlay.json').read_text())
hook=(w/'head-overlay/internal_lower_latent_hook.go').read_text();needle='for _, file := range program.Files() {';assert hook.count(needle)==1;hook=hook.replace(needle,needle+'\n if !latentProbeFile(program.FileName(file)) { continue }',1)
needle='for _, node := range file.Statements.Nodes {\n\t\t\tlatentFindingOwner';assert hook.count(needle)==1;hook=hook.replace(needle,'for _, node := range file.Statements.Nodes {\n if !latentProbeUnit(program.Where(node)) { continue }\n\t\t\tlatentFindingOwner',1)
(w/'probe/selected-hook.go.txt').write_text(hook);m['Replace'][str(root/'internal/lower/latent_hook.go')]=str(w/'probe/selected-hook.go.txt')
f=(w/'head-overlay/internal_lower_functions.go').read_text()
for method,params,args in [('latentMainSignatureRaw','index int, declaration *ast.Node, this int','index, declaration, this'),('lowerBody','index int, declaration *ast.Node, this int, defaults []defaulted, patterns []patterned','index, declaration, this, defaults, patterns')]:
 needle='func (l *lowering) '+method+'(';assert f.count(needle)==1;raw='latentProbe_'+method;f=f.replace(needle,'func (l *lowering) '+raw+'(',1);f+='\nfunc (l *lowering) '+method+'('+params+') (err error) { latentProbeStage("'+method+'_entry",l,declaration,nil); err=l.'+raw+'('+args+'); latentProbeStage("'+method+'_exit",l,declaration,err); return err }\n'
(w/'probe/trace-functions.go.txt').write_text(f);m['Replace'][str(root/'internal/lower/functions.go')]=str(w/'probe/trace-functions.go.txt')
helper='''package lower
import("encoding/json";"os";"strings";"github.com/microsoft/TypeScript/tsc/shim/ast")
var latentProbeUnits=func() map[string]bool { b,e:=os.ReadFile(os.Getenv("LATENT_PROBE_UNITS"));if e!=nil{panic(e)};m:=map[string]bool{};for _,s:=range strings.Split(strings.TrimSpace(string(b)),"\\n"){m[s]=true};return m }()
func latentProbeUnit(where string) bool{return latentProbeUnits[where]}
func latentProbeFile(file string) bool{for u:=range latentProbeUnits{if strings.HasPrefix(u,file+":"){return true}};return false}
func latentProbeStage(stage string,l *lowering,node *ast.Node,err error){p:=os.Getenv("LATENT_PROBE_STAGES");if p==""{return};f,e:=os.OpenFile(p,os.O_CREATE|os.O_APPEND|os.O_WRONLY,0600);if e!=nil{panic(e)};text:="";if err!=nil{text=err.Error()};e=json.NewEncoder(f).Encode(map[string]string{"unit":latentFindingOwner,"declaration":l.program.Where(node),"stage":stage,"error":text});ce:=f.Close();if e!=nil{panic(e)};if ce!=nil{panic(ce)}}
'''
(w/'probe/stages.go.txt').write_text(helper);m['Replace'][str(root/'internal/lower/latent_caller_probe.go')]=str(w/'probe/stages.go.txt');(w/'probe/caller-overlay.json').write_text(json.dumps(m,indent=2)+'\n');print('selected roots',len(selected))
