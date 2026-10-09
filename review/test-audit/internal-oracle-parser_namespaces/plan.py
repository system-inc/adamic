import pathlib,json,difflib,re
p=pathlib.Path('review/test-audit/internal-oracle-parser_namespaces');(p/'diffs').mkdir(exist_ok=True)
files=['internal/lower/'+f for f in ['namespaces.go','namespace_receiver_scope.go','namespace_callable.go','predicates_proof.go','predicates.go','unknown.go','lower.go']]+['internal/oracle/oracle_test.go'];original={f:pathlib.Path(f).read_text() for f in files};(p/'original-files.json').write_text(json.dumps(original))
def function(file,header):
 s=original[file];start=s.index(header);brace=s.index('{',start);depth=1;end=brace+1
 while depth:
  if s[end]=='{':depth+=1
  elif s[end]=='}':depth-=1
  end+=1
 return s[start:end],s[start:brace+1]
ready,readyheader=function('internal/lower/namespaces.go','func (l *lowering) namespaceReadyValue(');lower,lowerheader=function('internal/lower/lower.go','func Lower(');agree,agreeheader=function('internal/oracle/oracle_test.go','func disagreement(')
items=[
('M01','internal/lower/namespace_receiver_scope.go','node.Kind != ast.KindArrowFunction','node.Kind == ast.KindArrowFunction','flip condition','bool'),
('M02','internal/lower/namespaces.go',ready,readyheader+'\n\treturn value\n}','return early','early'),
('M03','internal/lower/namespaces.go','body = append(body, ir.Assign{Local: l.namespaceReadyLocal(node), Value: ir.BooleanConstant{Value: true}})','body = append(body, ir.Assign{Local: l.namespaceReadyLocal(node), Value: ir.BooleanConstant{Value: false}})','change constant','statement'),
('M04','internal/lower/namespaces.go','if node.Kind == ast.KindElementAccessExpression {\n\t\treturn nil, true','if node.Kind == ast.KindPropertyAccessExpression {\n\t\treturn nil, true','change constant','guard'),
('M05','internal/lower/namespace_callable.go','if called(node) {','if !called(node) {','flip condition','guard'),
('M06','internal/lower/lower.go','\tif err := lowering.namespaceInitialization(modules); err != nil {\n\t\treturn nil, err\n\t}\n','','drop statement','drop'),
('M07','internal/lower/predicates_proof.go','\t\t\tcounts.Checked++\n','','drop statement','drop'),
('M08','internal/lower/predicates_proof.go','"unobservable", "no narrowed read in the "','"proven", "no narrowed read in the "','change constant','status'),
('M09','internal/lower/predicates_proof.go','name := "true"','name := "false"','change constant','statement'),
('M10','internal/lower/predicates_proof.go','implementation.Body().ForEachChild(writes)\n\tif changed {','implementation.Body().ForEachChild(writes)\n\tif !changed {','flip condition','guard'),
('M11','internal/lower/unknown.go','"use a discriminant, or a Map"','"use a discriminant"','change constant','string'),
('M12','internal/lower/unknown.go','objects > 1','objects > 2','off-by-one bound','bool'),
('M13','internal/lower/predicates.go','"a type predicate whose return is not proven ("','"a predicate whose return is not proven ("','change constant','string'),
('M14','internal/lower/predicates_proof.go','"overload %d of %s result: predicate %s is false"','"overload %d of %s result: predicate %s failed"','change constant','string'),
('P01','internal/lower/lower.go',lower,lowerheader+'\n\treturn nil, nil\n}','nil entry probe','early'),
('P02','internal/lower/lower.go',lower,lowerheader+'\n\treturn &ir.Program{}, nil\n}','empty IR entry probe','early'),
('W01','internal/oracle/oracle_test.go',agree,agreeheader+'\n\treturn ""\n}','weaken witnessed comparison','witness')]
manifest=[];switched=original.copy()
for id,file,old,new,menu,mode in items:
 s=original[file];assert s.count(old)==1,(id,s.count(old));line=s[:s.index(old)].count('\n')+1;manifest.append(dict(id=id,file=file,line=line,old=old,new=new,menu=menu,mode=mode));one=s.replace(old,new,1)
 if id.startswith('P'):
  for l in one.splitlines(True):
   m=re.match(r'\s*"([^\"]+)"\s*$',l)
   if m and m[1].split('/')[-1]+'.' not in one:one=one.replace(l,'',1)
 (p/'diffs'/(id+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),one.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 if mode=='bool':replacement='auditU067Bool("'+id+'", func() bool { return '+old+' }, func() bool { return '+new+' })'
 elif mode=='string':replacement='auditU067String("'+id+'", '+old+', '+new+')'
 elif mode=='status':replacement='auditU067String("'+id+'", "unobservable", "proven"), "no narrowed read in the "'
 elif mode=='drop':replacement='if !auditU067Selected("'+id+'") { '+old.strip()+' }\n'
 elif mode=='statement' and id=='M09':replacement='name := auditU067String("M09", "true", "false")'
 elif mode=='statement':replacement='if auditU067Selected("'+id+'") { '+new+' } else { '+old+' }'
 elif mode=='guard':
  if id=='M04':replacement='if auditU067Bool("M04", func() bool { return node.Kind == ast.KindElementAccessExpression }, func() bool { return node.Kind == ast.KindPropertyAccessExpression }) {\n\t\treturn nil, true'
  elif id=='M05':replacement='if auditU067Bool("M05", func() bool { return called(node) }, func() bool { return !called(node) }) {'
  else:replacement='implementation.Body().ForEachChild(writes)\n\tif auditU067Bool("M10", func() bool { return changed }, func() bool { return !changed }) {'
 elif mode=='witness':switched[file]=switched[file].replace(agreeheader,agreeheader+'\n if os.Getenv("ADAMIC_MUTANT") == "W01" { return "" }',1);continue
 else:
  header=readyheader if id=='M02' else lowerheader;ret='return value' if id=='M02' else 'return nil, nil' if id=='P01' else 'return &ir.Program{}, nil';switched[file]=switched[file].replace(header,header+'\n if auditU067Selected("'+id+'") { '+ret+' }',1);continue
 switched[file]=switched[file].replace(old,replacement,1)
(p/'manifest.json').write_text(json.dumps(manifest,indent=2));patch=''.join(''.join(difflib.unified_diff(original[f].splitlines(True),switched[f].splitlines(True),fromfile='a/'+f,tofile='b/'+f)) for f in files);(p/'switched-source.diff').write_text(patch)
(p/'plan.md').write_text('Fixed before any mutant results. CODE UNDER TEST: Adamic lowering (namespace receiver scope, readiness, member dispatch and predicate proof/records/refusals), reached through Lower. ORACLES: untouched source runs on Node; own counts snapshot and handwritten diagnostics/panic conditions. The three IR-mutant tests are witnesses of disagreement; only W01 weakening judges them. Production matrix excludes witnesses. All reached lowering functions are in reached-functions.txt. Fourteen menu mutants, P01 nil and P02 empty-IR probes, W01 allowed witness comparison edit. No reference fixture or Node changes.\n\n'+'\n'.join(x['id']+' '+x['file']+':'+str(x['line'])+' '+x['menu']+' '+repr(x['old'])+' -> '+repr(x['new']) for x in manifest)+'\n')
helper='''package lower
import "os"
func auditU067Selected(id string) bool { return os.Getenv("ADAMIC_MUTANT")==id }
func auditU067Bool(id string, old, changed func() bool) bool {if auditU067Selected(id){return changed()};return old()}
func auditU067String(id,old,changed string) string {if auditU067Selected(id){return changed};return old}
''';pathlib.Path('internal/lower/audit_u067.go').write_text(helper);(p/'switched-helper.go.txt').write_text(helper)
