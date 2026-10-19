from pathlib import Path
import json,subprocess,time
root=Path.cwd();out=root/'review/compiler/loom-r-views-fx-b-ac-m'
source=root/'internal/lower/interface_cast.go';text=source.read_text();before='if contract := l.result.ViewContracts[id-1]; contract.Unsupported != "" {';assert text.count(before)==1
mutant=out/'unsupported-registration.go.txt';mutant.write_text(text.replace(before,'if contract := l.result.ViewContracts[id-1]; false && contract.Unsupported != "" {',1));overlay=out/'unsupported-registration-overlay.json';overlay.write_text(json.dumps({'Replace':{str(source):str(mutant)}}))
cmd=['go','build','-overlay',str(overlay),'-o','/tmp/views-b-unsupported-mutant','./cmd/adamic'];start=time.monotonic()
with (out/'unsupported-registration-build.log').open('w') as log:p=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,timeout=120)
assert p.returncode==0
ns={};script=(out/'a-check.py').read_text();exec(script[:script.index("h=Harness('final')")],ns)
Harness=ns['Harness'];Process=ns['Process']
class MutantHarness(Harness):
 def spawn(self,cmd,stdout,stderr):return Process(self,['/tmp/views-b-unsupported-mutant',*cmd[1:]])
h=MutantHarness('unsupported-registration');paths=['review/compiler/loom-views-fx-b/native_p21.a','review/compiler/loom-views-fx-b/native_p37.a'];code=h.run(paths);assert code==1,h.failure
(out/'unsupported-registration-mutant.json').write_text(json.dumps({'build_command':cmd,'build_exit':p.returncode,'a_check_exit':code,'caught':True,'paths':paths,'failure':h.failure,'seconds':time.monotonic()-start},indent=2)+'\n');print('unsupported-registration mutant caught by both unchanged a-check predicates')
