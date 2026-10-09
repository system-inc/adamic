from pathlib import Path
import json,os,subprocess,tempfile
root=Path('/workspace/adamic');source=root/'internal/fresh/fresh.go';text=source.read_text();anchor='if !a.proof.direct[target] {';assert text.count(anchor)==1
with tempfile.TemporaryDirectory(prefix='devirt-mutant-') as d:
 p=Path(d);changed=p/'fresh.go';changed.write_text(text.replace(anchor,'if false && !a.proof.direct[target] {'));overlay=p/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(source):str(changed)}}))
 cases=[('./internal/fresh','^TestMethodKeepsArgument$','area-stack-group4-method-escape-mutant.log',['want refusal, got <nil>']),('./internal/oracle','^TestFreshWriteProbesStayRefused$/^devirt_fresh_method_keeps_argument[.]a$','area-stack-group4-method-escape-oracle-mutant.log',['accepted a program that closes a cycle','Node exit 0','LeakSanitizer'])]
 for package,pattern,name,catchers in cases:
  log=Path('/tmp')/name
  with log.open('w') as f:r=subprocess.run(['go','test','-overlay='+str(overlay),package,'-run',pattern,'-count=1','-timeout','10m'],cwd=root,env=dict(os.environ,ADAMIC_GATE_UNCACHED='1'),stdout=f,stderr=subprocess.STDOUT)
  output=log.read_text();assert r.returncode==1 and all(x in output for x in catchers) and '[build failed]' not in output and 'clang failed' not in output,output
  print('omit-non-direct-argument-escape: caught by '+', '.join(catchers),flush=True)
source=root/'internal/native/element_borrow.go';text=source.read_text();anchor='if changing[target] {';assert text.count(anchor)==2
with tempfile.TemporaryDirectory(prefix='devirt-borrow-mutant-') as d:
 p=Path(d);changed=p/'element_borrow.go';changed.write_text(text.replace(anchor,'if true || changing[target] {'));overlay=p/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(source):str(changed)}}));log=Path('/tmp/area-stack-group4-borrow-proof-mutant.log')
 with log.open('w') as f:r=subprocess.run(['go','test','-overlay='+str(overlay),'./internal/native','-run','^TestDevirtualizeBorrowDocClaim$','-count=1'],cwd=root,stdout=f,stderr=subprocess.STDOUT)
 output=log.read_text();assert r.returncode==1 and 'throughVirtual: length-only targets prevented borrowing' in output and 'throughClosure: length-only targets prevented borrowing' in output and '[build failed]' not in output,output
 print('treat-length-only-targets-as-changing: caught by both borrowing-plan claims',flush=True)
