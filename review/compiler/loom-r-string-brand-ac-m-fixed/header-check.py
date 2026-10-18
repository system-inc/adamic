import collections,difflib,hashlib,importlib.util,json,pathlib,re,subprocess,time
root=pathlib.Path.cwd();out=root/'review/compiler/loom-r-string-brand-ac-m-fixed'
paths=subprocess.check_output(['git','diff','--name-only','539788fef5~1','539788fef5','--','*.a'],text=True).splitlines()
compiler='/tmp/string-brand-fixed-adamic'
spec=importlib.util.spec_from_file_location('pinned_gate',root/'review/compiler/loom-r-string-brand-ac/pinned-gate.py');module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
class Process:
 def __init__(self,cmd,label):self.cmd=cmd;self.label=label
 def communicate(self):
  r=subprocess.run([compiler,*self.cmd[1:]],capture_output=True,text=True,timeout=20)
  self.returncode=r.returncode
  key=self.cmd[-1];self.label.observations[key]={'exit':r.returncode,'stderr':r.stderr,'stdout_bytes':len(r.stdout.encode())}
  return r.stdout,r.stderr
class Harness:
 def __init__(self):self.arguments=type('Args',(),{'tree':str(root)})();self.result={};self.steps={};self.exits={};self.failure=None;self.observations={}
 def step(self,name,cmd):
  # The branch compiler was already built by the recorded bounded go build.
  assert name=='a-check-build' and cmd[-1]=='./cmd/adamic';return True
 def spawn(self,cmd,stdout,stderr):return Process(cmd,self)
 def fail(self,name,detail):self.failure=detail

def scan(label,files):
 h=Harness();module.Gate.aCheck(h,files)
 record={'exit':h.exits['a-check'],'results':h.result['a_check'],'observations':h.observations,'failure':h.failure}
 (out/(label+'.json')).write_text(json.dumps(record,indent=2)+'\n')
 print(label,record['exit'],flush=True);return record


baseline=scan('a-check',paths);assert baseline['exit']==0,baseline['failure']
selected=next(name for name in paths if (root/name).read_bytes().startswith(b'// a-check: refused '));p=root/selected;original=p.read_bytes();first,body=original.split(b'\n',1);mutants=[]
for mode in ['removed-header','wrong-reason']:
 try:
  p.write_bytes(body if mode=='removed-header' else b'// a-check: refused deliberately unmatched brand reason\n'+body)
  result=scan(mode,[selected]);assert result['exit']==1 and result['results'][selected]['outcome']=='refused'
  mutants.append(dict(file=selected,mutation=mode,exit=1,caught_by='unchanged README-pinned Gate.aCheck with current branch compiler',failure=result['failure']))
 finally:p.write_bytes(original)
final=scan('a-check-final',paths);assert final['exit']==0,final['failure']
(out/'header-mutants.json').write_text(json.dumps(mutants,indent=2))
print('PASS 12 programs and both header mutants')
