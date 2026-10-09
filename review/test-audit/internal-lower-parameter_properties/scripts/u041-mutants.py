from pathlib import Path
import json,difflib,subprocess
r=Path('review/test-audit/internal-lower-parameter_properties');menu=[]
def base(file):return subprocess.check_output(['git','show','origin/main:'+file],text=True)
def add(mid,file,old,new,change):
 b=base(file);assert b.count(old)==1,(mid,b.count(old));menu.append(dict(id=mid,file=file,line=b[:b.index(old)].count('\n')+1,old=old,new=new,change=change))
def early(mid,file,sig,value):add(mid,file,sig,sig+'\n\tif true { return '+value+' }','return early: '+value)
p='internal/lower/parameter_properties.go'
add('M01',p,'return node.Kind == ast.KindParameter &&','return node.Kind != ast.KindParameter &&','flip parameter-kind condition')
add('M02',p,'\t\t\t\t\tmembers = append(members, parameter)\n','','drop parameter-member append')
early('M03',p,'func (l *lowering) parameterPropertyStores(declaration *ast.Node, this int) ([]ir.Statement, error) {','nil, nil')
early('M04',p,'func (l *lowering) parameterPropertyDefault(declaration, initializer *ast.Node) error {','nil')
add('M05',p,'\t\tdelete(available, parameter.Name().Text())\n','','drop inherited availability reset')
early('M06','internal/lower/class_inheritance.go','func (l *lowering) checkMemberOverrides(declaration *ast.Node, classType, base *checker.Type, static bool) error {','nil')
p='internal/lower/expression.go';b=base(p);start=b.index('\tfor _, parameter := range symbol.Declarations[0].Parameters() {');end=b.index('\tcallee := l.result.Functions[target]',start)
add('M07',p,b[start:end],'','drop named-function this-parameter guard loop')
p='internal/lower/predicates_proof.go'
add('M08',p,'claim.ParameterIndex() != 0','claim.ParameterIndex() != 1','off-by-one helper predicate parameter index')
b=base(p);start=b.index('\t\t\t\t\tclaim := predicateOfSignature');end=b.index('\t\t\t\t\treturn values[path.cell]',start)
add('M09',p,b[start:end],'','drop claim declaration and parameter-index guard block')
add('M10',p,'\t\twhere = l.program.Where(node)\n','','drop predicate diagnostic location assignment')
p='internal/load/load.go'
add('M11',p,'\t\tall = append(all, p.compiler.GetSemanticDiagnostics(ctx, nil)...)\n','','drop semantic diagnostic collection')
add('M12',p,'NoUncheckedIndexedAccess:   core.TSTrue','NoUncheckedIndexedAccess:   core.TSFalse','disable unchecked-index option')
early('P1','internal/lower/lower.go','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','nil, nil')
early('P2','internal/load/load.go','func Load(paths []string) (*Program, error) {','nil, nil')
early('P3','internal/lower/predicates_proof.go','func (l *lowering) provePredicate(node *ast.Node) (predicateProof, error) {','predicateProof{}, nil')
(r/'menu.json').write_text(json.dumps(menu,indent=2)+'\n');files={m['file'] for m in menu};bases={f:base(f) for f in files}
for m in menu:
 b=bases[m['file']];changed=b.replace(m['old'],m['new']);(r/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(b.splitlines(True),changed.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
for file,b in bases.items():
 s=b
 for m in reversed([m for m in menu if m['file']==file]):
  tag='os.Getenv("ADAMIC_MUTANT") == "'+m['id']+'"';old,new=m['old'],m['new']
  if 'if true { return' in new:replacement=new.replace('if true','if '+tag)
  elif m['id']=='M01':replacement='return (('+tag+' && node.Kind != ast.KindParameter) || (!('+tag+') && node.Kind == ast.KindParameter)) &&'
  elif m['id']=='M08':replacement='(('+tag+' && claim.ParameterIndex() != 1) || (!('+tag+') && claim.ParameterIndex() != 0))'
  elif m['id']=='M12':replacement='NoUncheckedIndexedAccess: auditUncheckedIndexOption()'
  else:replacement='\tif !('+tag+') {\n'+old+'\n\t}\n'
  assert old in s,m['id'];s=s.replace(old,replacement)
 if file=='internal/load/load.go':s+='\nfunc auditUncheckedIndexOption() core.Tristate { if os.Getenv("ADAMIC_MUTANT") == "M12" { return core.TSFalse }; return core.TSTrue }\n'
 if '"os"' not in s:s=s.replace('import (','import (\n\t"os"',1)
 Path(file).write_text(s)
for name in ['npm','list','baseline']:(r/(name+'.log')).write_bytes(Path('/tmp/u041-'+name+'.log').read_bytes())
