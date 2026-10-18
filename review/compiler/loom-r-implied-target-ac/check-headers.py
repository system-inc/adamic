from pathlib import Path
import subprocess, json, hashlib, time, importlib.util, collections
root=Path.cwd(); out=root/'review/compiler/loom-r-implied-target-ac'
cli='/tmp/loom-r-implied-target-ac-adamic'
paths=subprocess.check_output(['git','diff','--name-only','0f15b9f^','0f15b9f','--','*.a'],text=True).splitlines()
rows={}
def compile_file(path,label):
 start=time.monotonic()
 with (out/(label+'.stdout')).open('w') as stdout, (out/(label+'.stderr')).open('w') as stderr:
  p=subprocess.run([cli,'c',path],stdout=stdout,stderr=stderr,timeout=55)
 err=(out/(label+'.stderr')).read_text()
 return dict(exit=p.returncode,stderr=err,seconds=round(time.monotonic()-start,3),command=[cli,'c',path])
for i,path in enumerate(paths):
 original=subprocess.check_output(['git','show','0f15b9f:'+path]); assert Path(path).read_bytes()==original
 r=compile_file(path,'before-'+str(i)); err=r['stderr']
 if r['exit']==0: status='checked';reason=''
 elif "can't lower" in err and ' yet' in err: status='NotYet';reason=err.strip()
 elif 'Adamic 0.1 refuses ' in err: status='refused';reason=err.split('Adamic 0.1 refuses ',1)[1].splitlines()[0]
 elif ' error TS' in err:
  import re
  status='type error';reason=re.search(r' error (TS\d+)',err)[1]
 else: raise RuntimeError(r)
 header='// a-check: '+status+' '+reason+'\n' if status in ('refused','type error') else ''
 if header: Path(path).write_bytes(header.encode()+original)
 assert Path(path).read_bytes()==header.encode()+original
 rows[path]=dict(status=status,reason=reason,header=header.rstrip(),body_sha256=hashlib.sha256(original).hexdigest(),before=r)
 (out/'statuses.json').write_text(json.dumps(rows,indent=2)+'\n')
 print(path,status,r['seconds'],flush=True)
spec=importlib.util.spec_from_file_location('pinned_gate',out/'pinned-gate.py'); gate=importlib.util.module_from_spec(spec);spec.loader.exec_module(gate)
class Process:
 def __init__(self,p,path,diagnostics): self.p=p;self.path=path;self.diagnostics=diagnostics;self.start=time.monotonic()
 def communicate(self):
  stdout,stderr=self.p.communicate(timeout=55);self.returncode=self.p.returncode
  self.diagnostics[self.path]=dict(exit=self.returncode,stderr=stderr,stdout_bytes=len(stdout.encode()),seconds=round(time.monotonic()-self.start,3))
  return stdout,stderr
class Harness:
 def __init__(self,label):
  self.arguments=type('Args',(),{'tree':str(root)})();self.result={};self.steps={};self.exits={};self.failure=None;self.diagnostics={};self.label=label
 def step(self,name,cmd):
  with (out/(self.label+'-build.log')).open('w') as log:
   p=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,timeout=90)
  assert p.returncode==0
  return True
 def spawn(self,cmd,stdout,stderr): return Process(subprocess.Popen(cmd,stdout=stdout,stderr=stderr,text=True),cmd[-1],self.diagnostics)
 def fail(self,name,detail): self.failure=detail
 def scan(self,paths):
  gate.Gate.aCheck(self,paths)
  result=dict(result=self.result,steps=self.steps,exits=self.exits,failure=self.failure,diagnostics=self.diagnostics)
  (out/(self.label+'.json')).write_text(json.dumps(result,indent=2)+'\n');return result
h=Harness('a-check');final=h.scan(paths);assert final['exits']['a-check']==0,final
for path,r in rows.items():r['after']=final['diagnostics'][path]
(out/'statuses.json').write_text(json.dumps(rows,indent=2)+'\n')
path=next(p for p,r in rows.items() if r['status']=='refused');accepted=Path(path).read_bytes();body=accepted.split(b'\n',1)[1]
mutants=[]
try:
 for name,content in [('removed-header',body),('wrong-reason',b'// a-check: refused deliberate-wrong-reason-header-mutant\n'+body)]:
  Path(path).write_bytes(content)
  r=Harness(name).scan([path]);assert r['exits']['a-check']==1 and r['failure'],r
  mutants.append(dict(name=name,path=path,caught=True,failure=r['failure']))
  Path(path).write_bytes(accepted)
finally:Path(path).write_bytes(accepted)
(out/'header-mutants.json').write_text(json.dumps(mutants,indent=2)+'\n')
for path,r in rows.items():
 body=Path(path).read_bytes();body=body.split(b'\n',1)[1] if r['header'] else body
 assert hashlib.sha256(body).hexdigest()==r['body_sha256']
print('PASS',dict(collections.Counter(r['status'] for r in rows.values())),'two mutants caught',flush=True)
