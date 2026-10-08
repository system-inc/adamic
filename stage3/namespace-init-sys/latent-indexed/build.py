"""Extend the existing scratch latent tool with production-contract attribution only."""
import argparse,hashlib,json,subprocess
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('--scratch',type=Path,required=True);args=p.parse_args()
repo=Path(__file__).resolve().parents[3];territory=repo/'stage3/census/latent';scratch=args.scratch.resolve();scratch.mkdir(parents=True,exist_ok=True)
baseline='5ad36d2cf4fce475ec072f3cc68e21c393dffc45'
subprocess.run(['git','diff','--quiet',baseline,'--','internal'],cwd=repo,check=True)
base=(territory/'make_overlay.py').read_text()
base=base.replace('territory = pathlib.Path(__file__).resolve().parent',f'territory = pathlib.Path({str(territory)!r})')
old='return nil, &CheckError{Diagnostics: diagnostics}'
new='return nil, &CheckError{Diagnostics: diagnostics, OptionSites: sites, ScheduledOptionSites: scheduled, OptionDispositions: loaded.optionDispositions, OrdinaryDiagnostics: ordinaryDiagnostics}'
assert (repo/'internal/load/load.go').read_text().count(new)==1
base=base.replace("needle = '"+old+"'","needle = '"+new+"'")
base=base.replace("original.index('\\tfiles := program.Files()')", "original.index('\\tfiles := []*ast.SourceFile{}')")
# The refusal AST collector predates the record-operation and discriminant passes.
import shutil
rewrite=scratch/'refusalrewrite'
shutil.copytree(territory/'refusalrewrite',rewrite,dirs_exist_ok=True)
rp=rewrite/'rewrite.go';rs=rp.read_text()
rs=rs.replace('ident(value, "found") || ident(value, "contractError")', 'ident(value, "found") || ident(value, "contractError") || ident(value, "err") || ident(value, "recordDeclaration") || ident(value, "recordOperation")')
rp.write_text(rs)
(rewrite/'go.mod').write_text('module github.com/system-inc/adamic/stage3/census/latent/refusalrewrite\ngo 1.24\n')
base=base.replace("check=True, cwd=territory)", "check=True, cwd=scratch / 'refusalrewrite', env={**__import__('os').environ, 'GOWORK': 'off'})")
base=base.replace("str(territory / 'refusalrewrite/cmd')", "str(scratch / 'refusalrewrite/cmd')")
script=scratch/'make_overlay.py';script.write_text(base)
subprocess.run(['python3',str(script),str(repo),str(scratch)],check=True)
def edit(name,needle,replacement):
 path=scratch/name;s=path.read_text();assert s.count(needle)==1,(name,needle);path.write_text(s.replace(needle,replacement))
edit('internal_load_load.go','loaded.latentDiagnostics = diagnostics','loaded.latentDiagnostics = diagnostics; _ = ordinaryDiagnostics')
# Pending production option diagnostics have byte positions even when no ordinary
# checker diagnostic exists. Scheduled contracts are never treated as body errors.
path=scratch/'internal_load_latent_hook.go';s=path.read_text()
s=s.replace('if diagnostic.File() == file &&', 'if !p.latentScheduledMessage(p.formatDiagnostic(diagnostic)) && diagnostic.File() == file &&')
s=s.replace('for _, diagnostic := range p.latentSites {\n\t\tfile :=', 'for _, diagnostic := range p.latentSites {\n if p.latentScheduledMessage(p.formatDiagnostic(diagnostic)) { continue }\n\t\tfile :=')
s+='''
func (p *Program) latentScheduledMessage(message string) bool {
 for _, row := range p.OptionDispositions() { if row.State == OptionCheckScheduled && row.Site.Message == message { return true } }
 return false
}
'''
cut=s.index('// Preserve raw checker spans')
s=s[:cut].replace('return result','''for _, row := range p.OptionDispositions() {
        if row.State == OptionRemainingError && row.Site.File == p.FileName(file) && row.Site.Position >= start && row.Site.Position < node.End() {
            result = append(result, row.Site.Message)
        }
    }
    return result''')+s[cut:];path.write_text(s)
