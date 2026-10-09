import json, os, subprocess
from pathlib import Path
repo=Path.cwd(); root=repo/'review/compiler/lowering-chain-fixed/mutants';root.mkdir(exist_ok=True)
experiments=[]
p=repo/'internal/lower/object.go';s=p.read_text();needle='func (l *lowering) elementAccess(node *ast.Node) (ir.Expression, error) {';assert needle in s
experiments.append(('element-access-old-stop',p,s.replace(needle,needle+'\n return nil, l.notYet(node, "an ElementAccessExpression")',1),'./internal/oracle','^TestElementAccess(NodeArray|Sorted|Template)Agreement$','an ElementAccessExpression'))
p=repo/'internal/native/runtime/exceptions.c';s=p.read_text();needle='\tadamic_output_flush();';assert needle in s
experiments.append(('uncaught-old-panic',p,s.replace(needle,'\tadamic_panic("Error: boundary", sizeof "Error: boundary" - 1);',1),'./cmd/adamic','^TestWASIRequestThrows$','uncaught handler exception'))
p=repo/'internal/lower/optional_chain.go';s=p.read_text();a=s.index('\treceiver := node.AsPropertyAccessExpression().Expression',s.index('func (l *lowering) chainProperty'));b=s.index('\tswitch object.Type() {',a)
experiments.append(('regex-chain-old-stop',p,s[:a]+s[b:],'./internal/native','^TestRegexProgramsKeepCheckedFieldReads$','without a represented receiver'))
for name,path,source,package,selection,catcher in experiments:
 directory=root/name;directory.mkdir(exist_ok=True);changed=directory/(path.name+'.txt');changed.write_text(source)
 overlay=directory/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(path):str(changed)}}))
 with (directory/'test.log').open('w') as log:
  result=subprocess.run(['go','test','-p','1','-parallel','4','-timeout','90s','-overlay',str(overlay),package,'-run',selection,'-count=1','-v'],stdout=log,stderr=subprocess.STDOUT,timeout=120)
 text=(directory/'test.log').read_text()
 if result.returncode==0 or catcher not in text or '[build failed]' in text: raise SystemExit(name+': not caught by expected behavior')
 print(name+': caught by '+catcher,flush=True)
