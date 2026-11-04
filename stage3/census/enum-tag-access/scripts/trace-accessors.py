from pathlib import Path
import json
root=Path('/workspace/scratch'); src=root/'head'; overlay=root/'head-overlay'; data=json.loads((overlay/'overlay.json').read_text()); text=(src/'internal/lower/enum_tag_flow.go').read_text()
text=text.replace('import (','import (\n "encoding/json"\n "os"',1)
needle='func (l *lowering) enumTagRefinementProven(reference *ast.Node, declared *checker.Type) bool {'
assert text.count(needle)==1
text=text.replace(needle,needle+'''\n previous := censusAccessorReference
 censusAccessorReference = reference
 defer func(){censusAccessorReference = previous}()
''')
needle='return l.writtenAccessorNames[name] || l.writtenAccessorNames[""]'
assert text.count(needle)==1
text=text.replace(needle,'''hit := l.writtenAccessorNames[name] || l.writtenAccessorNames[""]
 if hit && censusAccessorReference != nil {
  if path := os.Getenv("CENSUS_ACCESSOR_TRACE"); path != "" {
   output, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
   if err != nil {panic(err)}
   err = json.NewEncoder(output).Encode(map[string]any{"where":l.program.Where(censusAccessorDiagnostic),"reference_where":l.program.Where(censusAccessorReference),"diagnostic_kind":censusAccessorDiagnostic.Kind.String(),"diagnostic_start":censusAccessorDiagnostic.Pos(),"diagnostic_end":censusAccessorDiagnostic.End(),"name":name,"unit":latentFindingOwner,"phase":latentFindingPhase,"stage":censusAccessorStage})
   output.Close()
   if err != nil {panic(err)}
  }
 }
 return hit''')
for method, stage in [('func (l *lowering) accessLocalTests(read *ast.Node) ([]*ast.Node, bool) {','access_path'), ('func (l *lowering) codeBetween(from int, read *ast.Node) bool {','code_between')]:
 assert text.count(method)==1
 text=text.replace(method, method+'\n previousStage := censusAccessorStage\n censusAccessorStage = '+json.dumps(stage)+'\n defer func(){censusAccessorStage = previousStage}()\n')
text+='\nvar censusAccessorReference *ast.Node\nvar censusAccessorDiagnostic *ast.Node\nvar censusAccessorStage string\n'
out=root/'accessor-trace.go'; out.write_text(text); data['Replace'][str(src/'internal/lower/enum_tag_flow.go')]=str(out)
(root/'accessor-overlay.json').write_text(json.dumps(data,indent=2)+'\n')
# Replay original refusal scans only for files with head open-tag findings.
reason='an object refinement using an open numeric enum as a literal tag'
records=[json.loads(s) for s in (root/'head-results/latent.jsonl').read_text().splitlines()]
targets=sorted({r['file'] for r in records[1:] if any(f['phase']!='lowering' and f['reason']==reason for f in r.get('findings',[]))})
(root/'accessor-files.json').write_text(json.dumps(targets)+'\n')
text=out.read_text()+'''
var censusAccessorTargets map[string]bool
func censusAccessorSelected(where string) bool {
 if censusAccessorTargets == nil {
  data,err:=os.ReadFile(os.Getenv("CENSUS_ACCESSOR_FILES")); if err!=nil {panic(err)}
  var names []string; if err:=json.Unmarshal(data,&names); err!=nil {panic(err)}
  censusAccessorTargets=map[string]bool{}
  for _,name:=range names {censusAccessorTargets[name]=true}
 }
 return censusAccessorTargets[where]
}
'''
out.write_text(text)
hook_path=Path(data['Replace'][str(src/'internal/lower/latent_hook.go')])
hook=hook_path.read_text()
needle='for _, file := range program.Files() {\n\t\tlatentFindings = nil'
assert hook.count(needle)==1
hook=hook.replace(needle,'for _, file := range program.Files() {\n if !censusAccessorSelected(program.FileName(file)) {continue}\n\t\tlatentFindings = nil')
needle='for _, node := range file.Statements.Nodes {\n\t\t\tlatentFindingOwner, latentFindingPhase'
assert hook.count(needle)==1
hook=hook.replace(needle,'for _, node := range file.Statements.Nodes {\n if os.Getenv("CENSUS_ACCESSOR_SCAN_ONLY")=="1" {continue}\n\t\t\tlatentFindingOwner, latentFindingPhase')
hook_out=root/'accessor-hook.go'; hook_out.write_text(hook)
data['Replace'][str(src/'internal/lower/latent_hook.go')]=str(hook_out)
(root/'accessor-overlay.json').write_text(json.dumps(data,indent=2)+'\n')

# Keep the refused diagnostic node distinct from the flow reference (a property name can use its parent access).
enums=(src/'internal/lower/enums.go').read_text()
needle='func (l *lowering) enumRefusal(node *ast.Node) error {'
assert enums.count(needle)==1
enums=enums.replace(needle,needle+'''
 previousDiagnostic := censusAccessorDiagnostic
 censusAccessorDiagnostic = node
 defer func(){censusAccessorDiagnostic = previousDiagnostic}()
''')
enums_out=root/'accessor-enums.go'; enums_out.write_text(enums)
data['Replace'][str(src/'internal/lower/enums.go')]=str(enums_out)
(root/'accessor-overlay.json').write_text(json.dumps(data,indent=2)+'\n')
(root/'accessor-results/trace.jsonl').unlink(missing_ok=True)
