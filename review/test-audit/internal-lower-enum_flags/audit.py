import json,subprocess,time,os,re,difflib
from pathlib import Path
P=Path('review/test-audit/internal-lower-enum_flags'); S=Path('/tmp/u031'); plan=[]
def add(id,file,before,after,menu='change constant',probe=False):
 source=Path(file).read_text(); assert source.count(before)==1,(id,source.count(before)); pos=source.index(before); plan.append(dict(id=id,file=file,line=source[:pos].count('\n')+1,before=before,after=after,menu=menu,probe=probe))
E='internal/lower/enums.go';N='internal/lower/enum_never.go';F='internal/lower/enum_flags.go';G='internal/lower/namespaces_call_graph.go'
add('M01',E,'ir.NumberConstant{Value: value}, nil','ir.NumberConstant{Value: value + 1}, nil')
a=Path(E).read_text();start=a.index('\t\tif number, numeric := value.(ir.NumberConstant); numeric {');end=a.index('\n\t}\n\treturn fields, nil',start)
add('M02',E,'set(key, ir.StringConstant{Index: l.constant(member.Name().Text())})','set(key, ir.StringConstant{Index: l.constant("")})')
add('M03',E,'\t\tset(member.Name().Text(), value)\n','','drop statement')
add('M04',E,'\t\tif position, exists := positions[name]; exists {\n\t\t\tfields[position].Value = value\n\t\t\treturn\n\t\t}\n','','drop whole duplicate-key block')
add('M05',E,'if math.IsNaN(value) || math.IsInf(value, 0) {','if true || math.IsNaN(value) || math.IsInf(value, 0) {','flip condition')
add('M06',E,'func (l *lowering) numericEnum(symbol *ast.Symbol) bool {','func (l *lowering) numericEnum(symbol *ast.Symbol) bool {\n if true { return false }','return early')
add('M07',E,'if proven == whole {','if false && proven == whole {','flip condition')
add('M08',E,'return !l.openNumericEnumType(from) && from.Flags()&checker.TypeFlagsNumberLiteral != 0 && l.checker.IsTypeAssignableTo(from, to)','return true')
add('M09',E,'return from.Flags()&checker.TypeFlagsEnumLike != 0 && l.enumIdentity(from) == l.enumIdentity(to) && l.checker.IsTypeAssignableTo(from, to)','return true')
add('M10',N,'func (l *lowering) enumNeverIdentity(node *ast.Node, seen map[*ast.Node]bool) *ast.Symbol {','func (l *lowering) enumNeverIdentity(node *ast.Node, seen map[*ast.Node]bool) *ast.Symbol {\n if true { return nil }','return early')
add('M11',N,'declaration.Parent.Flags&ast.NodeFlagsConst != 0','declaration.Parent.Flags&ast.NodeFlagsConst == 0','flip condition')
add('M12',N,'Name: "enum_never",','Name: "enum_unreachable",')
add('M13',E,'func (l *lowering) enumDefaultUnreachable(node *ast.Node) bool {','func (l *lowering) enumDefaultUnreachable(node *ast.Node) bool {\n if true { return false }','return early')
add('M14',F,'number.Float() == value','number.Float() != value','flip condition')
add('M15',F,'bit <= 30','bit < 30','off-by-one bound')
add('M16',F,'return l.flagDomainSeen(binary.Left, target, seen) || l.flagDomainSeen(binary.Right, target, seen)','return l.flagDomainSeen(binary.Left, target, seen) && l.flagDomainSeen(binary.Right, target, seen)','flip condition')
add('M17',G,'\treturn declaration\n','\treturn nil\n','return early')
add('M18',G,'g.walks++','g.walks += 2')
add('M19',G,'if declaration := g.lowering.namespaceRuntimeEnum(node); declaration != nil {\n\t\t\t\tentry.reads[declaration] = node\n\t\t\t}','if declaration := g.lowering.namespaceRuntimeEnum(node); declaration != nil {\n\t\t\t}','drop statement')
add('M20','internal/lower/namespaces.go','case ast.KindFunctionDeclaration:\n\t\t\tif declaration.Body() != nil {\n\t\t\t\treturn declaration','case ast.KindFunctionDeclaration:\n\t\t\tif declaration.Body() != nil {\n\t\t\t\treturn nil','return early')
add('P_LOWER','internal/lower/lower.go','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {\n if true { return nil, nil }','empty answer',True)
add('P_REACH',G,'func (g *namespaceCallGraph) reach(function *ast.Node) map[*ast.Node]*ast.Node {','func (g *namespaceCallGraph) reach(function *ast.Node) map[*ast.Node]*ast.Node {\n if true { return nil }','empty answer',True)
(P/'plan.json').write_text(json.dumps(plan,indent=2)); originals={f:Path(f).read_text() for f in {m['file'] for m in plan}}
for file,body in originals.items():
 f=S/'original'/file; f.parent.mkdir(parents=True,exist_ok=True);f.write_text(body)
