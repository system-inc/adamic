import json, subprocess, time
from pathlib import Path
repo=Path.cwd();root=repo/'review/compiler/lowering-chain-fixed/final-mutants';root.mkdir(exist_ok=True)
rows=[]
p=repo/'internal/lower/namespaces.go';s=p.read_text();needle='ir.Throw{Value: fit(ir.MakeError{Message: ir.StringConstant{Index: l.constant(message[len("TypeError: "):])}, Constructor: "TypeError"}, ir.Union)}';assert needle in s
rows.append(('namespace-old-panic',p,s.replace(needle,'ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}}'),'./internal/oracle','^TestLoweringChainNamespaceCatch$','exit codes differ'))
p=repo/'internal/native/readiness_error.go';s=p.read_text();a=s.index('func (e *emitter) readinessError');s=s[:a]+"""func (e *emitter) readinessError(message string) {
 e.line("adamic_panic(%s, %d);", cString("ReferenceError: "+message),len("ReferenceError: "+message))
}
""";s=s.replace('import "github.com/system-inc/adamic/internal/ir"','')
rows.append(('lexical-old-panic',p,s,'./internal/oracle','^TestLoweringChainTDZCatch$','exit codes differ'))
p=repo/'internal/lower/library_array.go';s=p.read_text();needle='func (l *lowering) libraryArrayMethodArguments(node, receiver *ast.Node, name string, written []*ast.Node) (ir.Expression, bool, error) {';assert needle in s
rows.append(('multiple-push-old-stop',p,s.replace(needle,needle+'\n if name=="push" && len(written)!=1 { return nil,true,l.notYet(node,"push with other than one value") }'),'./stage1/cohere/yaml','^TestLexerGaps$','push with other than one value'))
p=repo/'internal/lower/exceptions.go';s=p.read_text();needle='func (l *lowering) throwStatement(node *ast.Node) ([]ir.Statement, error) {';assert needle in s
rows.append(('structural-throw-old-stop',p,s.replace(needle,needle+'\n return nil,l.notYet(node,"throwing an Error that is not made where it is thrown")'),'./internal/oracle','^TestLoweringChainStructuralError$','throwing an Error'))
p=repo/'internal/lower/readiness.go';s=p.read_text();needle='defer readinessExceptions(program)';assert needle in s
rows.append(('readiness-old-effects',p,s.replace(needle,'// Mutant omits propagation of newly inserted readiness checks.'),'./internal/oracle','^TestModuleNamespaceReadsMatchNode$/early$','exit codes differ'))
for name,path,source,package,test,catcher in rows:
 directory=root/name;directory.mkdir(exist_ok=True);changed=directory/(path.name+'.txt');changed.write_text(source)
 overlay=directory/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(path):str(changed)}}));started=time.monotonic()
 with (directory/'test.log').open('w') as log:
  result=subprocess.run(['go','test','-p','1','-parallel','4','-timeout','90s','-overlay',str(overlay),package,'-run',test,'-count=1','-v'],stdout=log,stderr=subprocess.STDOUT,timeout=120)
 text=(directory/'test.log').read_text()
 if result.returncode==0 or catcher not in text or '[build failed]' in text:raise SystemExit(name+': not caught by expected assertion: '+text[-3000:])
 print(name+': caught %.2fs'%(time.monotonic()-started),flush=True)
