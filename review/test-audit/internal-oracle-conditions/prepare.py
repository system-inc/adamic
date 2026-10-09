import pathlib,subprocess,json,difflib
base=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip()
out=pathlib.Path('review/test-audit/internal-oracle-conditions'); out.mkdir(parents=True,exist_ok=True)
menu=[('M1','internal/lower/lower.go','\tcounters(lowering.result)','', 'drop statement'),('M2','internal/javascript/javascript.go','\t\te.line("debugger;")','', 'drop statement'),('M3','internal/javascript/javascript.go','case ir.ObjectCall:\n\t\tif expression.Checked {','case ir.ObjectCall:\n\t\tif false {','flip condition to false'),('M4','internal/javascript/javascript.go','return "Boolean(" + e.values(expression.Arguments) + ")"','return "!Boolean(" + e.values(expression.Arguments) + ")"','change emitted operator constant'),('M5','internal/lower/namespaces.go','func (l *lowering) namespaceInitialization(modules []*ast.SourceFile) error {','func (l *lowering) namespaceInitialization(modules []*ast.SourceFile) error {\n\tif len(modules) > 0 { return nil }','return early'),('M6','internal/javascript/javascript.go','if expression.Checked {\n\t\t\treturn fmt.Sprintf("(%s ? %s : adamicUnready(%s))"','if false {\n\t\t\treturn fmt.Sprintf("(%s ? %s : adamicUnready(%s))"','flip condition to false'),('M7','internal/lower/diagnostics.go','return "an " + name','return "a " + name','change article constant'),('M8','internal/lower/lower.go','\tborrow(lowering.result)','', 'drop statement')]
records=[]
for mid,file,old,new,kind in menu:
 original=subprocess.check_output(['git','show',base+':'+file],text=True); assert original.count(old)==1,(mid,original.count(old))
 changed=original.replace(old,new,1)
 diff=''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+file,tofile='b/'+file))
 (out/(mid+'.diff')).write_text(diff)
 records.append(dict(id=mid,file=file,line=original[:original.index(old)].count('\n')+1,old=old,new=new,kind=kind))
(out/'menu.json').write_text(json.dumps(records,indent=2)+'\n'); (out/'base.txt').write_text(base+'\n')
(out/'scope.json').write_text(pathlib.Path('/tmp/u059/rows.json').read_text())
