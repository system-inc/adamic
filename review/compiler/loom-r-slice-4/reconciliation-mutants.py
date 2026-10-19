from pathlib import Path
import json,subprocess,time
out=Path(__file__).resolve().parent;root=Path.cwd();results=[]
for name,old,new,test in [
 ('field-pure-load','return a.call([]value{a.value(expression.Object)}, expression.Type())','return a.load(a.value(expression.Object), expression.Name)','TestIteratorFieldAccessorEscapesReceiver'),
 ('method-pure-load','case ir.IteratorMethod:\n\t\t// Protocol lookup can invoke an accessor and captures its receiver.\n\t\treturn a.call([]value{a.value(expression.Object)}, ir.Closure)','case ir.IteratorMethod:\n\t\treturn a.load(a.value(expression.Object), expression.Name)','TestIteratorMethodAccessorEscapesReceiver')]:
 p=root/'internal/fresh/fresh.go';s=p.read_text();assert s.count(old)==1
 changed=out/(name+'.go.txt');changed.write_text(s.replace(old,new));overlay=out/(name+'.json');overlay.write_text(json.dumps({'Replace':{str(p):str(changed)}}))
 cmd=['go','test','-overlay='+str(overlay),'./internal/fresh','-run','^'+test+'$','-count=1','-timeout','90s','-v']
 with (out/(name+'.log')).open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,timeout=150)
 text=(out/(name+'.log')).read_text();assert r.returncode==1 and 'receiver escaped' in text and '[build failed]' not in text
 results.append(dict(mutant=name,command=cmd,exit=r.returncode,caught=True,catcher='receiver escape assertion'));print(name,'caught',flush=True)
p=root/'internal/ir/call_targets_guard_test.go';s=p.read_text();lines=[line for line in s.splitlines(True) if '"internal/oracle/iteration_dispatch_mutant_test.go:TestIterationDispatchMutants:CallClosure.Closure"' in line];assert len(lines)==1
changed=out/'reader-audit.go.txt';changed.write_text(s.replace(lines[0],''));overlay=out/'reader-audit.json';overlay.write_text(json.dumps({'Replace':{str(p):str(changed)}}))
cmd=['go','test','-overlay='+str(overlay),'./internal/ir','-run','^TestCallTargetReaders$','-count=1','-timeout','90s','-v']
with (out/'reader-audit.log').open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,timeout=150)
text=(out/'reader-audit.log').read_text();assert r.returncode==1 and 'unapproved' in text and '[build failed]' not in text
results.append(dict(mutant='unowned oracle operand',command=cmd,exit=r.returncode,caught=True,catcher='unapproved call-target read'));print('reader audit caught',flush=True)
(out/'reconciliation-mutants.json').write_text(json.dumps(results,indent=2)+'\n')