for m in plan:
 source=originals[m['file']]; changed=source.replace(m['before'],m['after'],1); target=S/'standalone'/m['id']/Path(m['file']).name;target.parent.mkdir(parents=True,exist_ok=True);target.write_text(changed);subprocess.run(['gofmt','-w',str(target)],check=True);changed=target.read_text();(P/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(source.splitlines(True),changed.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])));(target.parent/'overlay.json').write_text(json.dumps({'Replace':{str(Path(m['file']).resolve()):str(target.resolve())}}))
# Fixed plan above, no kills inspected. Switch uses a constant selector per test process.
for file,source in originals.items():
 changed=source
 for m in [x for x in plan if x['file']==file]:
  old=m['before'];new=m['after'];select='auditMutant == "'+m['id']+'"'
  if old.startswith('func '): replacement=old+'\n if '+select+' { '+new.split('if true {',1)[1].rsplit('}',1)[0]+' }'
  elif old.startswith('if ') and old.endswith(' {'):replacement='if '+select+' || ('+old[3:-2]+') {' if m['id']=='M05' else 'if !('+select+') && ('+old[3:-2]+') {'
  elif old=='bit <= 30':replacement='bit <= func() int { if '+select+' { return 29 }; return 30 }()'
  elif old=='number.Float() == value':replacement='func() bool { if '+select+' { return number.Float() != value }; return number.Float() == value }()'
  elif old.lstrip().startswith('return '):replacement='if '+select+' { '+new+' }; '+old
  elif old.startswith('case '):replacement=old.replace('return declaration','if '+select+' { return nil }; return declaration')
  elif old.startswith('if declaration :='):replacement=old.replace('entry.reads[declaration] = node','if !('+select+') { entry.reads[declaration] = node }')
  elif m['id']=='M11':replacement='func() bool { if '+select+' { return '+new+' }; return '+old+' }()'
  elif m['id']=='M01':replacement='ir.NumberConstant{Value: func() float64 { if '+select+' { return value + 1 }; return value }()}, nil'
  elif m['id']=='M12':replacement='Name: func() string { if '+select+' { return "enum_unreachable" }; return "enum_never" }(),'
  elif m['id']=='M18':replacement='if '+select+' { g.walks += 2 } else { g.walks++ }'
  elif new=='':replacement='if !('+select+') {\n'+old+'\n}'
  else:replacement='if '+select+' { '+new+' } else { '+old+' }'
  if old.endswith('\n'): replacement += '\n'
  assert old in changed,m['id'];changed=changed.replace(old,replacement,1)
 Path(file).write_text(changed)
Path('internal/lower/audit_mutant.go').write_text('package lower\nimport "os"\nvar auditMutant = os.Getenv("ADAMIC_MUTANT")\n')
subprocess.run(['gofmt','-w',*originals,'internal/lower/audit_mutant.go'],check=True)
(P/'switch.diff').write_text(subprocess.check_output(['git','diff','origin/main','--','internal/lower'],text=True)+'\n'+''.join(difflib.unified_diff([],Path('internal/lower/audit_mutant.go').read_text().splitlines(True),fromfile='/dev/null',tofile='b/internal/lower/audit_mutant.go')))
print('fixed plan installed',len(plan))
