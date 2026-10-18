import pathlib,subprocess,json,importlib.util,hashlib,collections
root=pathlib.Path.cwd();folder=root/'review/compiler/loom-r-hidden-boundaries-ac-m'
script=root/'review/compiler/loom-r-hidden-boundaries-ac/gate.py';spec=importlib.util.spec_from_file_location('gate',script);gate=importlib.util.module_from_spec(spec);spec.loader.exec_module(gate)
paths=subprocess.check_output(['git','diff','--name-only','8c931b22^','8c931b22','--','*.a'],text=True).splitlines();observations={}
class Process:
 def __init__(self,command):self.command=command;self.p=subprocess.Popen(['/workspace/scratch/hidden-ac-m-compiler']+command[1:],stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
 def communicate(self):
  out,err=self.p.communicate(timeout=45);self.returncode=self.p.returncode;observations[self.command[-1]]=dict(exit=self.returncode,stderr=err);return out,err
class Harness:
 def __init__(self):self.arguments=type('Args',(),{'tree':str(root)})();self.result={};self.exits={};self.steps={};self.failure=None
 def step(self,*args):return True # Own branch compiler built separately, unchanged predicate.
 def spawn(self,command,*args,**kw):return Process(command)
 def fail(self,name,detail):self.failure=detail
h=Harness();gate.Gate.aCheck(h,paths);(folder/'a-check.json').write_text(json.dumps(dict(result=h.result,exit=h.exits,failure=h.failure,observations=observations),indent=2)+'\n');assert h.exits['a-check']==0,h.failure
statuses=[]
for p in paths:
 row=observations[p];diag=row['stderr'];status=h.result['a_check'][p]['outcome']
 if "can't lower" in diag and ' yet' in diag:status='NotYet'
 old=subprocess.check_output(['git','show','8c931b22:'+p]);new=(root/p).read_bytes();assert new==old,p
 statuses.append(dict(path=p,status=status,diagnostic=diag,sha256=hashlib.sha256(new).hexdigest()))
(folder/'a-check-statuses.json').write_text(json.dumps(statuses,indent=2)+'\n');print('PASS 60 unchanged programs',collections.Counter(r['status'] for r in statuses))
p=root/'internal/lower/testdata/method_destructuring_receiver.a';original=p.read_bytes();body=original.split(b'\n',1)[1];mutants=[]
try:
 for name,data in [('removed-header',body),('wrong-reason',b'// a-check: refused deliberately unrelated reason\n'+body)]:
  p.write_bytes(data);h=Harness();gate.Gate.aCheck(h,[str(p.relative_to(root))]);assert h.exits['a-check']==1
  mutants.append(dict(name=name,exit=1,failure=h.failure));print('CAUGHT',name)
finally:p.write_bytes(original)
(folder/'a-check-mutants.json').write_text(json.dumps(mutants,indent=2)+'\n')
