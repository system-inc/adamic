import pathlib,subprocess,time,json,difflib,os,re,statistics
root=pathlib.Path('/workspace/adamic');os.chdir(root)
out=root/'review/test-audit/stage3-census-latent-refusalrewrite';src=root/'stage3/census/latent/refusalrewrite/rewrite.go';original=src.read_text();pkg='./stage3/census/latent/refusalrewrite/'
rows=re.findall(r'^Test\w+', (out/'list.log').read_text(),re.M)
plan=[
('M1','change constant','clone.Name.Name = "latentRefuse"','clone.Name.Name = "latentRefused"'),
('M2','drop statement','    defer func(){ latentFindingOwner = outerOwner }()','    _ = outerOwner'),
('M3','drop statement','latentRecord(%s); %s = nil; node.ForEachChild(%s); latentStop = false','latentRecord(%s); _ = %s; node.ForEachChild(%s); latentStop = false'),
('M4','change constant','latentStop = false }()','latentStop = true }()'),
('M5','off-by-one bound','len(l.program.LatentDiagnosticsIn(node.Body())) > 0','len(l.program.LatentDiagnosticsIn(node.Body())) > 1'),
('M6','drop statement','\tcollectReturns(clone.Body)','\t// collectReturns dropped'),
('M7','drop statement','\t\teraseWalks(walker.Body)','\t\t// eraseWalks dropped'),
('M8','change constant','Fun: ast.NewIdent("latentRecord"), Args: ret.Results','Fun: ast.NewIdent("latentRecord"), Args: []ast.Expr{ast.NewIdent("nil")}'),
('M9','drop statement','\t\t\t\trewriteBlocks(child, rewrite)','\t\t\t\t_ = child'),
('M10','flip condition','if latentFullEnabled() && node.Kind == ast.KindFunctionDeclaration','if latentFullEnabled() || node.Kind == ast.KindFunctionDeclaration'),
('M11','off-by-one bound','badReturn || outerReturns < 2','badReturn || outerReturns < 3'),
('M12','change constant','const function = "lowering.refuse"','const function = "lowering.refused"'),
('W1','weaken missing-function validation','return nil, fail("expected exactly one function, found %d", len(matches))','return source, nil'),
('W2','weaken child-walk validation','if count != 1 {','if count > 1 {'),
('P1','empty entry probe','func Rewrite(source []byte) ([]byte, error) {','func Rewrite(source []byte) ([]byte, error) {\n\treturn nil, nil'),
]
meta=[]
for ident,menu,old,new in plan:
 assert original.count(old)==1,(ident,original.count(old))
 line=original[:original.index(old)].count('\n')+1
 diff=''.join(difflib.unified_diff(original.splitlines(True),original.replace(old,new).splitlines(True),fromfile='a/'+str(src.relative_to(root)),tofile='b/'+str(src.relative_to(root))))
 (out/(ident+'.diff')).write_text(diff)
 meta.append(dict(id=ident,menu=menu,line=line,old=old,new=new))
(out/'plan.json').write_text(json.dumps(meta,indent=2))
(out/'inventory.json').write_text(json.dumps({'commit':subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),'nproc':subprocess.check_output(['nproc'],text=True).strip(),'rows':rows,'functions':['Rewrite','fail','ident','printed','statements','eraseWalks','collectReturns','rewriteBlocks'],'oracle_kind':'self','setup':'warm env works; skipped'},indent=2))
results={}
def run(name,cmd,env=None):
 start=time.monotonic()
 with (out/(name+'.log')).open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,env=env)
 v={'exit':r.returncode,'wall_seconds':time.monotonic()-start};results[name]=v;return v
for row in rows:
 for i in range(3):run(row+'-'+str(i+1),['timeout','120','go','test','-count=1','-timeout','90s',pkg,'-run','^'+row+'$'])
