#!/usr/bin/env python3
"""Scratch compatibility overlay. Does not modify shared repository files."""
import json
import subprocess
from pathlib import Path
owned = Path(__file__).resolve().parent
repository = owned.parents[5]
scratch = Path('/tmp/lint-wave1-10-next')
scratch.mkdir(exist_ok=True)
paths = ['stage1/cohere/lint/registry/registry.go', 'stage1/cohere/lint/lint_test.go', 'stage1/cohere/lint/profile_test.go']
for relative in paths:
    target = scratch / relative
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_bytes((repository / relative).read_bytes())
with (scratch / 'patch.log').open('w') as log:
    subprocess.run(['patch', '--batch', '-p1', '-d', str(scratch), '-i', str(owned / 'compatibility.patch')], stdout=log, stderr=subprocess.STDOUT, check=True)
# The existing fixture copier supplies isolated source trees. Baselines and mutants
# both receive exactly these compatibility changes, not a substituted rule answer.
p = scratch / paths[1]
s = p.read_text()
s = s.replace('side, err := filepath.Abs("testdata/oracle.go")', 'side, err := filepath.Abs("'+str(scratch / 'oracle.go')+'")')
s = s.replace('func buildPort(t *testing.T, directory string, sanitize bool) string {', 'func buildPort(t *testing.T, directory string, sanitize bool) string {\n if absolute, _ := filepath.Abs("."); directory == "." || directory == absolute { directory = mutant(t, "", "", "") }')
s = s.replace('func node(t *testing.T, directory, manifest string, count bool) execution {', 'func node(t *testing.T, directory, manifest string, count bool) execution {\n if absolute, _ := filepath.Abs("."); directory == "." || directory == absolute { directory = mutant(t, "", "", "") }')
s = s.replace('func emitted(t *testing.T, directory, path string) execution {', 'func emitted(t *testing.T, directory, path string) execution {\n if absolute, _ := filepath.Abs("."); directory == "." || directory == absolute { directory = mutant(t, "", "", "") }')
s = s.replace('source = rewritePortImports(t, file, source)', 'source = compatibilityModule(file, source)\n source = rewritePortImports(t, file, source)')
s = s.replace('Rule, Source, Outcome, FixedSource string', 'Rule, File, Source, Outcome, FixedSource string')
s = s.replace('var keys []string', 'captureData, _ := json.Marshal(unique); os.WriteFile(\"/tmp/lint-wave1-10-next/captured.json\", captureData, 0644)\n var keys []string')
s += '''
func compatibilityModule(file, source string) string {
 if file == "lint.ts" {
  source = strings.ReplaceAll(source, "finding.end > previous", "finding.editEnd > previous")
  source = strings.ReplaceAll(source, "result.slice(0, finding.start)", "result.slice(0, finding.editStart)")
  source = strings.ReplaceAll(source, "result.slice(finding.end)", "result.slice(finding.editEnd)")
  source = strings.ReplaceAll(source, "previous = finding.start", "previous = finding.editStart")
 }
 if file == "main.ts" {
  source = strings.Replace(source, "    console.log(`fixed", "    for(const finding of linter.findings) { if(finding.repair !== '') { console.log(`edit ${offsets[finding.editStart] ?? 0} ${offsets[finding.editEnd] ?? 0}`); } }\\n    console.log(`fixed", 1)
 }
 return source
}
'''
p.write_text(s)
oracle=(repository/'stage1/cohere/lint/testdata/oracle.go').read_text()
oracle=oracle.replace('file := parser.ParseSourceFile', 'scriptKind := core.ScriptKindTS; if strings.HasSuffix(path, ".tsx") { scriptKind = core.ScriptKindTSX }; if strings.HasSuffix(path, ".jsx") { scriptKind = core.ScriptKindJSX }; if strings.HasSuffix(path, ".js") { scriptKind = core.ScriptKindJS }; file := parser.ParseSourceFile').replace('source, core.ScriptKindTS)', 'source, scriptKind)')
oracle=oracle.replace('if len(file.Diagnostics()) != 0 {', 'if len(file.Diagnostics()) != 0 && os.Getenv("ADAMIC_ALLOW_PARSE_ERRORS") != "1" {')
oracle=oracle.replace('len(d.Fixes) != 1 || d.Fixes[0].Range != d.Range', 'len(d.Fixes) != 1')
oracle=oracle.replace('\tresult, err := edit.FixText', '\tif os.Getenv("ADAMIC_ALLOW_PARSE_ERRORS") == "1" && len(diagnostics) == 0 { fmt.Fprintf(out, "fixed\\t%s\\n", written(source)); return 0 }\n\tresult, err := edit.FixText')
# Print independent edit ranges in the same order after all diagnostic records.
oracle=oracle.replace('\tresult, err := edit.FixText', '\tfor _, d := range diagnostics { if len(d.Fixes) == 1 { fmt.Fprintf(out, "edit %d %d\\n", d.Fixes[0].Range.Pos(), d.Fixes[0].Range.End()) } else if len(d.Suggestions) == 1 { f := d.Suggestions[0].Fixes[0]; fmt.Fprintf(out, "edit %d %d\\n", f.Range.Pos(), f.Range.End()) } }\n\tresult, err := edit.FixText')
(scratch/'oracle.go').write_text(oracle)
replacements={str(repository/relative):str(scratch/relative) for relative in paths}
replacements[str(repository/'stage1/cohere/lint/wave10_overlay_test.go')] = str(owned/'validation_test.go.txt')
replacements[str(repository/'stage1/cohere/lint/owned_original_test.go')] = str(repository/'stage1/cohere/lint/rules/sort-vars/validation_test.go.txt')
(scratch/'overlay.json').write_text(json.dumps({'Replace':replacements}))
print('overlay='+str(scratch/'overlay.json'))