edit('stage3_census_latent_tool_main.go','"diagnostic_sites": program.LatentDiagnosticSites()', '"diagnostic_sites": program.LatentDiagnosticSites(), "project_options": program.UsesProjectOptions(), "requires_indexed_presence": program.RequiresIndexedPresenceChecks(), "option_dispositions": program.OptionDispositions()')
# Attribute nested attempts without inventing a lexical environment or retrying
# a refused body. Wrappers record the existing return or rethrow its panic.
path=scratch/'internal_lower_functions.go';s=path.read_text()
s=s.replace('import (','import (\n "fmt"',1)
s=s.replace('func (l *lowering) lowerFunction(index int, declaration *ast.Node, this int) error {','func (l *lowering) latentOriginalLowerFunction(index int, declaration *ast.Node, this int) error {',1)
s=s.replace('func (l *lowering) signature(index int, declaration *ast.Node, this int) error {','func (l *lowering) latentOriginalSignature(index int, declaration *ast.Node, this int) error {',1)
s+='''
func (l *lowering) lowerFunction(index int, declaration *ast.Node, this int) (failure error) {
    defer func() {
        if value := recover(); value != nil {
            l.latentAttempt(declaration, "panic", fmt.Errorf("%v", value)); panic(value)
        }
        l.latentAttempt(declaration, "body", failure)
    }()
    return l.latentOriginalLowerFunction(index, declaration, this)
}
func (l *lowering) signature(index int, declaration *ast.Node, this int) error {
    failure := l.latentOriginalSignature(index, declaration, this)
    if failure != nil { l.latentAttempt(declaration, "signature", failure) }
    return failure
}
''';path.write_text(s)
path=scratch/'internal_lower_latent_hook.go';s=path.read_text()
s=s.replace('type LatentUnit struct {','type LatentUnit struct {\n Start int `json:"start"`\n End int `json:"end"`')
s=s.replace('type LatentFile struct {','type LatentFile struct {\n Checks []LatentCheck `json:"checks"`\n Reads []LatentRead `json:"reads"`\n FunctionAttempts []LatentFunctionAttempt `json:"function_attempts"`')
s=s.replace('case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindMethodDeclaration,','case ast.KindArrowFunction, ast.KindConstructor, ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindMethodDeclaration,')
s=s.replace('if name := node.Name(); name != nil {\n\t\t\t\ttext := file.Text()[name.Pos():name.End()]','if name := node.Name(); name != nil || ast.IsFunctionLike(node) {\n text := "<anonymous>"; if name != nil { text = file.Text()[name.Pos():name.End()] }')
s=s.replace('latentFindings = nil','latentFindings = nil\n latentChecks = nil; latentReads = nil; latentFunctionAttempts = nil',1)
s=s.replace('Kind: node.Kind.String(), Status: "attempted"','Kind: node.Kind.String(), Start:node.Pos(), End:node.End(), Status: "attempted"')
s=s.replace('l := fresh()','l := fresh()\n defer func() { for _, check := range ir.InsertedChecks(l.result) { latentChecks = append(latentChecks, LatentCheck{Unit:latentFindingOwner,Kind:check.Kind,Where:check.Where}) } }()',1)
s=s.replace('_, err := l.statement(node)','body, err := l.statement(node)\n if err == nil { l.result.Main = body }',1)
s=s.replace('emit(record)','record.Checks = latentChecks; record.Reads = latentReads; record.FunctionAttempts = latentFunctionAttempts\n emit(record)',1)
s+='''
type LatentCheck struct { Unit string `json:"unit"`; Kind string `json:"kind"`; Where string `json:"where"` }
type LatentRead struct { Unit string `json:"unit"`; Where string `json:"where"`; Start int `json:"start"`; End int `json:"end"`; Sites []load.OptionSite `json:"sites"` }
type LatentFunctionAttempt struct { Unit string `json:"unit"`; Where string `json:"where"`; Start int `json:"start"`; End int `json:"end"`; Phase string `json:"phase"`; Status string `json:"status"`; Failure *LatentFinding `json:"failure,omitempty"` }
var latentChecks []LatentCheck
var latentReads []LatentRead
var latentFunctionAttempts []LatentFunctionAttempt
func (l *lowering) latentAttempt(node *ast.Node, phase string, err error) {
    attempt := LatentFunctionAttempt{Unit:latentFindingOwner,Where:l.program.Where(node),Start:node.Pos(),End:node.End(),Phase:phase,Status:"completed"}
    if err != nil {
        attempt.Status = "blocked"
        saved := latentFindings
        latentFindings = nil; latentRecord(err)
        attempt.Failure = &latentFindings[0]; latentFindings = saved
    }
    latentFunctionAttempts = append(latentFunctionAttempts, attempt)
}
func (l *lowering) latentIndexedRead(node *ast.Node) {
    read := LatentRead{Unit:latentFindingOwner,Where:l.program.Where(node),Start:node.Pos(),End:node.End(),Sites:[]load.OptionSite{}}
    for _, site := range l.program.OptionSites() {
        if site.File != l.program.FileName(ast.GetSourceFileOfNode(node)) { continue }
        match := site.Position >= node.Pos() && site.Position < node.End()
        for parent:=node; !match && parent != nil && parent.Kind != ast.KindSourceFile; parent=parent.Parent {
            match = l.program.Where(parent) == fmt.Sprintf("%s:%d:%d",site.File,site.Line,site.Column)
        }
        if match { read.Sites = append(read.Sites,site) }
    }
    latentReads = append(latentReads,read)
}
''';path.write_text(s)
indexed=(repo/'internal/lower/indexed_checks.go').read_text();needle='message := "indexed read is absent: " + l.program.Where(node)';assert indexed.count(needle)==1
indexed=indexed.replace(needle,'l.latentIndexedRead(node)\n\t'+needle)
(scratch/'internal_lower_indexed_checks.go').write_text(indexed)
overlay=json.loads((scratch/'overlay.json').read_text());overlay['Replace'][str(repo/'internal/lower/indexed_checks.go')]=str(scratch/'internal_lower_indexed_checks.go');(scratch/'overlay.json').write_text(json.dumps(overlay,indent=2)+'\n')
subprocess.run(['gofmt','-w',*[str(x) for x in scratch.glob('*.go')]],check=True)
subprocess.run(['go','build','-buildvcs=false','-overlay',str(scratch/'overlay.json'),'-o',str(scratch/'latent-census'),'./stage3/census/latent/tool'],cwd=repo,check=True)
(scratch/'provenance.json').write_text(json.dumps({'compiler_baseline':baseline,'compiler':subprocess.check_output(['git','rev-parse','HEAD'],cwd=repo,text=True).strip(),'production_compiler_modified':False,'overlay_sha256':{str(k):hashlib.sha256(Path(v).read_bytes()).hexdigest() for k,v in overlay['Replace'].items()}},indent=2)+'\n')
print(scratch/'latent-census')
