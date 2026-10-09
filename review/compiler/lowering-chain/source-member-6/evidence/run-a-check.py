import importlib.util,json,pathlib,subprocess,sys
root=pathlib.Path.cwd();spec=importlib.util.spec_from_file_location('namespace_gate','/tmp/namespace-value-gate/gate.py');mod=importlib.util.module_from_spec(spec);spec.loader.exec_module(mod)
tracked=subprocess.check_output(['git','diff','origin/main','--name-only','--','*.a'],text=True).splitlines()
untracked=subprocess.check_output(['git','ls-files','--others','--exclude-standard','--','*.a'],text=True).splitlines()
paths=sorted(set(tracked+untracked));observations={}
class Process:
 def __init__(self,p,path):self.p=p;self.path=path
 def communicate(self):
  out,err=self.p.communicate();self.returncode=self.p.returncode;observations[self.path]={'exit':self.returncode,'stdout_bytes':len(out.encode()),'stderr':err};return out,err
class Harness:
 def __init__(self):self.arguments=type('Args',(),{'tree':str(root)})();self.result={};self.steps={};self.exits={};self.failure=None
 def step(self,name,cmd):
  with open('/tmp/namespace-value-a-check-build.log','w') as log:code=subprocess.run(cmd,cwd=root,stdout=log,stderr=subprocess.STDOUT).returncode
  return code==0
 def spawn(self,cmd,stdout,stderr):return Process(subprocess.Popen(cmd,cwd=root,stdout=stdout,stderr=stderr,text=True),cmd[-1])
 def fail(self,name,detail):self.failure=detail
h=Harness();mod.Gate.aCheck(h,paths)
pathlib.Path('/tmp/namespace-value-a-check.json').write_text(json.dumps({'gate_pin':'fbac28c62493f27a788edc02a18bc8edb68de5da','paths':paths,'result':h.result,'observations':observations,'steps':h.steps,'exits':h.exits,'failure':h.failure},indent=2)+'\n')
print('changed .a only:',len(paths),paths);print(h.failure or 'PASS');sys.exit(h.exits['a-check'])
