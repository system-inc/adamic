import hashlib,json,os,runpy,subprocess
from pathlib import Path
from types import SimpleNamespace
root=Path.cwd();e=root/'review/compiler/loom-r-v4-2-ac-m';Gate=runpy.run_path(str(e/'pinned-gate.py'))['Gate'];captured={}
class Process:
 def __init__(self,process,path):self.process=process;self.path=path
 def communicate(self):
  out,err=self.process.communicate(timeout=60);self.returncode=self.process.returncode;captured[self.path]=dict(exit=self.returncode,stderr=err,stdout_bytes=len(out.encode()));return out,err
class Harness:
 def __init__(self,replay=None):self.arguments=SimpleNamespace(tree=str(root));self.result={};self.steps={};self.exits={};self.failure=None;self.replay=replay
 def step(self,name,command):
  assert name=='a-check-build'
  binary=root.parent/'adamic-binaries/adamic-acheck';binary.parent.mkdir(exist_ok=True)
  with (e/'a-check-build.log').open('w') as log:r=subprocess.run(command,cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=180)
  return r.returncode==0
 def spawn(self,command,stdout,stderr):
  path=command[-1]
  if self.replay is not None:
   data=self.replay[path]
   class Replay:
    returncode=data['exit']
    def communicate(self):return '',data['stderr']
   return Replay()
  return Process(subprocess.Popen(command,cwd=root,stdout=stdout,stderr=stderr,text=True),path)
 def fail(self,name,detail):self.failure=detail
paths=subprocess.check_output(['git','diff','--cached','--name-only','--diff-filter=AMR','HEAD','--','*.a'],text=True).splitlines();(e/'a-paths.txt').write_text('\n'.join(paths)+'\n')
h=Harness();Gate.aCheck(h,paths);(e/'a-check-gate.json').write_text(json.dumps(dict(results=h.result,steps=h.steps,exits=h.exits,failure=h.failure,diagnostics=captured),indent=2)+'\n')
rows=[]
for path in paths:
 data=captured[path];body=(root/path).read_bytes();err=data['stderr'];status='checked' if data['exit']==0 else 'NotYet' if "can't lower" in err and ' yet' in err else 'refused' if 'Adamic 0.1 refuses' in err else 'type error' if ' error TS' in err else 'failed'
 rows.append(dict(path=path,status=status,first_line=body.splitlines()[0].decode(),sha256=hashlib.sha256(body).hexdigest(),diagnostic=err,exit=data['exit']))
(e/'a-statuses.json').write_text(json.dumps(rows,indent=2)+'\n')
assert h.exits['a-check']==0,h.failure
selected=next(r['path'] for r in rows if r['status']=='refused');p=root/selected;old=p.read_bytes();header,body=old.split(b'\n',1);mutants=[]
try:
 for name,value in [('removed-header',body),('wrong-reason',b'// a-check: refused deliberately wrong recut reason\n'+body)]:
  p.write_bytes(value);m=Harness(captured);m.step=lambda name,command: True;Gate.aCheck(m,[selected]);assert m.exits['a-check']==1;mutants.append(dict(mutant=name,path=selected,exit=m.exits['a-check'],failure=m.failure))
finally:p.write_bytes(old)
(e/'a-header-mutants.json').write_text(json.dumps(mutants,indent=2)+'\n');print('a-check PASS',len(rows),'files; both header mutants caught',flush=True)
