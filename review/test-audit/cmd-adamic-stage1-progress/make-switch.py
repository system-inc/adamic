from pathlib import Path
import subprocess,json,difflib
p=Path('review/test-audit/cmd-adamic-stage1-progress'); (p/'diffs').mkdir(exist_ok=True); file='cmd/adamic-stage1-progress/main.go'; original=Path(file).read_text()
plan=[
('M1','change constant',"bytes.Count(source, []byte{'\\n'})","bytes.Count(source, []byte{'\\r'})"),
('M2','change constant','strings.HasSuffix(name, "_test.go")','strings.HasSuffix(name, "_audit_test.go")'),
('M3','drop statement','\t\tn++\n',''),
('M4','change condition option','if function.Doc != nil {','if false && function.Doc != nil {'),
('M5','off-by-one bound','n <= files.Position(function.End()).Line','n < files.Position(function.End()).Line'),
('M6','drop complete validation loop','\tfor name, found := range wanted {\n\t\tif !found {\n\t\t\treturn nil, fmt.Errorf("mapped declaration %s missing", name)\n\t\t}\n\t}\n',''),
('M7','change constant','100 * float64(result.Ported) / float64(result.Total)','1 * float64(result.Ported) / float64(result.Total)'),
('M8','flip condition','if report.Cohere != first.Cohere {','if report.Cohere == first.Cohere {'),
('M9','change constant','strings.HasPrefix(ref, "origin/codex/stage1-")','strings.HasPrefix(ref, "origin/codex/stage2-")'),
('M10','change constant','return strings.Join(paragraph, " ")','return strings.Join(paragraph, "|")'),
('M11','off-by-one bound','data := make([]byte, size+1)','data := make([]byte, size)'),
('M12','flip condition','if !known[directory] {','if known[directory] {'),
('P1','probe','func physicalLines(source []byte) int {','func physicalLines(source []byte) int {\n\tif true { return 0 }'),
('P2','probe','func production(name string) bool {','func production(name string) bool {\n\tif true { return false }'),
('P3','probe','func credit(source []byte, declarations []string) (map[int]bool, error) {','func credit(source []byte, declarations []string) (map[int]bool, error) {\n\tif true { return map[int]bool{}, nil }'),
('P4','probe','func union(reports []*report) (*report, error) {','func union(reports []*report) (*report, error) {\n\tif true { return &report{}, nil }'),
('P5','probe','func measure(root, ref string, cache map[string]*inventory) (*report, error) {','func measure(root, ref string, cache map[string]*inventory) (*report, error) {\n\tif true { return &report{}, nil }'),
('P6','probe','func pending(root string) ([]string, []string, error) {','func pending(root string) ([]string, []string, error) {\n\tif true { return nil, nil, nil }'),
]
entries=[]; changed=original
for id,menu,before,after in plan:
 assert original.count(before)==1,(id,original.count(before)); mutant=original.replace(before,after,1); line=original[:original.index(before)].count('\n')+1
 (p/'diffs'/(id+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),mutant.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 entries.append(dict(id=id,file=file,line=line,kind='probe' if id.startswith('P') else 'production',menu=menu,before=before,after=after))
 selector='os.Getenv("ADAMIC_MUTANT") == "'+id+'"'
 if id=='M1': replacement="bytes.Count(source, []byte{func() byte { if "+selector+" { return '\\r' }; return '\\n' }()})"
 elif id=='M2':replacement='strings.HasSuffix(name, func() string { if '+selector+' { return "_audit_test.go" }; return "_test.go" }())'
 elif id=='M3':replacement='\t\tif !('+selector+') { n++ }\n'
 elif id=='M4':replacement='if !('+selector+') && function.Doc != nil {'
 elif id=='M5':replacement='n <= files.Position(function.End()).Line - func() int { if '+selector+' { return 1 }; return 0 }()'
 elif id=='M6':replacement='\tif !('+selector+') {\n'+before+'\t}\n'
 elif id=='M7':replacement='func() float64 { if '+selector+' { return 1 }; return 100 }() * float64(result.Ported) / float64(result.Total)'
 elif id=='M8':replacement='if (report.Cohere != first.Cohere) != ('+selector+') {'
 elif id=='M9':replacement='strings.HasPrefix(ref, func() string { if '+selector+' { return "origin/codex/stage2-" }; return "origin/codex/stage1-" }())'
 elif id=='M10':replacement='return strings.Join(paragraph, func() string { if '+selector+' { return "|" }; return " " }())'
 elif id=='M11':replacement='data := make([]byte, size + func() int { if '+selector+' { return 0 }; return 1 }())'
 elif id=='M12':replacement='if (!known[directory]) != ('+selector+') {'
 else:replacement=after.replace('if true', 'if '+selector)
 assert changed.count(before)==1,(id,'switch');changed=changed.replace(before,replacement,1)
(p/'plan.json').write_text(json.dumps(entries,indent=2)+'\n')
(p/'code-and-oracles.md').write_text('Starting commit: '+subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip()+'\n\nCODE UNDER TEST: the Go progress inventory implementation, not an Adamic stage1 port. Reached functions: files, git, physicalLines, production, loadInventory, credit, blob, measure, finish, union, firstParagraph, appendUnique, pending. Coverage proof is functions.txt and reached.cover. main and fail were not reached.\n\nORACLE: self. All four tests assert hand-written fixture counts/boundaries/classifications. Git is run to construct and read real committed snapshots, not to compute a second progress inventory to compare against. No outside expected-value authority is cited.\n\nPlan fixed before any mutant outcome: twelve production mutations from the fixed menu, distributed across the reached functions, and six distinct-entry empty-answer probes. No construction, witness, helper or family rows. No skip opt-ins. Go source already imports os, so all selectors use ADAMIC_MUTANT.\n')
Path(file).write_text(changed)
subprocess.run(['gofmt','-w',file],check=True)
(p/'scratch-switch.diff').write_text(subprocess.check_output(['git','diff','--',file],text=True))
print('Fixed 12 production mutants and 6 separate probes; saved standalone diffs and switch')
