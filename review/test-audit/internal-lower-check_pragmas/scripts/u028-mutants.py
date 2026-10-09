from pathlib import Path
import json,difflib,subprocess,re
r=Path('review/test-audit/internal-lower-check_pragmas')
menu=[]
def add(mid,file,old,new,description):
 b=subprocess.check_output(['git','show','origin/main:'+file],text=True)
 assert b.count(old)==1,(mid,b.count(old))
 line=b[:b.index(old)].count('\n')+1
 menu.append(dict(id=mid,file=file,line=line,old=old,new=new,change=description))
def early(mid,file,signature,result): add(mid,file,signature,signature+'\n\tif true { return '+result+' }', 'return early: '+result)
add('M01','internal/lower/refusals.go','pragma.Name == "ts-nocheck" || pragma.Name == "ts-check"','pragma.Name == "ts-nocheck" || pragma.Name != "ts-check"','flip ts-check condition')
add('M02','internal/lower/refusals.go','line+1, column+1), What: "@" + pragma.Name','line+2, column+1), What: "@" + pragma.Name','off-by-one pragma line')
b=subprocess.check_output(['git','show','origin/main:internal/lower/refusals.go'],text=True); block=b[b.index('\tfor _, pragma := range module.Pragmas {'):b.index('\t// Validate arguments')]
add('M03','internal/lower/refusals.go',block,'','drop whole pragma loop')
add('M04','internal/lower/class_inheritance.go','Private: member.Name().Kind == ast.KindPrivateIdentifier','Private: false','change private metadata constant')
add('M05','internal/lower/class_static.go','Constructor: -1, Static: true','Constructor: -1, Static: false','change static metadata constant')
early('M06','internal/lower/class_inheritance.go','func (l *lowering) checkMemberOverrides(declaration *ast.Node, classType, base *checker.Type, static bool) error {','nil')
early('M07','internal/lower/class_accessors.go','func (l *lowering) checkAccessorSpreads() error {','nil')
early('M08','internal/lower/class_static.go','func (l *lowering) staticInitializationReads(node, declaration *ast.Node, available map[string]bool, seen map[*ast.Node]uint8, receiver bool) error {','nil')
add('M09','internal/lower/object.go','if !l.classAssignable(declared, observed) {','if l.classAssignable(declared, observed) {','flip narrowed accessor condition')
early('M10','internal/lower/cycles.go','func (l *lowering) findCycles(modules []*ast.SourceFile) error {','nil')
b=subprocess.check_output(['git','show','origin/main:internal/lower/cycles.go'],text=True)
start=b.index('\t\t\tfor _, accessor := range f.l.accessorCaptures {'); end=b.index('\t\t\tfor _, field := range f.fields(proven)',start)
add('M11','internal/lower/cycles.go',b[start:end],'','drop accessor capture traversal loop')
start=b.index('\t\t\tif f.l.isStaticType(proven) {'); end=b.index('\t\t\tfor _, accessor',start)
add('M12','internal/lower/cycles.go',b[start:end],'','drop static parent traversal block')
start=b.index('\t\t\tif len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) > 0 {'); end=b.index('\t\t\t// What a function value holds',start)
add('M13','internal/lower/cycles.go',b[start:end],'','drop construct-signature traversal block')
early('M14','internal/lower/class_static.go','func (l *lowering) needsStatics(declaration *ast.Node) bool {','false')
add('M15','internal/lower/class_static.go','\t\tstatements = append(statements, ir.Evaluate{Value: ir.Call{Function: index, Arguments: []ir.Expression{object}}})\n','','drop static initializer call statement')
add('M16','internal/load/load.go','\t\tall = append(all, p.compiler.GetSemanticDiagnostics(ctx, nil)...)\n','','drop semantic diagnostic collection')
add('M17','internal/load/load.go','"error TS%d: %s"','"diagnostic TS%d: %s"','change diagnostic label constant')
b=subprocess.check_output(['git','show','origin/main:internal/lower/lower.go'],text=True)
start=b.index('\tfor _, module := range modules {\n\t\tbody, err := lowering.statements'); end=b.index('\tif lowering.unlowerable',start)
add('M18','internal/lower/lower.go',b[start:end],'','drop whole module-body lowering loop to avoid unused body local')
early('P1','internal/lower/lower.go','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','nil, nil')
early('P2','internal/load/load.go','func Load(paths []string) (*Program, error) {','nil, nil')
(r/'menu.json').write_text(json.dumps(menu,indent=2)+'\n')
files={x['file'] for x in menu}; bases={f:subprocess.check_output(['git','show','origin/main:'+f],text=True) for f in files}
for x in menu:
 b=bases[x['file']]; altered=b.replace(x['old'],x['new'])
 (r/(x['id']+'.diff')).write_text(''.join(difflib.unified_diff(b.splitlines(True),altered.splitlines(True),fromfile='a/'+x['file'],tofile='b/'+x['file'])))
for file,b in bases.items():
 scratch=b
 for x in reversed([x for x in menu if x['file']==file]):
  old,new=x['old'],x['new']; tag='os.Getenv("ADAMIC_MUTANT") == "'+x['id']+'"'
  if 'if true { return' in new: replacement=new.replace('if true', 'if '+tag)
  elif x['id']=='M01': replacement='pragma.Name == "ts-nocheck" || ('+tag+' && pragma.Name != "ts-check") || (!('+tag+') && pragma.Name == "ts-check")'
  elif x['id']=='M02': replacement='line+1+auditPragmaOffset(), column+1), What: "@" + pragma.Name'
  elif x['id']=='M04': replacement='Private: !('+tag+') && member.Name().Kind == ast.KindPrivateIdentifier'
  elif x['id']=='M05': replacement='Constructor: -1, Static: !('+tag+')'
  elif x['id']=='M09': replacement='if (!('+tag+') && !l.classAssignable(declared, observed)) || ('+tag+' && l.classAssignable(declared, observed)) {'
  elif x['id']=='M17': replacement='auditDiagnosticLabel()'
  else: replacement='\tif !('+tag+') {\n'+old+'\n\t}\n'
  assert old in scratch,(x['id'],file)
  scratch=scratch.replace(old,replacement)
 if file.endswith('refusals.go'): scratch+='\nfunc auditPragmaOffset() int { if os.Getenv("ADAMIC_MUTANT") == "M02" { return 1 }; return 0 }\n'
 if file=='internal/load/load.go': scratch+='\nfunc auditDiagnosticLabel() string { if os.Getenv("ADAMIC_MUTANT") == "M17" { return "diagnostic TS%d: %s" }; return "error TS%d: %s" }\n'
 if '"os"' not in scratch: scratch=scratch.replace('import (','import (\n\t"os"',1)
 Path(file).write_text(scratch)
with (r/'switch.diff').open('w') as f: subprocess.run(['git','diff','--',*sorted(files)],stdout=f,check=True)
for name in ['npm','list','list-retry','baseline']:
 src=Path('/tmp/u028-'+name+'.log')
 if src.exists(): (r/(name+'.log')).write_bytes(src.read_bytes())