# Instrument with one environment selector, leaving instrumentation strings chosen at the Go level.
sw=original.replace('"go/token"','"go/token"\n "os"')
for ident,menu,old,new in plan:
 if ident=='P1':sw=sw.replace(old,old+'\n if os.Getenv("ADAMIC_MUTANT") == "P1" { return nil,nil }');continue
 if ident=='M12':
  sw=sw.replace('return fmt.Errorf("latent refusal rewrite: %s: %s", function, fmt.Sprintf(message, args...))','name := function; if os.Getenv("ADAMIC_MUTANT") == "M12" { name = "lowering.refused" }; return fmt.Errorf("latent refusal rewrite: %s: %s", name, fmt.Sprintf(message, args...))');continue
 if ident in ['M2','M3','M4','M5','M10']:
  continue
 if ident in ['M8','M11','W2']:
  if ident=='M8':sw=sw.replace(old,'Fun: ast.NewIdent("latentRecord"), Args: auditArgs(ret.Results)')
  elif ident=='M11':sw=sw.replace(old,'badReturn || outerReturns < auditBound()')
  else:sw=sw.replace(old,'if (os.Getenv("ADAMIC_MUTANT") != "W2" && count != 1) || (os.Getenv("ADAMIC_MUTANT") == "W2" && count > 1) {')
 else:
  sw=sw.replace(old,'if os.Getenv("ADAMIC_MUTANT") == "'+ident+'" { '+new.strip()+' } else { '+old.strip()+' }' if not new.strip().startswith('//') else 'if os.Getenv("ADAMIC_MUTANT") != "'+ident+'" { '+old.strip()+' }')
# Choose generated instrumentation before parsing it.
a='prefix := statements(`';b='binding, binding, name))'
start=sw.index(a);end=sw.index(b,start)+len(b)
block=sw[start:end]
block=block.replace('prefix := statements(`','prefixText := `',1).replace('` + fmt.Sprintf(', '` + fmt.Sprintf(',1)
block=block[:-1] # remove statements closing parenthesis
for ident,menu,old,new in plan:
 if ident in ['M2','M3','M4','M5','M10']:
  # M3 lives in fmt template, represented after interpolation instead.
  if ident=='M3':old='; '+ '%s'+' = nil;';new='; _ = '+'%s'+';';old='; '+ '" + binding + "'+' = nil;' # handled below
  if ident=='M3':
   block+='\n if os.Getenv("ADAMIC_MUTANT") == "M3" { prefixText = strings.ReplaceAll(prefixText, "; " + binding + " = nil;", "; _ = " + binding + ";") }'
  else:block+='\n if os.Getenv("ADAMIC_MUTANT") == "'+ident+'" { prefixText = strings.ReplaceAll(prefixText, '+json.dumps(old)+', '+json.dumps(new)+') }'
block+='\n prefix := statements(prefixText)'
sw=sw[:start]+block+sw[end:]
sw=sw.replace('"os"','"os"\n "strings"')
sw+='\nfunc auditArgs(args []ast.Expr) []ast.Expr { if os.Getenv("ADAMIC_MUTANT") == "M8" { return []ast.Expr{ast.NewIdent("nil")} }; return args }\nfunc auditBound() int { if os.Getenv("ADAMIC_MUTANT") == "M11" { return 3 }; return 2 }\n'
try:
 src.write_text(sw)
 r=run('switch-gofmt',['gofmt','-w',str(src)])
 if r['exit']:raise RuntimeError('gofmt failed')
 (out/'switch.diff').write_text(subprocess.check_output(['git','diff','--',str(src.relative_to(root))],text=True))
 r=run('switch-control',['timeout','120','go','test','-json','-count=1','-timeout','90s',pkg,'-run','.'])
 if r['exit']:raise RuntimeError('control failed')
 for ident,menu,old,new in plan:
  env=os.environ.copy();env['ADAMIC_MUTANT']=ident
  run(ident,['timeout','120','go','test','-json','-count=1','-timeout','90s',pkg,'-run','.'],env)
finally:
 src.write_text(original)
 (out/'runs.json').write_text(json.dumps(results,indent=2))
# Independently validate every replay diff, including probes and witness edits.
for ident,menu,old,new in plan:
 try:
  run(ident+'-apply',['git','apply','--check',str(out/(ident+'.diff'))])
  src.write_text(original.replace(old,new))
  run(ident+'-vet',['go','vet',pkg])
 finally:src.write_text(original)
(out/'runs.json').write_text(json.dumps(results,indent=2))
print(json.dumps(results,indent=2))
