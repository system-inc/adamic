import subprocess,json
from pathlib import Path
root=Path.cwd();out=root/'review/compiler/loom-slice-4-2';rows=[]
for name,file,before,after,pkg,test,catcher in [
 ('field-pure-load','internal/fresh/fresh.go','return a.call([]value{a.value(expression.Object)}, expression.Type())','return a.load(a.value(expression.Object), expression.Name)','./internal/fresh','TestIteratorFieldAccessorEscapesReceiver','protocol accessor must leave the receiver escaped'),
 ('method-pure-load','internal/fresh/fresh.go','return a.call([]value{a.value(expression.Object)}, ir.Closure)','return a.load(a.value(expression.Object), expression.Name)','./internal/fresh','TestIteratorMethodAccessorEscapesReceiver','protocol accessor must leave the receiver escaped'),
 ('unowned-oracle-operand','internal/ir/call_targets_guard_test.go',None,'','./internal/ir','TestCallTargetReaders','unapproved call-target read internal/oracle/iteration_dispatch_mutant_test.go:TestIterationDispatchMutants:CallClosure.Closure')]:
 src=root/file;s=src.read_text()
 if before is None:before=next(l for l in s.splitlines(True) if 'iteration_dispatch_mutant_test.go:TestIterationDispatchMutants:' in l)
 assert s.count(before)==1
 mutant=out/(name+'.go.txt');mutant.write_text(s.replace(before,after));overlay=out/(name+'.json');overlay.write_text(json.dumps({'Replace':{str(src):str(mutant)}}));log=out/(name+'.log')
 with log.open('w') as f:r=subprocess.run(['go','test','-overlay',str(overlay),pkg,'-run','^'+test+'$','-count=1','-timeout','90s','-v'],stdout=f,stderr=subprocess.STDOUT,timeout=120)
 data=log.read_text();assert r.returncode and catcher in data and '[build failed]' not in data,(name,data[-3000:]);print(name,'caught',flush=True);rows.append({'mutant':name,'exit':r.returncode,'catcher':catcher})
(out/'supplement-mutants.json').write_text(json.dumps(rows,indent=2)+'\n')
