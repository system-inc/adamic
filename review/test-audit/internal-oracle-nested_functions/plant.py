from pathlib import Path
import json,difflib,subprocess,hashlib
P=Path('review/test-audit/internal-oracle-nested_functions'); (P/'diffs').mkdir(exist_ok=True)
plan=[]
def add(id,file,old,new,menu,kind='production'):
 text=subprocess.check_output(['git','show','HEAD:'+file],text=True)
 assert text.count(old)==1,(id,text.count(old))
 line=text[:text.index(old)].count('\n')+1
 plan.append(dict(id=id,file=file,line=line,old=old,new=new,menu=menu,kind=kind))
add('M01','internal/load/load.go','len(diagnostics) > 0','len(diagnostics) > 1','off-by-one')
add('M02','internal/load/load.go','\t\tall = append(all, p.compiler.GetSemanticDiagnostics(ctx, nil)...)','\t\t// Semantic diagnostic collection dropped.','drop statement')
add('M03','internal/load/load.go','diagnostic.Code(), message','diagnostic.Code()+1, message','off-by-one')
add('M04','internal/lower/cycles.go','!declared.Captured || declared.Global','declared.Captured || declared.Global','flip condition')
add('M05','internal/lower/cycles.go','if node == target {\n\t\t\t\treturn true','if node == target {\n\t\t\t\treturn false','change constant')
add('M06','internal/lower/cycles.go','&& !l.closedFrameInput(local)','&& l.closedFrameInput(local)','flip condition')
add('M07','internal/lower/typed_arrays.go','"typed array element type "+name','"typed array facility "+name','change constant')
add('M08','internal/lower/typed_arrays.go','if node.Kind == ast.KindNewExpression {','if node.Kind == ast.KindCallExpression {','change constant')
add('M09','internal/lower/new_class_value.go','if !closed {','if closed {','flip condition')
add('M10','internal/lower/new_class_value.go','closed && !initializer.Closure','closed && initializer.Closure','flip condition')
add('M11','internal/lower/new_class_value.go','len(value.Arguments) == 0 && len(targets) > 0','len(value.Arguments) <= 1 && len(targets) > 0','off-by-one')
add('M12','internal/lower/nested_functions.go','l.function.NestedFrame = true','l.function.NestedFrame = false','change constant')
add('M13','internal/lower/nested_functions.go','l.result.Locals[local].Preallocated = true\n\t\treturn []ir.Statement','l.result.Locals[local].Preallocated = false\n\t\treturn []ir.Statement','change constant')
add('M14','internal/lower/nested_functions.go','\t\tl.result.Functions[function].ForwardedNestedParent = parent','\t\t// Forwarded parent registration dropped.','drop statement')
add('M15','internal/lower/nested_functions.go','\t\tl.result.Functions[index].Environment = slices.Clone(environment)','\t\t// Sibling environment unification dropped.','drop statement')
add('M16','internal/lower/diagnostics.go','What: what','What: "unsupported construct"','change constant')
add('E00','internal/lower/lower.go','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) { return &ir.Program{}, nil;','return early','empty-answer')
add('W01','internal/oracle/oracle_test.go','return fmt.Sprintf("exit %d\\n%s", report.exitCode, report.stderr)','return ""','return early','witness')
add('W02','internal/oracle/oracle_test.go','func disagreement(oracle run, native run) string {','func disagreement(oracle run, native run) string { return "";','return early','witness')
(P/'plan.json').write_text(json.dumps(plan,indent=2)+'\n')
for m in plan:
 original=subprocess.check_output(['git','show','HEAD:'+m['file']],text=True)
 changed=original.replace(m['old'],m['new'])
 diff=''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file']))
 (P/'diffs'/(m['id']+'.diff')).write_text(diff)
print('Plan frozen',hashlib.sha256((P/'plan.json').read_bytes()).hexdigest())
# Expressions preserve lexical declarations; drop mutations guard original statement.
for file in dict.fromkeys(m['file'] for m in plan):
 original=subprocess.check_output(['git','show','HEAD:'+file],text=True); text=original
 for m in [m for m in plan if m['file']==file]:
  old,new=m['old'],m['new']; selector='auditMutant("'+m['id']+'")'
  if m['id']=='M02' or m['id']=='M14' or m['id']=='M15':
   replacement='if !'+selector+' { '+old.strip()+' }'
  elif m['id']=='M06':replacement='&& (l.closedFrameInput(local) == '+selector+')'
  elif m['id']=='M03':replacement='auditDiagnosticCode(diagnostic.Code()), message'
  elif m['id']=='M05':replacement='if node == target {\n\t\t\t\treturn !'+selector
  elif m['id']=='M07':replacement='auditElementMessage(name)'
  elif m['id']=='M12':replacement='l.function.NestedFrame = !'+selector
  elif m['id']=='M13':replacement='l.result.Locals[local].Preallocated = !'+selector+'\n\t\treturn []ir.Statement'
  elif m['id']=='M16':replacement='What: auditNotYetWhat(what)'
  elif m['id'] in ['E00','W02']:
   action='return &ir.Program{}, nil' if m['id']=='E00' else 'return ""'
   replacement=old+'\nif '+selector+' { '+action+' }'
  elif m['id']=='W01':replacement='if '+selector+' { return "" }; '+old
  else:
   # condition replacements: choose mutated condition only for its selector.
   if old.startswith('if '):
    condition=old[3:-2]; mutated=new[3:-2]
    replacement='if ((!'+selector+' && ('+condition+')) || ('+selector+' && ('+mutated+'))) {'
   else:replacement='((!'+selector+' && ('+old+')) || ('+selector+' && ('+new+')))'
  assert old in text,m['id'];text=text.replace(old,replacement)
 Path(file).write_text(text)
for package in ['load','lower','oracle']:
 helper='package '+package+'\nimport "os"\nfunc auditMutant(id string) bool { return os.Getenv("ADAMIC_MUTANT") == id }\n'
 if package=='load':helper+='func auditDiagnosticCode(n int32) int32 { if auditMutant("M03") { return n+1 }; return n }\n'
 if package=='lower':helper+='func auditElementMessage(name string) string { if auditMutant("M07") {return "typed array facility "+name};return "typed array element type "+name }\nfunc auditNotYetWhat(what string) string {if auditMutant("M16") {return "unsupported construct"};return what}\n'
 Path('internal/'+package+'/audit_mutant.go').write_text(helper)
print('switch planted')
