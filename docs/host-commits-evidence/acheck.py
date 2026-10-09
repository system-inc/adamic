import importlib.util, json, pathlib, subprocess, sys, time, collections
spec=importlib.util.spec_from_file_location('fastgate','/tmp/host-commits/gate.py'); mod=importlib.util.module_from_spec(spec); spec.loader.exec_module(mod)
root=pathlib.Path.cwd(); label=sys.argv[1]
class Process:
 def __init__(self,p,path): self.p=p; self.path=path
 def communicate(self):
  out,err=self.p.communicate(); self.returncode=self.p.returncode
  diagnostics[self.path]={'exit':self.returncode,'stderr':err,'stdout_bytes':len(out.encode())}
  return out,err
class Harness:
 def __init__(self):
  self.arguments=type('Args',(),{'tree':str(root)})(); self.result={};self.steps={};self.exits={};self.failure=None
 def step(self,name,cmd):
  with open('/tmp/host-commits/'+label+'-build.log','w') as log: code=subprocess.run(cmd,cwd=root,stdout=log,stderr=subprocess.STDOUT).returncode
  assert code==0,code
  return True
 def spawn(self,cmd,stdout,stderr): return Process(subprocess.Popen(cmd,cwd=root,stdout=stdout,stderr=stderr,text=True),cmd[-1])
 def fail(self,name,detail): self.failure=detail
paths=json.loads(pathlib.Path(sys.argv[2]).read_text()) if len(sys.argv)>2 else sorted(str(p.relative_to(root)) for p in (root/'stage3').rglob('*.a')); diagnostics={}; h=Harness(); mod.Gate.aCheck(h,paths)
for path,row in h.result['a_check'].items(): row.update(diagnostics[path])
pathlib.Path('/tmp/host-commits/'+label+'.json').write_text(json.dumps(h.result['a_check'],indent=2)+'\n')
print(label,len(paths),dict(collections.Counter(r['outcome'] for r in h.result['a_check'].values())),h.exits,h.steps)
print(h.failure or 'PASS')

sys.exit(h.exits["a-check"])
